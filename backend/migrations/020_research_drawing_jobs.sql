-- Durable schema target for research drawing task state and billing idempotency.
-- The current handler still uses an in-memory compatibility map; wire this table
-- through a repository before enabling multi-instance or restart-safe refunds.
CREATE TABLE IF NOT EXISTS research_drawing_jobs (
    job_id VARCHAR(128) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    prompt TEXT NOT NULL DEFAULT '',
    request_params JSONB NOT NULL DEFAULT '{}'::jsonb,
    charged_amount NUMERIC(18, 8) NOT NULL DEFAULT 0,
    charged BOOLEAN NOT NULL DEFAULT FALSE,
    refunded BOOLEAN NOT NULL DEFAULT FALSE,
    result_candidates JSONB NOT NULL DEFAULT '[]'::jsonb,
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT research_drawing_jobs_status_check CHECK (status IN ('pending', 'running', 'success', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_research_drawing_jobs_user_id_created_at
    ON research_drawing_jobs (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_research_drawing_jobs_status_updated_at
    ON research_drawing_jobs (status, updated_at DESC);
