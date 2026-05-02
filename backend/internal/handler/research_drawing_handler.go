package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	researchDrawingUnitPrice             = 2.99
	researchDrawingDefaultMainModelName  = "openrouter/google/gemini-3-flash-preview"
	researchDrawingGPT55ModelName        = "openrouter/openai/gpt-5.5"
	researchDrawingDefaultImageModelName = "openrouter/google/gemini-3.1-flash-image-preview"
	researchDrawingGPTImage2ModelName    = "gpt-image-2"
	researchDrawingGPT55Image2AliasName  = "gpt-5.5-image2"
	researchDrawingLegacyGPTImage2Name   = "openrouter/openai/gpt-5.4-image-2"
	researchDrawingDefaultGPTImageBaseURL = "https://api.openai.com/v1"
)

type ResearchDrawingHandler struct {
	userService    *service.UserService
	settingService *service.SettingService
	httpClient     *http.Client
	mu             sync.Mutex
	// TODO(research-drawing): this in-memory charge map is only a short-term
	// compatibility guard. It is lost on process restart and must be replaced
	// with the research_drawing_jobs table for durable status/refund idempotency.
	jobs map[string]researchDrawingJobCharge
}

type researchDrawingJobCharge struct {
	UserID          int64
	Charge          float64
	Refunded        bool
	PaperBananaUser string
}

type ResearchDrawingGenerateRequest struct {
	MethodContent         string `json:"method_content" binding:"required"`
	Caption               string `json:"caption"`
	OptimizeMethodContent bool   `json:"optimize_method_content"`
	GenerationMode        string `json:"generation_mode"`
	ExpMode               string `json:"exp_mode"`
	RetrievalSetting      string `json:"retrieval_setting"`
	NumCandidates         int    `json:"num_candidates"`
	AspectRatio           string `json:"aspect_ratio"`
	MaxCriticRounds       int    `json:"max_critic_rounds"`
	MaxRefineResolution   string `json:"max_refine_resolution"`
	MainModelName         string `json:"main_model_name"`
	ImageGenModelName     string `json:"image_gen_model_name"`
}

type researchDrawingGenerateResponse struct {
	JobID           string  `json:"job_id"`
	Status          string  `json:"status"`
	PaperBananaURL  string  `json:"paperbanana_url,omitempty"`
	PaperBananaUser string  `json:"paperbanana_user,omitempty"`
	Charge          float64 `json:"charge"`
	QuotaNeed       int     `json:"quota_need"`
}

func NewResearchDrawingHandler(userService *service.UserService, settingService *service.SettingService) *ResearchDrawingHandler {
	return &ResearchDrawingHandler{
		userService:    userService,
		settingService: settingService,
		httpClient:     &http.Client{Timeout: 20 * time.Second},
		jobs:           make(map[string]researchDrawingJobCharge),
	}
}

