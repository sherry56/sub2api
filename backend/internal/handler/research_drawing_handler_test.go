package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/model"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type researchDrawingSettingRepoStub struct {
	values map[string]string
}

func (r *researchDrawingSettingRepoStub) Get(ctx context.Context, key string) (*service.Setting, error) {
	if value, ok := r.values[key]; ok {
		return &service.Setting{Key: key, Value: value}, nil
	}
	return nil, service.ErrSettingNotFound
}

func (r *researchDrawingSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if value, ok := r.values[key]; ok {
		return value, nil
	}
	return "", service.ErrSettingNotFound
}

func (r *researchDrawingSettingRepoStub) Set(ctx context.Context, key, value string) error {
	if r.values == nil {
		r.values = make(map[string]string)
	}
	r.values[key] = value
	return nil
}

func (r *researchDrawingSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (r *researchDrawingSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	if r.values == nil {
		r.values = make(map[string]string)
	}
	for key, value := range settings {
		r.values[key] = value
	}
	return nil
}

func (r *researchDrawingSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	out := make(map[string]string, len(r.values))
	for key, value := range r.values {
		out[key] = value
	}
	return out, nil
}

func (r *researchDrawingSettingRepoStub) Delete(ctx context.Context, key string) error {
	delete(r.values, key)
	return nil
}

type researchDrawingImage2RecordRepoStub struct {
	records []model.ResearchDrawingImage2Record
}

func (r *researchDrawingImage2RecordRepoStub) CreateResearchDrawingImage2Record(ctx context.Context, record model.ResearchDrawingImage2Record) error {
	r.records = append(r.records, record)
	return nil
}

func (r *researchDrawingImage2RecordRepoStub) ListResearchDrawingImage2Records(ctx context.Context, userID int64, limit int) ([]model.ResearchDrawingImage2Record, error) {
	return r.records, nil
}

func (r *researchDrawingImage2RecordRepoStub) GetResearchDrawingImage2Record(ctx context.Context, userID int64, jobID string) (*model.ResearchDrawingImage2Record, error) {
	for i := range r.records {
		if r.records[i].UserID == userID && r.records[i].JobID == jobID {
			return &r.records[i], nil
		}
	}
	return nil, nil
}

type researchDrawingRefundUserRepoStub struct {
	service.UserRepository
	updates []researchDrawingBalanceUpdate
}

type researchDrawingBalanceUpdate struct {
	userID int64
	amount float64
}

func (r *researchDrawingRefundUserRepoStub) UpdateBalance(ctx context.Context, userID int64, amount float64) error {
	r.updates = append(r.updates, researchDrawingBalanceUpdate{userID: userID, amount: amount})
	return nil
}

func TestResearchDrawingGPTImage2UsesDirectModeConfigFromEnv(t *testing.T) {
	t.Setenv("GPT_API_KEY", "sk-gpt-fallback")
	t.Setenv("GPT_BASE_URL", "https://fallback.example/v1")
	t.Setenv("GPT_IMAGE_API_KEY", "sk-image-env")
	t.Setenv("GPT_IMAGE_BASE_URL", "https://image.example/v1/")

	settingSvc := service.NewSettingService(&researchDrawingSettingRepoStub{values: map[string]string{
		service.SettingKeyResearchDrawingGPTImageAPIKey:  "sk-from-settings",
		service.SettingKeyResearchDrawingGPTImageBaseURL: "https://api.openai.com/v1",
	}}, &config.Config{})
	handler := NewResearchDrawingHandler(nil, settingSvc, nil)

	req := ResearchDrawingGenerateRequest{
		MethodContent:     "method",
		ImageGenModelName: researchDrawingGPTImage2ModelName,
	}
	req.normalize()

	if !req.isDirectGPTMode() {
		t.Fatal("gpt-image-2 must use direct GPT mode instead of PaperBanana")
	}
	req.forceDirectGPTMode()
	if req.NumCandidates != 1 || req.MaxCriticRounds != 0 || req.RetrievalSetting != "none" || req.MaxRefineResolution != "1K" {
		t.Fatalf("direct image mode did not collapse PaperBanana parameters: candidates=%d rounds=%d retrieval=%q resolution=%q", req.NumCandidates, req.MaxCriticRounds, req.RetrievalSetting, req.MaxRefineResolution)
	}

	cfg, err := handler.researchDrawingDirectGPTConfig(context.Background(), req)
	if err != nil {
		t.Fatalf("researchDrawingDirectGPTConfig returned error: %v", err)
	}
	if cfg.ImageAPIKey != "sk-image-env" {
		t.Fatalf("ImageAPIKey = %q, want GPT_IMAGE_API_KEY", cfg.ImageAPIKey)
	}
	if cfg.ImageBaseURL != "https://image.example/v1" {
		t.Fatalf("ImageBaseURL = %q, want trimmed GPT_IMAGE_BASE_URL", cfg.ImageBaseURL)
	}
	if cfg.KeySource != "GPT_IMAGE_API_KEY" || cfg.BaseURLSource != "GPT_IMAGE_BASE_URL" {
		t.Fatalf("sources = (%q, %q), want GPT_IMAGE sources", cfg.KeySource, cfg.BaseURLSource)
	}
}

func TestResearchDrawingGPTImage2FallsBackToGPTEnv(t *testing.T) {
	t.Setenv("GPT_API_KEY", "sk-gpt-fallback")
	t.Setenv("GPT_BASE_URL", "https://fallback.example/v1/")
	t.Setenv("GPT_IMAGE_API_KEY", "")
	t.Setenv("GPT_IMAGE_BASE_URL", "")

	handler := NewResearchDrawingHandler(nil, nil, nil)
	req := ResearchDrawingGenerateRequest{
		MethodContent:     "method",
		ImageGenModelName: researchDrawingGPTImage2ModelName,
	}
	req.normalize()

	cfg, err := handler.researchDrawingDirectGPTConfig(context.Background(), req)
	if err != nil {
		t.Fatalf("researchDrawingDirectGPTConfig returned error: %v", err)
	}
	if cfg.ImageAPIKey != "sk-gpt-fallback" {
		t.Fatalf("ImageAPIKey = %q, want GPT_API_KEY fallback", cfg.ImageAPIKey)
	}
	if cfg.ImageBaseURL != "https://fallback.example/v1" {
		t.Fatalf("ImageBaseURL = %q, want trimmed GPT_BASE_URL fallback", cfg.ImageBaseURL)
	}
	if cfg.KeySource != "GPT_API_KEY" || cfg.BaseURLSource != "GPT_BASE_URL" {
		t.Fatalf("sources = (%q, %q), want GPT fallback sources", cfg.KeySource, cfg.BaseURLSource)
	}
}

func TestResearchDrawingGPTImage2DoesNotDefaultToOpenAIBaseURL(t *testing.T) {
	t.Setenv("GPT_API_KEY", "")
	t.Setenv("GPT_BASE_URL", "")
	t.Setenv("GPT_IMAGE_API_KEY", "")
	t.Setenv("GPT_IMAGE_BASE_URL", "")

	settingSvc := service.NewSettingService(&researchDrawingSettingRepoStub{values: map[string]string{
		service.SettingKeyResearchDrawingGPTImageAPIKey:  "sk-from-settings",
		service.SettingKeyResearchDrawingGPTImageBaseURL: "https://api.openai.com/v1",
	}}, &config.Config{})
	handler := NewResearchDrawingHandler(nil, settingSvc, nil)
	req := ResearchDrawingGenerateRequest{
		MethodContent:     "method",
		ImageGenModelName: researchDrawingGPTImage2ModelName,
	}
	req.normalize()

	cfg, err := handler.researchDrawingDirectGPTConfig(context.Background(), req)
	if err == nil {
		t.Fatalf("researchDrawingDirectGPTConfig returned nil error with empty env and cfg=%+v", cfg)
	}
	if strings.Contains(cfg.ImageBaseURL, "api.openai.com") {
		t.Fatalf("ImageBaseURL = %q, want no default OpenAI fallback", cfg.ImageBaseURL)
	}
}

func TestResearchDrawingGPTImage2IgnoresGPT55MainModel(t *testing.T) {
	t.Setenv("GPT_API_KEY", "")
	t.Setenv("GPT_BASE_URL", "")
	t.Setenv("GPT_IMAGE_API_KEY", "")
	t.Setenv("GPT_IMAGE_BASE_URL", "")
	t.Setenv("RESEARCH_DRAWING_DIRECT_RESULTS_DIR", t.TempDir())

	textRequests := 0
	imageRequests := 0
	const pngB64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAAAAAA6fptVAAAACklEQVR42mNggAAAAAMAASsJTYQAAAAASUVORK5CYII="
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/chat/completions":
			textRequests++
			http.Error(w, "text model must not be called", http.StatusInternalServerError)
		case "/images/generations":
			imageRequests++
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode image request payload: %v", err)
				http.Error(w, "bad json", http.StatusBadRequest)
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer sk-image" {
				t.Errorf("Authorization = %q, want Bearer sk-image", got)
			}
			if len(payload) != 3 {
				t.Errorf("image request payload keys = %v, want only model/prompt/size", payload)
			}
			if payload["model"] != researchDrawingGPTImage2ModelName {
				t.Errorf("image model = %v, want %s", payload["model"], researchDrawingGPTImage2ModelName)
			}
			if payload["size"] != researchDrawingGPTImage2DirectSize {
				t.Errorf("image request size = %v, want %s", payload["size"], researchDrawingGPTImage2DirectSize)
			}
			prompt, _ := payload["prompt"].(string)
			if !strings.Contains(prompt, "raw method content") || !strings.Contains(prompt, "raw caption") {
				t.Errorf("prompt = %q, want raw method content and caption", prompt)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]string{{"b64_json": pngB64}},
			})
		default:
			t.Errorf("unexpected request path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	recordRepo := &researchDrawingImage2RecordRepoStub{}
	recordSvc := service.NewResearchDrawingImage2RecordService(recordRepo)
	handler := NewResearchDrawingHandler(nil, nil, recordSvc)
	handler.httpClient = server.Client()

	req := ResearchDrawingGenerateRequest{
		MethodContent:     "raw method content",
		Caption:           "raw caption",
		ImageGenModelName: researchDrawingGPTImage2ModelName,
		NumCandidates:     8,
		MaxCriticRounds:   5,
		RetrievalSetting:  "manual",
	}
	req.normalize()
	if !req.isDirectGPTMode() {
		t.Fatal("gpt-image-2 must use direct GPT mode")
	}
	req.forceDirectGPTMode()
	if req.NumCandidates != 1 || req.MaxCriticRounds != 0 || req.RetrievalSetting != "none" || req.MaxRefineResolution != "1K" {
		t.Fatalf("direct image mode did not collapse PaperBanana parameters: candidates=%d rounds=%d retrieval=%q resolution=%q", req.NumCandidates, req.MaxCriticRounds, req.RetrievalSetting, req.MaxRefineResolution)
	}

	jobID := "job-test-gpt-image-2"
	handler.jobs[jobID] = researchDrawingJobCharge{
		UserID:    1,
		Direct:    true,
		Status:    "running",
		StartedAt: time.Now(),
		Images:    make(map[int]researchDrawingDirectImage),
	}
	handler.runDirectGPTResearchDrawingJob(jobID, req, researchDrawingDirectGPTConfig{
		ImageAPIKey:  "sk-image",
		ImageBaseURL: server.URL,
	})

	if textRequests != 0 {
		t.Fatalf("text model requests = %d, want 0", textRequests)
	}
	if imageRequests != 1 {
		t.Fatalf("image generation requests = %d, want 1", imageRequests)
	}
	job := handler.jobs[jobID]
	if job.Status != "done" {
		t.Fatalf("job status = %q, error = %q, want done", job.Status, job.Error)
	}
	if _, ok := job.Images[0]; !ok {
		t.Fatal("candidate_0 image was not saved on the direct job")
	}
	if len(recordRepo.records) != 1 {
		t.Fatalf("image2 records = %d, want 1", len(recordRepo.records))
	}
	if recordRepo.records[0].UserID != 1 || recordRepo.records[0].JobID != jobID || recordRepo.records[0].Model != researchDrawingGPTImage2ModelName {
		t.Fatalf("unexpected image2 record: %+v", recordRepo.records[0])
	}
}

