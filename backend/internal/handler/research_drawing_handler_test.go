package handler

import (
	"context"
	"testing"

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

func TestResearchDrawingGPTImage2UsesDirectModeConfigFromSettings(t *testing.T) {
	t.Setenv("GPT_API_KEY", "")
	t.Setenv("GPT_BASE_URL", "")
	t.Setenv("GPT_IMAGE_API_KEY", "")
	t.Setenv("GPT_IMAGE_BASE_URL", "")

	settingSvc := service.NewSettingService(&researchDrawingSettingRepoStub{values: map[string]string{
		service.SettingKeyResearchDrawingGPTImageAPIKey:  "sk-from-settings",
		service.SettingKeyResearchDrawingGPTImageBaseURL: "https://openai.example/v1/",
	}}, &config.Config{})
	handler := NewResearchDrawingHandler(nil, settingSvc)

	req := ResearchDrawingGenerateRequest{
		MethodContent:     "method",
		MainModelName:     researchDrawingDefaultMainModelName,
		ImageGenModelName: researchDrawingGPTImage2ModelName,
	}
	req.normalize()

	if !req.isDirectGPTMode() {
		t.Fatal("gpt-image-2 must use direct GPT mode instead of PaperBanana")
	}

	cfg, err := handler.researchDrawingDirectGPTConfig(context.Background(), req)
	if err != nil {
		t.Fatalf("researchDrawingDirectGPTConfig returned error: %v", err)
	}
	if cfg.ImageAPIKey != "sk-from-settings" {
		t.Fatalf("ImageAPIKey = %q, want settings key", cfg.ImageAPIKey)
	}
	if cfg.ImageBaseURL != "https://openai.example/v1" {
		t.Fatalf("ImageBaseURL = %q, want trimmed settings base URL", cfg.ImageBaseURL)
	}
}

func TestResearchDrawingGPT55ForcesGPTImage2DirectMode(t *testing.T) {
	t.Setenv("GPT_API_KEY", "")
	t.Setenv("GPT_BASE_URL", "")
	t.Setenv("GPT_IMAGE_API_KEY", "")
	t.Setenv("GPT_IMAGE_BASE_URL", "")

	settingSvc := service.NewSettingService(&researchDrawingSettingRepoStub{values: map[string]string{
		service.SettingKeyResearchDrawingGPTImageAPIKey:  "sk-from-settings",
		service.SettingKeyResearchDrawingGPTImageBaseURL: "https://openai.example/v1",
	}}, &config.Config{})
	handler := NewResearchDrawingHandler(nil, settingSvc)

	req := ResearchDrawingGenerateRequest{
		MethodContent:     "method",
		MainModelName:     researchDrawingGPT55ModelName,
		ImageGenModelName: researchDrawingDefaultImageModelName,
		NumCandidates:     8,
		MaxCriticRounds:   5,
		RetrievalSetting:  "manual",
	}
	req.normalize()
	if !req.isDirectGPTMode() {
		t.Fatal("gpt-5.5 must use direct GPT mode instead of PaperBanana")
	}
	req.forceDirectGPTMode()

	if req.ImageGenModelName != researchDrawingGPTImage2ModelName {
		t.Fatalf("ImageGenModelName = %q, want %q", req.ImageGenModelName, researchDrawingGPTImage2ModelName)
	}
	if req.NumCandidates != 1 || req.MaxCriticRounds != 1 || req.RetrievalSetting != "none" {
		t.Fatalf("direct mode did not collapse PaperBanana parameters: candidates=%d rounds=%d retrieval=%q", req.NumCandidates, req.MaxCriticRounds, req.RetrievalSetting)
	}

	cfg, err := handler.researchDrawingDirectGPTConfig(context.Background(), req)
	if err != nil {
		t.Fatalf("researchDrawingDirectGPTConfig returned error: %v", err)
	}
	if cfg.TextAPIKey != "sk-from-settings" || cfg.ImageAPIKey != "sk-from-settings" {
		t.Fatalf("cfg keys = text:%q image:%q, want settings key for both", cfg.TextAPIKey, cfg.ImageAPIKey)
	}
}
