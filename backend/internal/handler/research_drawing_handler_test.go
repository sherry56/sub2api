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

func TestResearchDrawingGPTImage2UsesDirectModeConfigFromEnv(t *testing.T) {
	t.Setenv("GPT_API_KEY", "sk-gpt-fallback")
	t.Setenv("GPT_BASE_URL", "https://fallback.example/v1")
	t.Setenv("GPT_IMAGE_API_KEY", "sk-image-env")
	t.Setenv("GPT_IMAGE_BASE_URL", "https://image.example/v1/")

	settingSvc := service.NewSettingService(&researchDrawingSettingRepoStub{values: map[string]string{
		service.SettingKeyResearchDrawingGPTImageAPIKey:  "sk-from-settings",
		service.SettingKeyResearchDrawingGPTImageBaseURL: "https://api.openai.com/v1",
	}}, &config.Config{})
	handler := NewResearchDrawingHandler(nil, settingSvc)

	req := ResearchDrawingGenerateRequest{
		MethodContent:     "method",
		ImageGenModelName: researchDrawingGPTImage2ModelName,
	}
	req.normalize()

	if !req.isDirectGPTMode() {
		t.Fatal("gpt-image-2 must use direct GPT mode instead of PaperBanana")
	}
	req.forceDirectGPTMode()
	if req.NumCandidates != 1 || req.MaxCriticRounds != 0 || req.RetrievalSetting != "none" {
		t.Fatalf("direct image mode did not collapse PaperBanana parameters: candidates=%d rounds=%d retrieval=%q", req.NumCandidates, req.MaxCriticRounds, req.RetrievalSetting)
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

	handler := NewResearchDrawingHandler(nil, nil)
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
	handler := NewResearchDrawingHandler(nil, settingSvc)
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
			if _, ok := payload["size"]; !ok {
				t.Error("image request payload missing size")
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

	handler := NewResearchDrawingHandler(nil, nil)
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
	if req.NumCandidates != 1 || req.MaxCriticRounds != 0 || req.RetrievalSetting != "none" {
		t.Fatalf("direct image mode did not collapse PaperBanana parameters: candidates=%d rounds=%d retrieval=%q", req.NumCandidates, req.MaxCriticRounds, req.RetrievalSetting)
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
}

func TestResearchDrawingGeminiImageModelKeepsPaperBananaPath(t *testing.T) {
	req := ResearchDrawingGenerateRequest{
		MethodContent:     "method",
		ImageGenModelName: researchDrawingDefaultImageModelName,
	}
	req.normalize()
	if req.isDirectGPTMode() {
		t.Fatal("Gemini/OpenRouter image model must stay on the PaperBanana path")
	}
}
