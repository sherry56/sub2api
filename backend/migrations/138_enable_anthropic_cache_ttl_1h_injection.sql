INSERT INTO settings (key, value)
VALUES ('enable_anthropic_cache_ttl_1h_injection', 'true')
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW();
