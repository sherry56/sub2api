CREATE TABLE IF NOT EXISTS research_drawing_image2_records (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id TEXT NOT NULL PRIMARY KEY,
    model TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_research_drawing_image2_records_user_created_at
    ON research_drawing_image2_records(user_id, created_at DESC);
