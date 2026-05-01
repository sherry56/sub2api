from __future__ import annotations

import json
import time
import base64
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlparse


USER = {
    "id": 1,
    "username": "admin",
    "email": "admin@sub2api.local",
    "role": "admin",
    "balance": 9999,
    "concurrency": 5,
    "status": "active",
    "allowed_groups": None,
    "balance_notify_enabled": False,
    "balance_notify_threshold": None,
    "balance_notify_extra_emails": [],
    "created_at": "2026-04-29T00:00:00+08:00",
    "updated_at": "2026-04-29T00:00:00+08:00",
    "run_mode": "standard",
}


PUBLIC_SETTINGS = {
    "registration_enabled": True,
    "email_verify_enabled": False,
    "force_email_on_third_party_signup": False,
    "registration_email_suffix_whitelist": [],
    "promo_code_enabled": True,
    "password_reset_enabled": False,
    "invitation_code_enabled": False,
    "turnstile_enabled": False,
    "turnstile_site_key": "",
    "site_name": "Sub2API",
    "site_logo": "/logo.png",
    "site_subtitle": "",
    "api_base_url": "http://127.0.0.1:8080",
    "contact_info": "",
    "doc_url": "",
    "home_content": "",
    "hide_ccs_import_button": False,
    "payment_enabled": False,
    "table_default_page_size": 20,
    "table_page_size_options": [10, 20, 50, 100],
    "custom_menu_items": [],
    "custom_endpoints": [],
    "linuxdo_oauth_enabled": False,
    "wechat_oauth_enabled": False,
    "wechat_oauth_open_enabled": False,
    "wechat_oauth_mp_enabled": False,
    "wechat_oauth_mobile_enabled": False,
    "oidc_oauth_enabled": False,
    "oidc_oauth_provider_name": "OIDC",
    "backend_mode_enabled": False,
    "version": "preview",
    "balance_low_notify_enabled": False,
    "account_quota_notify_enabled": False,
    "balance_low_notify_threshold": 0,
    "channel_monitor_enabled": True,
    "channel_monitor_default_interval_seconds": 60,
    "available_channels_enabled": True,
    "affiliate_enabled": False,
}


SETTINGS = {
    **PUBLIC_SETTINGS,
    "research_drawing_exp_mode": "demo_planner_critic",
    "research_drawing_retrieval_setting": "auto",
    "research_drawing_num_candidates": 1,
    "research_drawing_aspect_ratio": "16:9",
    "research_drawing_max_critic_rounds": 2,
    "research_drawing_main_model_name": "openrouter/google/gemini-2.5-flash",
    "research_drawing_image_gen_model_name": "openrouter/google/gemini-2.5-flash-image-preview",
    "research_drawing_max_refine_resolution": "2K",
}

JOBS: dict[str, dict[str, object]] = {}
MOCK_PNG = base64.b64decode(
    "iVBORw0KGgoAAAANSUhEUgAAAMgAAAB4CAIAAAD2HxkiAAABRklEQVR4nO3awQmAIBRAwRj775xsdAYhwfKCJy+/IHjJzByAnp7vAfjHYQAGYACWxfMCjLk5T5fMXruL5wUYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAEYgAGYDWAF6kBbXvoKuoIAAAAASUVORK5CYII="
)


def ok(data: object) -> bytes:
    return json.dumps({"code": 0, "message": "ok", "data": data}, ensure_ascii=False).encode("utf-8")


class Handler(BaseHTTPRequestHandler):
    def _send(self, body: bytes, status: int = 200) -> None:
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Access-Control-Allow-Origin", self.headers.get("Origin", "*"))
        self.send_header("Access-Control-Allow-Credentials", "true")
        self.send_header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept-Language")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _send_raw(self, body: bytes, content_type: str, status: int = 200) -> None:
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Access-Control-Allow-Origin", self.headers.get("Origin", "*"))
        self.send_header("Access-Control-Allow-Credentials", "true")
        self.send_header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept-Language")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_OPTIONS(self) -> None:
        self._send(b"{}")

    def do_GET(self) -> None:
        path = urlparse(self.path).path
        if path == "/health":
            self._send(ok({"status": "ok"}))
            return
        if path == "/api/v1/settings/public":
            self._send(ok(PUBLIC_SETTINGS))
            return
        if path == "/api/v1/auth/me":
            self._send(ok(USER))
            return
        if path == "/api/v1/admin/settings":
            self._send(ok(SETTINGS))
            return
        if path.startswith("/api/v1/research-drawing/jobs/") and "/images/" in path:
            self._send_raw(MOCK_PNG, "image/png")
            return
        if path.startswith("/api/v1/research-drawing/jobs/"):
            job_id = path.rsplit("/", 1)[-1]
            job = JOBS.get(job_id)
            if not job:
                self._send(json.dumps({"code": 404, "message": "job not found"}).encode("utf-8"), 404)
                return
            elapsed = int(time.time() - float(job["started_at"]))
            status = "done" if elapsed >= 6 else "running"
            self._send(ok({
                "ok": True,
                "job_id": job_id,
                "status": status,
                "elapsed_sec": elapsed,
                "candidate_count": 1 if status == "done" else 0,
                "candidate_ids": [0] if status == "done" else [],
                "images": [{"candidate_id": 0}] if status == "done" else [],
            }))
            return
        self._send(json.dumps({"code": 404, "message": f"mock route not found: {path}"}).encode("utf-8"), 404)

    def do_POST(self) -> None:
        path = urlparse(self.path).path
        if path == "/api/v1/auth/login":
            self._send(ok({
                "access_token": "mock-access-token",
                "refresh_token": "mock-refresh-token",
                "expires_in": 86400,
                "token_type": "Bearer",
                "user": USER,
            }))
            return
        if path == "/api/v1/auth/refresh":
            self._send(ok({
                "access_token": "mock-access-token",
                "refresh_token": "mock-refresh-token",
                "expires_in": 86400,
            }))
            return
        if path == "/api/v1/auth/logout":
            self._send(ok({"message": "logged out"}))
            return
        if path == "/api/v1/research-drawing/generate":
            job_id = f"mock-{int(time.time())}"
            JOBS[job_id] = {"started_at": time.time()}
            self._send(ok({
                "job_id": job_id,
                "status": "running",
                "paperbanana_url": "http://127.0.0.1:8000",
                "paperbanana_user": "s2a_1",
                "charge": 2.99,
                "quota_need": 3,
            }), 202)
            return
        self._send(json.dumps({"code": 404, "message": f"mock route not found: {path}"}).encode("utf-8"), 404)

    def do_PUT(self) -> None:
        path = urlparse(self.path).path
        if path == "/api/v1/admin/settings":
            length = int(self.headers.get("Content-Length", "0") or "0")
            if length:
                try:
                    payload = json.loads(self.rfile.read(length).decode("utf-8"))
                    if isinstance(payload, dict):
                        SETTINGS.update(payload)
                except Exception:
                    pass
            self._send(ok(SETTINGS))
            return
        self._send(json.dumps({"code": 404, "message": f"mock route not found: {path}"}).encode("utf-8"), 404)

    def log_message(self, format: str, *args: object) -> None:
        return


if __name__ == "__main__":
    server = ThreadingHTTPServer(("127.0.0.1", 8080), Handler)
    print("Mock Sub2API backend: http://127.0.0.1:8080", flush=True)
    server.serve_forever()
