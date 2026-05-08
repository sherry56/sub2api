package model

import "time"

type ResearchDrawingImage2Record struct {
	UserID    int64     `json:"user_id"`
	JobID     string    `json:"job_id"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}