func (h *ResearchDrawingHandler) Generate(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req ResearchDrawingGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	req.normalize()
	if strings.TrimSpace(req.MethodContent) == "" {
		response.BadRequest(c, "method_content is required")
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	if req.OptimizeMethodContent && !h.researchDrawingMethodOptimizationEnabled(c.Request.Context()) {
		response.BadRequest(c, "\u65b9\u6cd5\u5185\u5bb9\u4f18\u5316\u6682\u672a\u542f\u7528")
		return
	}

	charge := h.researchDrawingUnitPrice(c.Request.Context())
	if user.Balance < charge {
		response.ErrorFrom(c, infraerrors.New(http.StatusPaymentRequired, "INSUFFICIENT_BALANCE", "insufficient balance"))
		return
	}

	if err := h.userService.UpdateBalance(c.Request.Context(), subject.UserID, -charge); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	pbResp, err := h.submitToPaperBanana(c, user, req)
	if err != nil {
		_ = h.userService.UpdateBalance(c.Request.Context(), subject.UserID, charge)
		response.ErrorFrom(c, infraerrors.New(http.StatusBadGateway, "PAPERBANANA_SUBMIT_FAILED", err.Error()))
		return
	}

	jobID := strings.TrimSpace(stringFromMap(pbResp, "job_id"))
	pbUser := strings.TrimSpace(stringFromMap(pbResp, "username"))
	if jobID == "" {
		_ = h.userService.UpdateBalance(c.Request.Context(), subject.UserID, charge)
		response.ErrorFrom(c, infraerrors.New(http.StatusBadGateway, "PAPERBANANA_INVALID_RESPONSE", "PaperBanana did not return a job_id"))
		return
	}
	if pbUser == "" {
		pbUser = paperBananaUsername(user)
	}

	h.mu.Lock()
	h.jobs[jobID] = researchDrawingJobCharge{UserID: subject.UserID, Charge: charge, PaperBananaUser: pbUser}
	h.mu.Unlock()

	response.Accepted(c, researchDrawingGenerateResponse{
		JobID:           jobID,
		Status:          "running",
		PaperBananaURL:  paperBananaBaseURL(),
		PaperBananaUser: pbUser,
		Charge:          charge,
		QuotaNeed:       req.quotaNeed(),
	})
}

func (h *ResearchDrawingHandler) JobStatus(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	jobID := strings.TrimSpace(c.Param("job_id"))
	pbUser := strings.TrimSpace(c.Query("paperbanana_user"))
	if jobID == "" {
		response.BadRequest(c, "job_id is required")
		return
	}

	status, err := h.getPaperBananaStatus(c, subject.UserID, jobID, pbUser)
	if err != nil {
		response.ErrorFrom(c, infraerrors.New(http.StatusBadGateway, "PAPERBANANA_STATUS_FAILED", err.Error()))
		return
	}

	if strings.EqualFold(stringFromMap(status, "status"), "error") {
		h.refundJobOnce(c, subject.UserID, jobID)
	}
	response.Success(c, status)
}

func (h *ResearchDrawingHandler) JobImage(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	jobID := strings.TrimSpace(c.Param("job_id"))
	candidateID := strings.TrimSpace(c.Param("candidate_id"))
	pbUser := strings.TrimSpace(c.Query("paperbanana_user"))
	if jobID == "" || candidateID == "" {
		response.BadRequest(c, "job_id and candidate_id are required")
		return
	}

	body, contentType, err := h.getPaperBananaImage(c, subject.UserID, jobID, candidateID, pbUser)
	if err != nil {
		response.ErrorFrom(c, infraerrors.New(http.StatusBadGateway, "PAPERBANANA_IMAGE_FAILED", err.Error()))
		return
	}
	if contentType == "" {
		contentType = "image/png"
	}
	c.Data(http.StatusOK, contentType, body)
}

func (h *ResearchDrawingHandler) refundJobOnce(c *gin.Context, userID int64, jobID string) {
	h.mu.Lock()
	charge, ok := h.jobs[jobID]
	if !ok || charge.UserID != userID || charge.Refunded {
		h.mu.Unlock()
		return
	}
	charge.Refunded = true
	h.jobs[jobID] = charge
	h.mu.Unlock()
	_ = h.userService.UpdateBalance(c.Request.Context(), userID, charge.Charge)
}

func (h *ResearchDrawingHandler) submitToPaperBanana(c *gin.Context, user *service.User, req ResearchDrawingGenerateRequest) (map[string]any, error) {
	payload := map[string]any{
		"user_id":                 user.ID,
		"username":                paperBananaUsername(user),
		"email":                   user.Email,
		"method_content":          req.MethodContent,
		"caption":                 req.Caption,
		"optimize_method_content": req.OptimizeMethodContent,
		"generation_mode":         req.GenerationMode,
		"exp_mode":                req.ExpMode,
		"retrieval_setting":       req.RetrievalSetting,
		"num_candidates":          req.NumCandidates,
		"aspect_ratio":            req.AspectRatio,
		"max_critic_rounds":       req.MaxCriticRounds,
		"max_refine_resolution":   req.MaxRefineResolution,
		"main_model_name":         req.MainModelName,
		"image_gen_model_name":    req.ImageGenModelName,
	}
	if req.isGPTImage2() {
		apiKey, baseURL := h.researchDrawingGPTImageConfig(c.Request.Context())
		if apiKey != "" {
			payload["gpt_image_api_key"] = apiKey
		}
		if baseURL != "" {
			payload["gpt_image_base_url"] = baseURL
		}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	baseURL := paperBananaBaseURL()
	if baseURL == "" {
		return nil, fmt.Errorf("research drawing service URL is not configured")
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/sub2api/generate"
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if token := paperBananaServiceToken(); token != "" {
		httpReq.Header.Set("x-sub2api-service-token", token)
	}
	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("PaperBanana returned %d: %s", resp.StatusCode, stringFromMap(out, "error"))
	}
	return out, nil
}

func (h *ResearchDrawingHandler) getPaperBananaImage(c *gin.Context, userID int64, jobID, candidateID, pbUser string) ([]byte, string, error) {
	if pbUser == "" {
		h.mu.Lock()
		if charge, ok := h.jobs[jobID]; ok {
			pbUser = charge.PaperBananaUser
		}
		h.mu.Unlock()
	}
	if pbUser == "" {
		user, err := h.userService.GetProfile(c.Request.Context(), userID)
		if err == nil && user != nil {
			pbUser = paperBananaUsername(user)
		}
	}
	baseURL := paperBananaBaseURL()
	if baseURL == "" {
		return nil, "", fmt.Errorf("research drawing service URL is not configured")
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/sub2api/job/" + url.PathEscape(jobID) + "/image/" + url.PathEscape(candidateID) + "?username=" + url.QueryEscape(pbUser)
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", err
	}
	if token := paperBananaServiceToken(); token != "" {
		httpReq.Header.Set("x-sub2api-service-token", token)
	}
	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("PaperBanana returned %d", resp.StatusCode)
	}
	return body, resp.Header.Get("Content-Type"), nil
}

func (h *ResearchDrawingHandler) getPaperBananaStatus(c *gin.Context, userID int64, jobID, pbUser string) (map[string]any, error) {
	if pbUser == "" {
		h.mu.Lock()
		if charge, ok := h.jobs[jobID]; ok {
			pbUser = charge.PaperBananaUser
		}
		h.mu.Unlock()
	}
	if pbUser == "" {
		user, err := h.userService.GetProfile(c.Request.Context(), userID)
		if err == nil && user != nil {
			pbUser = paperBananaUsername(user)
		}
	}
	baseURL := paperBananaBaseURL()
	if baseURL == "" {
		return nil, fmt.Errorf("research drawing service URL is not configured")
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/sub2api/job/" + url.PathEscape(jobID) + "?username=" + url.QueryEscape(pbUser)
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if token := paperBananaServiceToken(); token != "" {
		httpReq.Header.Set("x-sub2api-service-token", token)
	}
	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("PaperBanana returned %d: %s", resp.StatusCode, stringFromMap(out, "error"))
	}
	return out, nil
}

func (r *ResearchDrawingGenerateRequest) normalize() {
	r.MethodContent = strings.TrimSpace(r.MethodContent)
	r.GenerationMode = strings.TrimSpace(r.GenerationMode)
	if r.ExpMode != "demo_full" && r.ExpMode != "demo_planner_critic" {
		r.ExpMode = "demo_planner_critic"
	}
	if r.RetrievalSetting != "manual" && r.RetrievalSetting != "random" && r.RetrievalSetting != "none" {
		r.RetrievalSetting = "auto"
	}
	if r.AspectRatio != "21:9" && r.AspectRatio != "3:2" {
		r.AspectRatio = "16:9"
	}
	if r.NumCandidates < 1 {
		r.NumCandidates = 1
	}
	if r.NumCandidates > 20 {
		r.NumCandidates = 20
	}
	if r.MaxCriticRounds < 1 {
		r.MaxCriticRounds = 2
	}
	if r.MaxCriticRounds > 5 {
		r.MaxCriticRounds = 5
	}
	r.MaxRefineResolution = strings.ToUpper(strings.TrimSpace(r.MaxRefineResolution))
	if r.MaxRefineResolution != "4K" {
		r.MaxRefineResolution = "2K"
	}
	r.MainModelName = strings.TrimSpace(r.MainModelName)
	if len(r.MainModelName) > 200 {
		r.MainModelName = r.MainModelName[:200]
	}
	switch r.MainModelName {
	case researchDrawingDefaultMainModelName, researchDrawingGPT55ModelName:
	default:
		r.MainModelName = researchDrawingDefaultMainModelName
	}
	r.ImageGenModelName = strings.TrimSpace(r.ImageGenModelName)
	if len(r.ImageGenModelName) > 200 {
		r.ImageGenModelName = r.ImageGenModelName[:200]
	}
	switch r.ImageGenModelName {
	case researchDrawingDefaultImageModelName:
	case researchDrawingGPTImage2ModelName, researchDrawingGPT55Image2AliasName, researchDrawingLegacyGPTImage2Name:
		r.ImageGenModelName = researchDrawingGPTImage2ModelName
	default:
		r.ImageGenModelName = researchDrawingDefaultImageModelName
	}
}

func (r ResearchDrawingGenerateRequest) isGPTImage2() bool {
	switch strings.TrimSpace(r.ImageGenModelName) {
	case researchDrawingGPTImage2ModelName, researchDrawingGPT55Image2AliasName, researchDrawingLegacyGPTImage2Name:
		return true
	default:
		return false
	}
}

func (r ResearchDrawingGenerateRequest) quotaNeed() int {
	candidates := r.NumCandidates
	if candidates < 1 {
		candidates = 1
	}
	rounds := r.MaxCriticRounds
	if rounds < 1 {
		rounds = 1
	}
	return candidates * (1 + rounds)
}

func (h *ResearchDrawingHandler) researchDrawingUnitPrice(ctx context.Context) float64 {
	if h.settingService == nil {
		return researchDrawingUnitPrice
	}
	settings, err := h.settingService.GetAllSettings(ctx)
	if err != nil || settings == nil || settings.ResearchDrawingUnitPrice <= 0 {
		return researchDrawingUnitPrice
	}
	return settings.ResearchDrawingUnitPrice
}

func (h *ResearchDrawingHandler) researchDrawingMethodOptimizationEnabled(ctx context.Context) bool {
	if h.settingService == nil {
		return true
	}
	settings, err := h.settingService.GetAllSettings(ctx)
	if err != nil || settings == nil {
		return true
	}
	return settings.ResearchDrawingMethodOptimizationEnabled
}

func (h *ResearchDrawingHandler) researchDrawingGPTImageConfig(ctx context.Context) (string, string) {
	apiKey := strings.TrimSpace(os.Getenv("GPT_IMAGE_API_KEY"))
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("GPT_IMAGE_BASE_URL")), "/")
	if h.settingService != nil {
		if settings, err := h.settingService.GetAllSettings(ctx); err == nil && settings != nil {
			if v := strings.TrimSpace(settings.ResearchDrawingGPTImageAPIKey); v != "" {
				apiKey = v
			}
			if v := strings.TrimRight(strings.TrimSpace(settings.ResearchDrawingGPTImageBaseURL), "/"); v != "" {
				baseURL = v
			}
		}
	}
	if baseURL == "" {
		baseURL = researchDrawingDefaultGPTImageBaseURL
	}
	return apiKey, baseURL
}

func isLocalRunMode() bool {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("SERVER_MODE")))
	runMode := strings.ToLower(strings.TrimSpace(os.Getenv("RUN_MODE")))
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	return mode == "debug" || mode == "dev" || mode == "development" || runMode == "local" || appEnv == "local" || appEnv == "dev" || appEnv == "development"
}

func paperBananaBaseURL() string {
	if v := strings.TrimSpace(os.Getenv("RESEARCH_DRAWING_API_URL")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("PAPERBANANA_BASE_URL")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("PB_INTERNAL_BASE_URL")); v != "" {
		return v
	}
	if isLocalRunMode() {
		return "http://127.0.0.1:8000"
	}
	return ""
}

func paperBananaServiceToken() string {
	if v := strings.TrimSpace(os.Getenv("RESEARCH_DRAWING_SERVICE_TOKEN")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("PAPERBANANA_SERVICE_TOKEN"))
}

func paperBananaUsername(user *service.User) string {
	if user == nil {
		return "sub2api_user"
	}
	if strings.TrimSpace(user.Email) != "" {
		return fmt.Sprintf("s2a_%d", user.ID)
	}
	return fmt.Sprintf("s2a_%d", user.ID)
}

func stringFromMap(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
