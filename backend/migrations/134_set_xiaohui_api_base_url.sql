INSERT INTO settings (key, value)
VALUES
    ('api_base_url', 'https://liuloys.top/v1'),
    ('email_verify_enabled', 'true'),
    ('smtp_host', 'smtp.qq.com'),
    ('smtp_port', '587'),
    ('smtp_username', 'liuloys@qq.com'),
    ('smtp_from', 'liuloys@qq.com'),
    ('smtp_from_name', 'XiaoHuiAPI'),
    ('smtp_use_tls', 'true')
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW();
