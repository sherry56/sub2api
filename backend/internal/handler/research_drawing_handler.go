package handler

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/model"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	researchDrawingUnitPrice             = 2.99
	researchDrawingNanoBanana2UnitPrice  = 2.99
	researchDrawingGPTImage2UnitPrice    = 0.99
	researchDrawingDefaultImageModelName = "openrouter/google/gemini-3.1-flash-image-preview"
	researchDrawingGPTImage2ModelName    = "gpt-image-2"
	researchDrawingGPTImage2DirectSize   = "1024x1024"
	researchDrawingGPTImage2MaxAttempts  = 2
	researchDrawingGPTImage2Timeout      = 300 * time.Second
	researchDrawingUpstreamBusyMessage   = "上游服务暂时繁忙，请稍后再试。生成失败不扣费。"
	researchDrawingUpstreamTimeoutMessage = "上游生成超时，请稍后再试。生成失败不扣费。"
)

type ResearchDrawingHandler struct {
	userService         *service.UserService
	settingService      *service.SettingService
	image2RecordService *service.ResearchDrawingImage2RecordService
	httpClient          *http.Client
	mu                  sync.Mutex
	// TODO(research-drawing): this in-memory charge map is only a short-term
	// compatibility guard. It is lost on process restart and must be replaced
	// with the research_drawing_jobs table for durable status/refund idempotency.
	jobs map[string]researchDrawingJobCharge
}

type researchDrawingJobCharge struct {
	UserID          int64
	Charge          float64
	ResolvedPrice   float64
	Charged         bool
	Charging        bool
	Refunded        bool
	PaperBananaUser string
	ImageGenModelName string
	Direct            bool
	Status            string
	Error             string
	StartedAt         time.Time
	FinishedAt        time.Time
	ImagePrompt       string
	Images            map[int]researchDrawingDirectImage
}

type researchDrawingDirectImage struct {
	ContentType string
	Bytes       []byte
	Path        string
}

type researchDrawingDirectGPTConfig struct {
	ImageAPIKey   string
	ImageBaseURL  string
	KeySource     string
	BaseURLSource string
}

type ResearchDrawingGenerateRequest struct {
	MethodContent       string `json:"method_content" binding:"required"`
	Caption             string `json:"caption"`
	GenerationMode      string `json:"generation_mode"`
	ExpMode             string `json:"exp_mode"`
	RetrievalSetting    string `json:"retrieval_setting"`
	NumCandidates       int    `json:"num_candidates"`
	AspectRatio         string `json:"aspect_ratio"`
	MaxCriticRounds     int    `json:"max_critic_rounds"`
	MaxRefineResolution string `json:"max_refine_resolution"`
	ImageGenModelName   string `json:"image_gen_model_name"`
}

type researchDrawingGenerateResponse struct {
	JobID           string  `json:"job_id"`
	Status          string  `json:"status"`
	Mode            string  `json:"mode,omitempty"`
	PaperBananaURL  string  `json:"paperbanana_url,omitempty"`
	PaperBananaUser string  `json:"paperbanana_user,omitempty"`
	Charge          float64 `json:"charge"`
	QuotaNeed       int     `json:"quota_need"`
}

type researchDrawingImage2RecordsResponse struct {
	Records []model.ResearchDrawingImage2Record `json:"records"`
}

