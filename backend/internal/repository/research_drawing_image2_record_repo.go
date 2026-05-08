package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/model"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type researchDrawingImage2RecordRepository struct {
	db *sql.DB
}

func NewResearchDrawingImage2RecordRepository(db *sql.DB) service.ResearchDrawingImage2RecordRepository {
	return &researchDrawingImage2RecordRepository{db: db}
}

func (r *researchDrawingImage2RecordRepository) CreateResearchDrawingImage2Record(ctx context.Context, record model.ResearchDrawingImage2Record) error {
	if r == nil || r.db == nil || record.UserID <= 0 || record.JobID == "" {
		return nil
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO research_drawing_image2_records (user_id, job_id, model, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (job_id) DO NOTHING
	`, record.UserID, record.JobID, record.Model, record.CreatedAt)
	return err
}

func (r *researchDrawingImage2RecordRepository) ListResearchDrawingImage2Records(ctx context.Context, userID int64, limit int) ([]model.ResearchDrawingImage2Record, error) {
	if r == nil || r.db == nil || userID <= 0 {
		return []model.ResearchDrawingImage2Record{}, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, job_id, model, created_at
		FROM research_drawing_image2_records
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	records := make([]model.ResearchDrawingImage2Record, 0)
	for rows.Next() {
		var record model.ResearchDrawingImage2Record
		if err := rows.Scan(&record.UserID, &record.JobID, &record.Model, &record.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func (r *researchDrawingImage2RecordRepository) GetResearchDrawingImage2Record(ctx context.Context, userID int64, jobID string) (*model.ResearchDrawingImage2Record, error) {
	if r == nil || r.db == nil || userID <= 0 || jobID == "" {
		return nil, nil
	}
	var record model.ResearchDrawingImage2Record
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, job_id, model, created_at
		FROM research_drawing_image2_records
		WHERE user_id = $1 AND job_id = $2
		LIMIT 1
	`, userID, jobID).Scan(&record.UserID, &record.JobID, &record.Model, &record.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}
