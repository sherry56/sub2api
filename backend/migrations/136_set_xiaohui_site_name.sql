INSERT INTO settings (key, value)
VALUES ('site_name', 'XiaoHuiAPI')
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW();