func NewResearchDrawingHandler(userService *service.UserService, settingService *service.SettingService, image2RecordService *service.ResearchDrawingImage2RecordService) *ResearchDrawingHandler {
	return &ResearchDrawingHandler{
		userService:         userService,
		settingService:      settingService,
		image2RecordService: image2RecordService,
		httpClient:          &http.Client{Timeout: 180 * time.Second},
		jobs:                make(map[string]researchDrawingJobCharge),
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
	directMode := req.isDirectGPTMode()
	if directMode {
		req.forceDirectGPTMode()
	}

	user, err := h.userService.GetProfile(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var directCfg researchDrawingDirectGPTConfig
	if directMode {
		var cfgErr error
		directCfg, cfgErr = h.researchDrawingDirectGPTConfig(c.Request.Context(), req)
		if cfgErr != nil {
			response.ErrorFrom(c, infraerrors.New(http.StatusBadGateway, "RESEARCH_DRAWING_GPT_CONFIG_INVALID", cfgErr.Error()))
			return
		}
	}

	resolvedPrice := resolveResearchDrawingPrice(req.ImageGenModelName)
	if user.Balance < resolvedPrice {
		response.ErrorFrom(c, infraerrors.New(http.StatusPaymentRequired, "INSUFFICIENT_BALANCE", "insufficient balance"))
		return
	}

	if directMode {
		jobID := newResearchDrawingJobID()
		h.mu.Lock()
		h.jobs[jobID] = researchDrawingJobCharge{
			UserID:            subject.UserID,
			Charge:            resolvedPrice,
			ResolvedPrice:     resolvedPrice,
			PaperBananaUser:   paperBananaUsername(user),
			ImageGenModelName: req.ImageGenModelName,
			Direct:            true,
			Status:            "running",
			StartedAt:         time.Now(),
			Images:            make(map[int]researchDrawingDirectImage),
		}
		h.mu.Unlock()
		if charged, err := h.chargeJobOnceWithContext(c.Request.Context(), subject.UserID, jobID, req.ImageGenModelName, resolvedPrice); err != nil {
			response.ErrorFrom(c, err)
			return
		} else if !charged {
			response.ErrorFrom(c, infraerrors.New(http.StatusConflict, "RESEARCH_DRAWING_JOB_ALREADY_CHARGED", "research drawing job already charged"))
			return
		}

		go h.runDirectGPTResearchDrawingJob(jobID, req, directCfg)

		response.Accepted(c, researchDrawingGenerateResponse{
			JobID:     jobID,
			Status:    "running",
			Mode:      "direct_gpt",
			Charge:    resolvedPrice,
			QuotaNeed: req.directQuotaNeed(),
		})
		return
	}

	if err := h.userService.UpdateBalance(c.Request.Context(), subject.UserID, -resolvedPrice); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	pbResp, err := h.submitToPaperBanana(c, user, req)
	if err != nil {
		h.refundUntrackedCharge(c.Request.Context(), subject.UserID, "", req.ImageGenModelName, resolvedPrice, "paperbanana_submit_failed")
		response.ErrorFrom(c, infraerrors.New(http.StatusBadGateway, "PAPERBANANA_SUBMIT_FAILED", h.researchDrawingUserFacingError("paperbanana_submit_failed", "", req.ImageGenModelName, err)))
		return
	}

	jobID := strings.TrimSpace(stringFromMap(pbResp, "job_id"))
	pbUser := strings.TrimSpace(stringFromMap(pbResp, "username"))
	if jobID == "" {
		h.refundUntrackedCharge(c.Request.Context(), subject.UserID, "", req.ImageGenModelName, resolvedPrice, "paperbanana_invalid_response")
		response.ErrorFrom(c, infraerrors.New(http.StatusBadGateway, "PAPERBANANA_INVALID_RESPONSE", "PaperBanana did not return a job_id"))
		return
	}
	if pbUser == "" {
		pbUser = paperBananaUsername(user)
	}

	if duplicate, existingCharge := h.recordPrechargedJob(jobID, subject.UserID, req.ImageGenModelName, resolvedPrice, pbUser, false); duplicate {
		h.refundUntrackedCharge(c.Request.Context(), subject.UserID, jobID, req.ImageGenModelName, resolvedPrice, "duplicate_charge_refunded")
		resolvedPrice = existingCharge
		h.logBillingEvent("charge_skipped", jobID, subject.UserID, req.ImageGenModelName, resolvedPrice, 0, 0, "running", nil)
	} else {
		h.logBillingEvent("charge", jobID, subject.UserID, req.ImageGenModelName, resolvedPrice, resolvedPrice, 0, "running", nil)
	}

	response.Accepted(c, researchDrawingGenerateResponse{
		JobID:           jobID,
		Status:          "running",
		PaperBananaURL:  paperBananaBaseURL(),
		PaperBananaUser: pbUser,
		Charge:          resolvedPrice,
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

	if status, handled, err := h.getDirectJobStatus(subject.UserID, jobID); handled {
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		if strings.EqualFold(stringFromMap(status, "status"), "error") {
			h.refundJobOnce(c, subject.UserID, jobID)
		}
		applyResearchDrawingFriendlyStatusError(status)
		if researchDrawingStatusDone(status) || researchDrawingStatusFailed(status) {
			h.logJobFinalStatus(subject.UserID, jobID, stringFromMap(status, "status"))
		}
		response.Success(c, status)
		return
	}

	status, err := h.getPaperBananaStatus(c, subject.UserID, jobID, pbUser)
	if err != nil {
		h.refundJobOnce(c, subject.UserID, jobID)
		response.ErrorFrom(c, infraerrors.New(http.StatusBadGateway, "PAPERBANANA_STATUS_FAILED", h.researchDrawingUserFacingError("paperbanana_status_failed", jobID, "", err)))
		return
	}

	if researchDrawingStatusFailed(status) {
		h.refundJobOnce(c, subject.UserID, jobID)
	}
	if researchDrawingStatusDone(status) && !researchDrawingStatusHasCandidates(status) {
		h.refundJobOnce(c, subject.UserID, jobID)
		status["status"] = "error"
		status["error"] = "research drawing finished without valid images"
	}
	applyResearchDrawingFriendlyStatusError(status)
	if researchDrawingStatusDone(status) || researchDrawingStatusFailed(status) {
		h.logJobFinalStatus(subject.UserID, jobID, stringFromMap(status, "status"))
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

	if body, contentType, handled, err := h.getDirectJobImage(c.Request.Context(), subject.UserID, jobID, candidateID); handled {
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		if contentType == "" {
			contentType = "image/png"
		}
		c.Data(http.StatusOK, contentType, body)
		return
	}

	body, contentType, err := h.getPaperBananaImage(c, subject.UserID, jobID, candidateID, pbUser)
	if err != nil {
		h.refundJobOnce(c, subject.UserID, jobID)
		response.ErrorFrom(c, infraerrors.New(http.StatusBadGateway, "PAPERBANANA_IMAGE_FAILED", h.researchDrawingUserFacingError("paperbanana_image_failed", jobID, "", err)))
		return
	}
	if !isValidResearchDrawingImage(body, contentType) {
		h.refundJobOnce(c, subject.UserID, jobID)
		response.ErrorFrom(c, infraerrors.New(http.StatusBadGateway, "PAPERBANANA_IMAGE_INVALID", "PaperBanana returned no valid image"))
		return
	}
	if contentType == "" {
		contentType = "image/png"
	}
	c.Data(http.StatusOK, contentType, body)
}

func (h *ResearchDrawingHandler) Image2Records(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	limit := 20
	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 100 {
		limit = 100
	}

	records, err := h.image2RecordService.ListByUser(c.Request.Context(), subject.UserID, limit)
	if err != nil {
		response.ErrorFrom(c, infraerrors.New(http.StatusInternalServerError, "RESEARCH_DRAWING_IMAGE2_RECORDS_FAILED", err.Error()))
		return
	}
	response.Success(c, researchDrawingImage2RecordsResponse{Records: records})
}

func (h *ResearchDrawingHandler) chargeJobOnceWithContext(ctx context.Context, userID int64, jobID, imageModelName string, resolvedPrice float64) (bool, error) {
	if h.userService == nil {
		return false, infraerrors.New(http.StatusInternalServerError, "RESEARCH_DRAWING_BILLING_UNAVAILABLE", "research drawing billing is unavailable")
	}
	imageModelName = normalizeResearchDrawingImageModelNameForLog(imageModelName)
	h.mu.Lock()
	job, ok := h.jobs[jobID]
	if !ok {
		job = researchDrawingJobCharge{
			UserID:            userID,
			Status:            "running",
			ImageGenModelName: imageModelName,
			StartedAt:         time.Now(),
		}
	}
	if job.UserID != userID {
		h.mu.Unlock()
		return false, infraerrors.New(http.StatusConflict, "RESEARCH_DRAWING_JOB_OWNER_MISMATCH", "research drawing job owner mismatch")
	}
	if job.Charged || job.Charging {
		h.mu.Unlock()
		h.logBillingEvent("charge_skipped", jobID, userID, imageModelName, resolvedPrice, 0, 0, job.Status, nil)
		return false, nil
	}
	job.Charging = true
	job.Charge = resolvedPrice
	job.ResolvedPrice = resolvedPrice
	job.ImageGenModelName = imageModelName
	h.jobs[jobID] = job
	h.mu.Unlock()

	if err := h.userService.UpdateBalance(ctx, userID, -resolvedPrice); err != nil {
		h.mu.Lock()
		if latest, ok := h.jobs[jobID]; ok && latest.UserID == userID {
			latest.Charging = false
			h.jobs[jobID] = latest
		}
		h.mu.Unlock()
		h.logBillingEvent("charge_failed", jobID, userID, imageModelName, resolvedPrice, resolvedPrice, 0, "charge_failed", err)
		return false, err
	}

	h.mu.Lock()
	job = h.jobs[jobID]
	job.Charging = false
	job.Charged = true
	job.Charge = resolvedPrice
	job.ResolvedPrice = resolvedPrice
	job.ImageGenModelName = imageModelName
	h.jobs[jobID] = job
	finalStatus := job.Status
	h.mu.Unlock()
	h.logBillingEvent("charge", jobID, userID, imageModelName, resolvedPrice, resolvedPrice, 0, finalStatus, nil)
	return true, nil
}

func (h *ResearchDrawingHandler) recordPrechargedJob(jobID string, userID int64, imageModelName string, resolvedPrice float64, pbUser string, direct bool) (bool, float64) {
	imageModelName = normalizeResearchDrawingImageModelNameForLog(imageModelName)
	h.mu.Lock()
	defer h.mu.Unlock()
	if existing, ok := h.jobs[jobID]; ok && existing.Charged {
		existingPrice := existing.ResolvedPrice
		if existingPrice <= 0 {
			existingPrice = existing.Charge
		}
		return true, existingPrice
	}
	h.jobs[jobID] = researchDrawingJobCharge{
		UserID:            userID,
		Charge:            resolvedPrice,
		ResolvedPrice:     resolvedPrice,
		Charged:           true,
		PaperBananaUser:   pbUser,
		ImageGenModelName: imageModelName,
		Direct:            direct,
		Status:            "running",
		StartedAt:         time.Now(),
	}
	return false, resolvedPrice
}

func (h *ResearchDrawingHandler) refundUntrackedCharge(ctx context.Context, userID int64, jobID, imageModelName string, resolvedPrice float64, finalStatus string) {
	if h.userService == nil {
		return
	}
	err := h.userService.UpdateBalance(ctx, userID, resolvedPrice)
	event := "refund"
	if err != nil {
		event = "refund_failed"
	}
	h.logBillingEvent(event, jobID, userID, imageModelName, resolvedPrice, 0, resolvedPrice, finalStatus, err)
}

func (h *ResearchDrawingHandler) refundJobOnce(c *gin.Context, userID int64, jobID string) {
	if c == nil || c.Request == nil {
		h.refundJobOnceWithContext(context.Background(), userID, jobID)
		return
	}
	h.refundJobOnceWithContext(c.Request.Context(), userID, jobID)
}

func (h *ResearchDrawingHandler) refundJobOnceWithContext(ctx context.Context, userID int64, jobID string) {
	if h.userService == nil {
		return
	}
	h.mu.Lock()
	charge, ok := h.jobs[jobID]
	if !ok || charge.UserID != userID || charge.Refunded || !charge.Charged {
		h.mu.Unlock()
		return
	}
	charge.Refunded = true
	h.jobs[jobID] = charge
	h.mu.Unlock()
	refundAmount := charge.ResolvedPrice
	if refundAmount <= 0 {
		refundAmount = charge.Charge
	}
	if err := h.userService.UpdateBalance(ctx, userID, refundAmount); err != nil {
		h.mu.Lock()
		if latest, ok := h.jobs[jobID]; ok && latest.UserID == userID {
			latest.Refunded = false
			h.jobs[jobID] = latest
		}
		h.mu.Unlock()
		h.logBillingEvent("refund_failed", jobID, userID, charge.ImageGenModelName, refundAmount, 0, refundAmount, charge.Status, err)
		return
	}
	h.logBillingEvent("refund", jobID, userID, charge.ImageGenModelName, refundAmount, 0, refundAmount, charge.Status, nil)
}

func (h *ResearchDrawingHandler) submitToPaperBanana(c *gin.Context, user *service.User, req ResearchDrawingGenerateRequest) (map[string]any, error) {
	payload := map[string]any{
		"user_id":               user.ID,
		"username":              paperBananaUsername(user),
		"email":                 user.Email,
		"method_content":        req.MethodContent,
		"caption":               req.Caption,
		"generation_mode":       req.GenerationMode,
		"exp_mode":              req.ExpMode,
		"retrieval_setting":     req.RetrievalSetting,
		"num_candidates":        req.NumCandidates,
		"aspect_ratio":          req.AspectRatio,
		"max_critic_rounds":     req.MaxCriticRounds,
		"max_refine_resolution": req.MaxRefineResolution,
		"image_gen_model_name":  req.ImageGenModelName,
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

func (h *ResearchDrawingHandler) runDirectGPTResearchDrawingJob(jobID string, req ResearchDrawingGenerateRequest, cfg researchDrawingDirectGPTConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(researchDrawingGPTImage2MaxAttempts)*researchDrawingGPTImage2Timeout+30*time.Second)
	defer cancel()

	imagePrompt := buildResearchDrawingDirectImagePrompt(req)
	imageBytes, err := h.generateResearchDrawingDirectImage(ctx, jobID, req, imagePrompt, cfg)
	if err != nil {
		h.failDirectJob(jobID, err)
		return
	}

	outPath, err := saveResearchDrawingDirectCandidate(jobID, imageBytes)
	if err != nil {
		h.failDirectJob(jobID, fmt.Errorf("save candidate_0.png: %w", err))
		return
	}

	userID := int64(0)
	resolvedPrice := float64(0)
	h.mu.Lock()
	job, ok := h.jobs[jobID]
	if ok && job.Direct {
		userID = job.UserID
		resolvedPrice = job.ResolvedPrice
		if resolvedPrice <= 0 {
			resolvedPrice = job.Charge
		}
		if job.Images == nil {
			job.Images = make(map[int]researchDrawingDirectImage)
		}
		job.Status = "done"
		job.Error = ""
		job.FinishedAt = time.Now()
		job.ImagePrompt = imagePrompt
		job.Images[0] = researchDrawingDirectImage{
			ContentType: "image/png",
			Bytes:       imageBytes,
			Path:        outPath,
		}
		h.jobs[jobID] = job
	}
	h.mu.Unlock()
	if userID > 0 {
		h.recordImage2Job(context.Background(), userID, jobID, researchDrawingGPTImage2ModelName)
		h.logBillingEvent("final_status", jobID, userID, req.ImageGenModelName, resolvedPrice, 0, 0, "done", nil)
	}
	log.Printf("[ResearchDrawing] direct GPT candidate image saved job_id=%s candidate_id=0 bytes=%d path=%s", jobID, len(imageBytes), outPath)
}

func (h *ResearchDrawingHandler) recordImage2Job(ctx context.Context, userID int64, jobID, modelName string) {
	if h.image2RecordService == nil || userID <= 0 || strings.TrimSpace(jobID) == "" {
		return
	}
	record := model.ResearchDrawingImage2Record{
		UserID:    userID,
		JobID:     strings.TrimSpace(jobID),
		Model:     strings.TrimSpace(modelName),
		CreatedAt: time.Now().UTC(),
	}
	if record.Model == "" {
		record.Model = researchDrawingGPTImage2ModelName
	}
	if err := h.image2RecordService.Create(ctx, record); err != nil {
		log.Printf("[ResearchDrawing] image2 record save failed job_id=%s user_id=%d error=%s", jobID, userID, err.Error())
		return
	}
	log.Printf("[ResearchDrawing] image2 record saved job_id=%s user_id=%d model=%s", jobID, userID, record.Model)
}

func (h *ResearchDrawingHandler) generateResearchDrawingDirectImage(ctx context.Context, jobID string, req ResearchDrawingGenerateRequest, imagePrompt string, cfg researchDrawingDirectGPTConfig) ([]byte, error) {
	endpoint := strings.TrimRight(cfg.ImageBaseURL, "/") + "/images/generations"
	size := researchDrawingDirectImageSize(req.AspectRatio)
	payload := map[string]any{
		"model":  researchDrawingGPTImage2ModelName,
		"prompt": imagePrompt,
		"size":   size,
	}

	var lastErr error
	for attempt := 0; attempt < researchDrawingGPTImage2MaxAttempts; attempt++ {
		start := time.Now()
		body, contentType, statusCode, err := h.postResearchDrawingGPTJSON(ctx, endpoint, cfg.ImageAPIKey, payload)
		elapsed := time.Since(start)
		retryCount := attempt
		if err != nil {
			lastErr = err
			h.logDirectGPTImageRequest(jobID, endpoint, imagePrompt, size, retryCount, statusCode, elapsed, false, cfg, contentType, nil, err)
			if attempt+1 < researchDrawingGPTImage2MaxAttempts && isResearchDrawingRetryableDirectFailure(statusCode, err) {
				continue
			}
			return nil, err
		}
		if statusCode < 200 || statusCode >= 300 {
			lastErr = fmt.Errorf("gpt-image-2 image request failed: status_code=%d content_type=%s response_preview=%s", statusCode, contentType, researchDrawingResponsePreview(body))
			h.logDirectGPTImageRequest(jobID, endpoint, imagePrompt, size, retryCount, statusCode, elapsed, false, cfg, contentType, body, lastErr)
			if attempt+1 < researchDrawingGPTImage2MaxAttempts && isResearchDrawingRetryableDirectFailure(statusCode, nil) {
				continue
			}
			return nil, lastErr
		}
		if strings.Contains(strings.ToLower(contentType), "text/html") {
			lastErr = fmt.Errorf("gpt-image-2 image request returned html: status_code=%d response_preview=%s", statusCode, researchDrawingResponsePreview(body))
			h.logDirectGPTImageRequest(jobID, endpoint, imagePrompt, size, retryCount, statusCode, elapsed, false, cfg, contentType, body, lastErr)
			return nil, lastErr
		}

		var parsed struct {
			Data []struct {
				B64JSON string `json:"b64_json"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			lastErr = fmt.Errorf("parse gpt-image-2 response: %w; response_preview=%s", err, researchDrawingResponsePreview(body))
			h.logDirectGPTImageRequest(jobID, endpoint, imagePrompt, size, retryCount, statusCode, elapsed, false, cfg, contentType, body, lastErr)
			return nil, lastErr
		}
		if len(parsed.Data) == 0 || strings.TrimSpace(parsed.Data[0].B64JSON) == "" {
			lastErr = fmt.Errorf("gpt-image-2 response missing data[0].b64_json")
			h.logDirectGPTImageRequest(jobID, endpoint, imagePrompt, size, retryCount, statusCode, elapsed, false, cfg, contentType, body, lastErr)
			return nil, lastErr
		}
		b64JSON := strings.TrimSpace(parsed.Data[0].B64JSON)
		imageBytes, err := decodeResearchDrawingImageBase64(b64JSON)
		if err != nil {
			lastErr = fmt.Errorf("decode data[0].b64_json: %w", err)
			h.logDirectGPTImageRequest(jobID, endpoint, imagePrompt, size, retryCount, statusCode, elapsed, false, cfg, contentType, body, lastErr)
			return nil, lastErr
		}
		if !isValidResearchDrawingImage(imageBytes, "") {
			lastErr = fmt.Errorf("decode data[0].b64_json: no valid image bytes")
			h.logDirectGPTImageRequest(jobID, endpoint, imagePrompt, size, retryCount, statusCode, elapsed, false, cfg, contentType, body, lastErr)
			return nil, lastErr
		}
		h.logDirectGPTImageRequest(jobID, endpoint, imagePrompt, size, retryCount, statusCode, elapsed, true, cfg, contentType, nil, nil)
		log.Printf("[ResearchDrawing] direct GPT image parsed job_id=%s image_field=data[0].b64_json request_url=%s size=%s retry_count=%d b64_len=%d decoded_bytes=%d", jobID, endpoint, size, retryCount, len(b64JSON), len(imageBytes))
		return imageBytes, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("gpt-image-2 image request failed")
}

func (h *ResearchDrawingHandler) postResearchDrawingGPTJSON(ctx context.Context, endpoint, apiKey string, payload any) ([]byte, string, int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, "", 0, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, "", 0, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	client := researchDrawingHTTPClientWithTimeout(h.httpClient, researchDrawingGPTImage2Timeout)
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, "", 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header.Get("Content-Type"), resp.StatusCode, err
	}
	return respBody, resp.Header.Get("Content-Type"), resp.StatusCode, nil
}

func (h *ResearchDrawingHandler) logDirectGPTImageRequest(jobID, endpoint, prompt, size string, retryCount, statusCode int, elapsed time.Duration, hasValidImage bool, cfg researchDrawingDirectGPTConfig, contentType string, body []byte, err error) {
	errMessage := ""
	if err != nil {
		errMessage = err.Error()
	}
	responsePreview := ""
	if len(body) > 0 && !hasValidImage {
		responsePreview = researchDrawingResponsePreview(body)
	}
	log.Printf(
		"[ResearchDrawing] direct GPT image request job_id=%s final_image_url=%s prompt_length=%d size=%s timeout_seconds=%d retry_count=%d status_code=%d elapsed_ms=%d has_valid_image=%t base_url_source=%s key_source=%s key_len=%d content_type=%s error=%s response_preview=%s",
		jobID,
		endpoint,
		utf8.RuneCountInString(prompt),
		size,
		int(researchDrawingGPTImage2Timeout/time.Second),
		retryCount,
		statusCode,
		elapsed.Milliseconds(),
		hasValidImage,
		cfg.BaseURLSource,
		cfg.KeySource,
		len(cfg.ImageAPIKey),
		contentType,
		errMessage,
		responsePreview,
	)
}

func isResearchDrawingRetryableDirectFailure(statusCode int, err error) bool {
	switch statusCode {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout, 524:
		return true
	}
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	type timeoutError interface {
		Timeout() bool
	}
	var timeoutErr timeoutError
	if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
		return true
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "timeout") || strings.Contains(errText, "context deadline exceeded")
}

func researchDrawingHTTPClientWithTimeout(base *http.Client, timeout time.Duration) *http.Client {
	if base == nil {
		return &http.Client{Timeout: timeout}
	}
	return &http.Client{
		Transport:     base.Transport,
		CheckRedirect: base.CheckRedirect,
		Jar:           base.Jar,
		Timeout:       timeout,
	}
}

func (h *ResearchDrawingHandler) failDirectJob(jobID string, err error) {
	technicalMessage := "direct GPT generation failed"
	if err != nil {
		technicalMessage = err.Error()
	}
	userMessage := researchDrawingFriendlyUpstreamErrorMessage(technicalMessage)
	userID := int64(0)
	resolvedPrice := float64(0)
	imageModelName := researchDrawingGPTImage2ModelName
	h.mu.Lock()
	job, ok := h.jobs[jobID]
	if ok && job.Direct {
		userID = job.UserID
		resolvedPrice = job.ResolvedPrice
		if resolvedPrice <= 0 {
			resolvedPrice = job.Charge
		}
		if strings.TrimSpace(job.ImageGenModelName) != "" {
			imageModelName = job.ImageGenModelName
		}
		job.Status = "error"
		job.Error = userMessage
		job.FinishedAt = time.Now()
		h.jobs[jobID] = job
	}
	h.mu.Unlock()
	if userID > 0 {
		h.refundJobOnceWithContext(context.Background(), userID, jobID)
		h.logBillingEvent("final_status", jobID, userID, imageModelName, resolvedPrice, 0, 0, "error", err)
	}
	log.Printf("[ResearchDrawing] direct GPT generation failed job_id=%s image_gen_model_name=%s error=%s user_error=%s", jobID, imageModelName, technicalMessage, userMessage)
}

func (h *ResearchDrawingHandler) getDirectJobStatus(userID int64, jobID string) (map[string]any, bool, error) {
	h.mu.Lock()
	job, ok := h.jobs[jobID]
	if !ok || !job.Direct {
		h.mu.Unlock()
		return nil, false, nil
	}
	if job.UserID != userID {
		h.mu.Unlock()
		return nil, true, infraerrors.New(http.StatusNotFound, "RESEARCH_DRAWING_JOB_NOT_FOUND", "research drawing job not found")
	}
	status := strings.TrimSpace(job.Status)
	if status == "" {
		status = "running"
	}
	elapsed := 0
	if !job.StartedAt.IsZero() {
		elapsed = int(time.Since(job.StartedAt).Seconds())
	}
	candidateIDs := []int{}
	images := []map[string]any{}
	if status == "done" {
		if _, ok := job.Images[0]; ok {
			candidateIDs = append(candidateIDs, 0)
			images = append(images, map[string]any{"candidate_id": 0})
		}
	}
	out := map[string]any{
		"ok":              true,
		"job_id":          jobID,
		"username":        job.PaperBananaUser,
		"status":          status,
		"mode":            "direct_gpt",
		"elapsed_sec":     elapsed,
		"candidate_count": len(candidateIDs),
		"candidate_ids":   candidateIDs,
		"images":          images,
	}
	if job.Error != "" {
		out["error"] = job.Error
	}
	h.mu.Unlock()
	return out, true, nil
}

func (h *ResearchDrawingHandler) getDirectJobImage(ctx context.Context, userID int64, jobID, candidateID string) ([]byte, string, bool, error) {
	h.mu.Lock()
	job, ok := h.jobs[jobID]
	if !ok || !job.Direct {
		h.mu.Unlock()
		return h.getPersistedDirectJobImage(ctx, userID, jobID, candidateID)
	}
	if job.UserID != userID {
		h.mu.Unlock()
		return nil, "", true, infraerrors.New(http.StatusNotFound, "RESEARCH_DRAWING_JOB_NOT_FOUND", "research drawing job not found")
	}
	if candidateID != "0" {
		h.mu.Unlock()
		return nil, "", true, infraerrors.New(http.StatusNotFound, "RESEARCH_DRAWING_IMAGE_NOT_FOUND", "research drawing image not found")
	}
	if !strings.EqualFold(job.Status, "done") {
		errMessage := job.Error
		if errMessage == "" {
			errMessage = "research drawing image is not ready"
		}
		h.mu.Unlock()
		return nil, "", true, infraerrors.New(http.StatusConflict, "RESEARCH_DRAWING_IMAGE_NOT_READY", errMessage)
	}
	image, ok := job.Images[0]
	if !ok || len(image.Bytes) == 0 {
		h.mu.Unlock()
		return nil, "", true, infraerrors.New(http.StatusNotFound, "RESEARCH_DRAWING_IMAGE_NOT_FOUND", "research drawing image not found")
	}
	body := append([]byte(nil), image.Bytes...)
	contentType := image.ContentType
	h.mu.Unlock()
	return body, contentType, true, nil
}

func (h *ResearchDrawingHandler) getPersistedDirectJobImage(ctx context.Context, userID int64, jobID, candidateID string) ([]byte, string, bool, error) {
	if h.image2RecordService == nil {
		return nil, "", false, nil
	}
	if candidateID != "0" {
		return nil, "", false, nil
	}
	record, err := h.image2RecordService.GetByUserJob(ctx, userID, jobID)
	if err != nil {
		return nil, "", true, infraerrors.New(http.StatusInternalServerError, "RESEARCH_DRAWING_IMAGE2_RECORD_LOOKUP_FAILED", err.Error())
	}
	if record == nil {
		return nil, "", false, nil
	}
	body, err := loadResearchDrawingDirectCandidate(jobID)
	if err != nil {
		return nil, "", true, infraerrors.New(http.StatusNotFound, "RESEARCH_DRAWING_IMAGE_NOT_FOUND", "research drawing image not found")
	}
	if !isValidResearchDrawingImage(body, "") {
		return nil, "", true, infraerrors.New(http.StatusNotFound, "RESEARCH_DRAWING_IMAGE_NOT_FOUND", "research drawing image not found")
	}
	return body, "image/png", true, nil
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
	if r.MaxRefineResolution != "1K" && r.MaxRefineResolution != "4K" {
		r.MaxRefineResolution = "2K"
	}
	r.ImageGenModelName = strings.TrimSpace(r.ImageGenModelName)
	if len(r.ImageGenModelName) > 200 {
		r.ImageGenModelName = r.ImageGenModelName[:200]
	}
	switch r.ImageGenModelName {
	case researchDrawingDefaultImageModelName:
		if r.MaxRefineResolution == "1K" {
			r.MaxRefineResolution = "2K"
		}
	case researchDrawingGPTImage2ModelName:
		r.forceDirectGPTMode()
	default:
		r.ImageGenModelName = researchDrawingDefaultImageModelName
		if r.MaxRefineResolution == "1K" {
			r.MaxRefineResolution = "2K"
		}
	}
}

func (r ResearchDrawingGenerateRequest) isGPTImage2() bool {
	return strings.TrimSpace(r.ImageGenModelName) == researchDrawingGPTImage2ModelName
}

func (r ResearchDrawingGenerateRequest) isDirectGPTMode() bool {
	return r.isGPTImage2()
}

func (r *ResearchDrawingGenerateRequest) forceDirectGPTMode() {
	r.GenerationMode = "default"
	r.ExpMode = "demo_planner_critic"
	r.RetrievalSetting = "none"
	r.NumCandidates = 1
	r.MaxCriticRounds = 0
	r.MaxRefineResolution = "1K"
}

func (r ResearchDrawingGenerateRequest) directQuotaNeed() int {
	return 1
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

func resolveResearchDrawingPrice(imageModelName string) float64 {
	switch strings.TrimSpace(imageModelName) {
	case researchDrawingGPTImage2ModelName:
		return researchDrawingGPTImage2UnitPrice
	case researchDrawingDefaultImageModelName:
		return researchDrawingNanoBanana2UnitPrice
	default:
		return researchDrawingUnitPrice
	}
}

func normalizeResearchDrawingImageModelNameForLog(imageModelName string) string {
	trimmed := strings.TrimSpace(imageModelName)
	if trimmed == "" {
		return researchDrawingDefaultImageModelName
	}
	return trimmed
}

func (h *ResearchDrawingHandler) logJobFinalStatus(userID int64, jobID, finalStatus string) {
	h.mu.Lock()
	job, ok := h.jobs[jobID]
	h.mu.Unlock()
	if !ok || job.UserID != userID {
		return
	}
	resolvedPrice := job.ResolvedPrice
	if resolvedPrice <= 0 {
		resolvedPrice = job.Charge
	}
	h.logBillingEvent("final_status", jobID, userID, job.ImageGenModelName, resolvedPrice, 0, 0, finalStatus, nil)
}

func (h *ResearchDrawingHandler) logBillingEvent(event, jobID string, userID int64, imageModelName string, resolvedPrice, chargeAmount, refundAmount float64, finalStatus string, err error) {
	errMessage := ""
	if err != nil {
		errMessage = err.Error()
	}
	log.Printf(
		"[ResearchDrawing] billing event=%s job_id=%s user_id=%d image_gen_model_name=%s resolved_price=%.4f charge_amount=%.4f refund_amount=%.4f final_status=%s error=%s",
		event,
		jobID,
		userID,
		normalizeResearchDrawingImageModelNameForLog(imageModelName),
		resolvedPrice,
		chargeAmount,
		refundAmount,
		finalStatus,
		errMessage,
	)
}

func (h *ResearchDrawingHandler) researchDrawingUserFacingError(event, jobID, imageModelName string, err error) string {
	if err == nil {
		return ""
	}
	technicalMessage := err.Error()
	userMessage := researchDrawingFriendlyUpstreamErrorMessage(technicalMessage)
	if userMessage != technicalMessage {
		log.Printf("[ResearchDrawing] upstream technical error event=%s job_id=%s image_gen_model_name=%s error=%s user_error=%s", event, jobID, normalizeResearchDrawingImageModelNameForLog(imageModelName), technicalMessage, userMessage)
	}
	return userMessage
}

func applyResearchDrawingFriendlyStatusError(status map[string]any) {
	if status == nil {
		return
	}
	for _, key := range []string{"error", "message", "detail"} {
		raw := strings.TrimSpace(stringFromMap(status, key))
		if raw == "" {
			continue
		}
		if friendly := researchDrawingFriendlyUpstreamErrorMessage(raw); friendly != raw {
			status["error"] = friendly
			return
		}
	}
}

func researchDrawingFriendlyUpstreamErrorMessage(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return trimmed
	}
	if researchDrawingErrorLooksLike524(trimmed) {
		return researchDrawingUpstreamTimeoutMessage
	}
	if researchDrawingErrorLooksLike502(trimmed) {
		return researchDrawingUpstreamBusyMessage
	}
	return trimmed
}

func researchDrawingErrorLooksLike502(raw string) bool {
	lower := strings.ToLower(raw)
	compact := strings.ReplaceAll(lower, " ", "")
	if lower == "502" ||
		strings.Contains(compact, "status_code=502") ||
		strings.Contains(compact, "status_code:502") ||
		strings.Contains(compact, "status_code\":502") {
		return true
	}
	for _, pattern := range []string{
		"status_code=502",
		"error code: 502",
		"upstream request failed",
		"upstream_error",
		"returned 502",
		"status code: 502",
		"status code 502",
		"status 502",
		"code 502",
		"http 502",
		" 502",
		"502 ",
	} {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func researchDrawingErrorLooksLike524(raw string) bool {
	lower := strings.ToLower(raw)
	compact := strings.ReplaceAll(lower, " ", "")
	if lower == "524" ||
		strings.Contains(compact, "status_code=524") ||
		strings.Contains(compact, "status_code:524") ||
		strings.Contains(compact, "status_code\":524") {
		return true
	}
	for _, pattern := range []string{
		"status_code=524",
		"error code: 524",
		"returned 524",
		"status code: 524",
		"status code 524",
		"status 524",
		"code 524",
		"http 524",
		" 524",
		"524 ",
	} {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func (h *ResearchDrawingHandler) researchDrawingDirectGPTConfig(_ context.Context, _ ResearchDrawingGenerateRequest) (researchDrawingDirectGPTConfig, error) {
	apiKey, keySource := firstNonEmptyResearchDrawingValue(
		"GPT_IMAGE_API_KEY", os.Getenv("GPT_IMAGE_API_KEY"),
		"GPT_API_KEY", os.Getenv("GPT_API_KEY"),
	)
	baseURL, baseURLSource := firstNonEmptyResearchDrawingValue(
		"GPT_IMAGE_BASE_URL", os.Getenv("GPT_IMAGE_BASE_URL"),
		"GPT_BASE_URL", os.Getenv("GPT_BASE_URL"),
	)
	cfg := researchDrawingDirectGPTConfig{
		ImageAPIKey:   apiKey,
		ImageBaseURL:  strings.TrimRight(baseURL, "/"),
		KeySource:     keySource,
		BaseURLSource: baseURLSource,
	}
	if cfg.ImageAPIKey == "" {
		return cfg, fmt.Errorf("GPT_IMAGE_API_KEY or GPT_API_KEY is required for gpt-image-2 direct mode")
	}
	if cfg.ImageBaseURL == "" {
		return cfg, fmt.Errorf("GPT_IMAGE_BASE_URL or GPT_BASE_URL is required for gpt-image-2 direct mode")
	}
	return cfg, nil
}

func firstNonEmptyResearchDrawingValue(nameValuePairs ...string) (string, string) {
	for i := 0; i+1 < len(nameValuePairs); i += 2 {
		if trimmed := strings.TrimSpace(nameValuePairs[i+1]); trimmed != "" {
			return trimmed, nameValuePairs[i]
		}
	}
	return "", ""
}

func buildResearchDrawingDirectImagePrompt(req ResearchDrawingGenerateRequest) string {
	methodContent := strings.TrimSpace(req.MethodContent)
	caption := strings.TrimSpace(req.Caption)
	if caption == "" {
		return methodContent
	}
	if methodContent == "" {
		return caption
	}
	return fmt.Sprintf("Caption:\n%s\n\nContent:\n%s", caption, methodContent)
}

func researchDrawingDirectImageSize(_ string) string {
	return researchDrawingGPTImage2DirectSize
}

func researchDrawingResponsePreview(body []byte) string {
	preview := string(body)
	preview = strings.ReplaceAll(preview, "\n", "\\n")
	preview = strings.ReplaceAll(preview, "\r", "\\r")
	if len(preview) > 500 {
		preview = preview[:500]
	}
	return preview
}

func researchDrawingStatusFailed(status map[string]any) bool {
	switch strings.ToLower(strings.TrimSpace(stringFromMap(status, "status"))) {
	case "error", "failed", "fail", "canceled", "cancelled":
		return true
	default:
		return false
	}
}

func researchDrawingStatusDone(status map[string]any) bool {
	switch strings.ToLower(strings.TrimSpace(stringFromMap(status, "status"))) {
	case "done", "success", "succeeded", "completed", "complete", "finished":
		return true
	default:
		return false
	}
}

func researchDrawingStatusHasCandidates(status map[string]any) bool {
	if images, ok := status["images"].([]any); ok && len(images) > 0 {
		return true
	}
	if candidateIDs, ok := status["candidate_ids"].([]any); ok && len(candidateIDs) > 0 {
		return true
	}
	if count := intFromMap(status, "candidate_count"); count > 0 {
		return true
	}
	return false
}

func isValidResearchDrawingImage(body []byte, contentType string) bool {
	if len(body) == 0 {
		return false
	}
	normalizedContentType := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if strings.HasPrefix(normalizedContentType, "image/") {
		return true
	}
	if len(body) >= 8 && bytes.Equal(body[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		return true
	}
	if len(body) >= 3 && body[0] == 0xff && body[1] == 0xd8 && body[2] == 0xff {
		return true
	}
	if len(body) >= 12 && string(body[:4]) == "RIFF" && string(body[8:12]) == "WEBP" {
		return true
	}
	if len(body) >= 6 && (string(body[:6]) == "GIF87a" || string(body[:6]) == "GIF89a") {
		return true
	}
	return false
}

func decodeResearchDrawingImageBase64(raw string) ([]byte, error) {
	value := strings.TrimSpace(raw)
	if idx := strings.Index(value, ","); idx >= 0 {
		value = value[idx+1:]
	}
	return base64.StdEncoding.DecodeString(value)
}

func saveResearchDrawingDirectCandidate(jobID string, imageBytes []byte) (string, error) {
	outPath := researchDrawingDirectCandidatePath(jobID)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(outPath, imageBytes, 0o644); err != nil {
		return "", err
	}
	return outPath, nil
}

func loadResearchDrawingDirectCandidate(jobID string) ([]byte, error) {
	return os.ReadFile(researchDrawingDirectCandidatePath(jobID))
}

func researchDrawingDirectCandidatePath(jobID string) string {
	root := strings.TrimSpace(os.Getenv("RESEARCH_DRAWING_DIRECT_RESULTS_DIR"))
	if root == "" {
		if dataDir := strings.TrimSpace(os.Getenv("DATA_DIR")); dataDir != "" {
			root = filepath.Join(dataDir, "research-drawing", "results")
		} else {
			root = filepath.Join(os.TempDir(), "sub2api-research-drawing-results")
		}
	}
	return filepath.Join(root, jobID, "candidate_0.png")
}

func newResearchDrawingJobID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return "rdgpt_" + hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("rdgpt_%d", time.Now().UnixNano())
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

func intFromMap(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch typed := v.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		value, _ := typed.Int64()
		return int(value)
	case string:
		value, _ := strconv.Atoi(strings.TrimSpace(typed))
		return value
	default:
		return 0
	}
}
