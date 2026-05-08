package service

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/model"
)

type ResearchDrawingImage2RecordRepository interface {
	CreateResearchDrawingImage2Record(ctx context.Context, record model.ResearchDrawingImage2Record) error
	ListResearchDrawingImage2Records(ctx context.Context, userID int64, limit int) ([]model.ResearchDrawingImage2Record, error)
	GetResearchDrawingImage2Record(ctx context.Context, userID int64, jobID string) (*model.ResearchDrawingImage2Record, error)
}

type ResearchDrawingImage2RecordService struct {
	repo ResearchDrawingImage2RecordRepository
}

func NewResearchDrawingImage2RecordService(repo ResearchDrawingImage2RecordRepository) *ResearchDrawingImage2RecordService {
	return &ResearchDrawingImage2RecordService{repo: repo}
}

func (s *ResearchDrawingImage2RecordService) Create(ctx context.Context, record model.ResearchDrawingImage2Record) error {
	if s == nil || s.repo == nil {
		return nil
	}
	return s.repo.CreateResearchDrawingImage2Record(ctx, record)
}

func (s *ResearchDrawingImage2RecordService) ListByUser(ctx context.Context, userID int64, limit int) ([]model.ResearchDrawingImage2Record, error) {
	if s == nil || s.repo == nil || userID <= 0 {
		return []model.ResearchDrawingImage2Record{}, nil
	}
	return s.repo.ListResearchDrawingImage2Records(ctx, userID, limit)
}

func (s *ResearchDrawingImage2RecordService) GetByUserJob(ctx context.Context, userID int64, jobID string) (*model.ResearchDrawingImage2Record, error) {
	if s == nil || s.repo == nil || userID <= 0 || jobID == "" {
		return nil, nil
	}
	return s.repo.GetResearchDrawingImage2Record(ctx, userID, jobID)
}