func TestResearchDrawingGPTImage2RetriesRetryableImageGenerationOnce(t *testing.T) {
	t.Setenv("RESEARCH_DRAWING_DIRECT_RESULTS_DIR", t.TempDir())

	imageRequests := 0
	const pngB64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAAAAAA6fptVAAAACklEQVR42mNggAAAAAMAASsJTYQAAAAASUVORK5CYII="
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/generations" {
			t.Errorf("unexpected request path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		imageRequests++
		if imageRequests == 1 {
			http.Error(w, "temporary upstream failure", http.StatusServiceUnavailable)
			return
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode image request payload: %v", err)
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if len(payload) != 3 {
			t.Errorf("image request payload keys = %v, want only model/prompt/size", payload)
		}
		if payload["size"] != researchDrawingGPTImage2DirectSize {
			t.Errorf("image request size = %v, want %s", payload["size"], researchDrawingGPTImage2DirectSize)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{{"b64_json": pngB64}},
		})
	}))
	defer server.Close()

	handler := NewResearchDrawingHandler(nil, nil, nil)
	handler.httpClient = server.Client()
	req := ResearchDrawingGenerateRequest{
		MethodContent:     "method",
		ImageGenModelName: researchDrawingGPTImage2ModelName,
	}
	req.normalize()

	jobID := "job-test-gpt-image-2-retry"
	handler.jobs[jobID] = researchDrawingJobCharge{
		UserID:    1,
		Direct:    true,
		Status:    "running",
		StartedAt: time.Now(),
		Images:    make(map[int]researchDrawingDirectImage),
	}
	handler.runDirectGPTResearchDrawingJob(jobID, req, researchDrawingDirectGPTConfig{
		ImageAPIKey:  "sk-image",
		ImageBaseURL: server.URL,
	})

	if imageRequests != 2 {
		t.Fatalf("image generation requests = %d, want 2", imageRequests)
	}
	job := handler.jobs[jobID]
	if job.Status != "done" {
		t.Fatalf("job status = %q, error = %q, want done", job.Status, job.Error)
	}
	if _, ok := job.Images[0]; !ok {
		t.Fatal("candidate_0 image was not saved after retry success")
	}
}

func TestResearchDrawingGPTImage2RetriesOnceThenRefundsAndDoesNotSaveEmptyImage(t *testing.T) {
	t.Setenv("RESEARCH_DRAWING_DIRECT_RESULTS_DIR", t.TempDir())

	imageRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/generations" {
			t.Errorf("unexpected request path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		imageRequests++
		http.Error(w, "temporary upstream failure", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	userRepo := &researchDrawingRefundUserRepoStub{}
	userSvc := service.NewUserService(userRepo, nil, nil, nil)
	handler := NewResearchDrawingHandler(userSvc, nil, nil)
	handler.httpClient = server.Client()
	req := ResearchDrawingGenerateRequest{
		MethodContent:     "method",
		ImageGenModelName: researchDrawingGPTImage2ModelName,
	}
	req.normalize()

	jobID := "job-test-gpt-image-2-refund"
	handler.jobs[jobID] = researchDrawingJobCharge{
		UserID:    7,
		Charge:    researchDrawingGPTImage2UnitPrice,
		Direct:    true,
		Status:    "running",
		StartedAt: time.Now(),
		Images:    make(map[int]researchDrawingDirectImage),
	}
	handler.runDirectGPTResearchDrawingJob(jobID, req, researchDrawingDirectGPTConfig{
		ImageAPIKey:  "sk-image",
		ImageBaseURL: server.URL,
	})

	if imageRequests != 2 {
		t.Fatalf("image generation requests = %d, want 2", imageRequests)
	}
	job := handler.jobs[jobID]
	if job.Status != "error" {
		t.Fatalf("job status = %q, want error", job.Status)
	}
	if !job.Refunded {
		t.Fatal("failed direct GPT job was not marked refunded")
	}
	if len(userRepo.updates) != 1 {
		t.Fatalf("balance updates = %d, want 1", len(userRepo.updates))
	}
	if userRepo.updates[0].userID != 7 || userRepo.updates[0].amount != researchDrawingGPTImage2UnitPrice {
		t.Fatalf("unexpected refund update: %+v", userRepo.updates[0])
	}
	if len(job.Images) != 0 {
		t.Fatalf("saved direct images = %d, want 0", len(job.Images))
	}
	if _, err := loadResearchDrawingDirectCandidate(jobID); err == nil {
		t.Fatal("failed direct GPT job saved an image file")
	}
}

func TestResearchDrawingGPTImage2DirectHTTPClientUses300SecondTimeout(t *testing.T) {
	base := &http.Client{Timeout: time.Second}
	got := researchDrawingHTTPClientWithTimeout(base, researchDrawingGPTImage2Timeout)
	if got.Timeout != 300*time.Second {
		t.Fatalf("direct GPT client timeout = %s, want 300s", got.Timeout)
	}
	if base.Timeout != time.Second {
		t.Fatalf("base client timeout was mutated to %s", base.Timeout)
	}
}

func TestResearchDrawingGeminiImageModelKeepsPaperBananaPath(t *testing.T) {
	req := ResearchDrawingGenerateRequest{
		MethodContent:       "method",
		ImageGenModelName:   researchDrawingDefaultImageModelName,
		MaxRefineResolution: "1K",
	}
	req.normalize()
	if req.isDirectGPTMode() {
		t.Fatal("Gemini/OpenRouter image model must stay on the PaperBanana path")
	}
	if req.MaxRefineResolution != "2K" {
		t.Fatalf("Gemini/OpenRouter max refine resolution = %q, want PaperBanana default 2K", req.MaxRefineResolution)
	}
}
