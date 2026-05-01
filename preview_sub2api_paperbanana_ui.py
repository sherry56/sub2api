from __future__ import annotations

from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


CSS = """
:root {
  --bg: #f8fafc;
  --panel: #ffffff;
  --line: #e5e7eb;
  --text: #111827;
  --muted: #667085;
  --primary: #2563eb;
  --primary-soft: #eff6ff;
  --primary-line: #bfdbfe;
  --shadow: 0 10px 28px rgba(15, 23, 42, 0.08);
}
* { box-sizing: border-box; }
body {
  margin: 0;
  min-height: 100vh;
  background: var(--bg);
  color: var(--text);
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}
.app { display: flex; min-height: 100vh; }
.sidebar {
  width: 256px;
  flex: 0 0 256px;
  background: #fff;
  border-right: 1px solid var(--line);
  display: flex;
  flex-direction: column;
}
.brand {
  height: 64px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 22px;
  border-bottom: 1px solid #f1f5f9;
}
.logo {
  width: 36px;
  height: 36px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  color: #fff;
  font-weight: 800;
  background: linear-gradient(135deg, #2563eb, #22c55e);
  box-shadow: 0 8px 18px rgba(37, 99, 235, .24);
}
.brand-title { font-weight: 800; letter-spacing: .01em; }
.version { font-size: 11px; color: var(--muted); margin-top: 2px; }
.nav { padding: 16px 12px; overflow: auto; }
.section-title {
  padding: 12px 12px 8px;
  color: #9ca3af;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: .08em;
  text-transform: uppercase;
}
.nav a {
  text-decoration: none;
  color: #4b5563;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-radius: 14px;
  font-size: 14px;
  font-weight: 650;
  margin-bottom: 4px;
}
.nav a.active {
  color: var(--primary);
  background: var(--primary-soft);
}
.ico { width: 20px; text-align: center; opacity: .9; }
.main { flex: 1; min-width: 0; }
.topbar {
  height: 64px;
  background: rgba(255,255,255,.82);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--line);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;
}
.crumb { color: var(--muted); font-size: 13px; }
.user-pill {
  padding: 8px 12px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: #fff;
  color: #475467;
  font-size: 13px;
}
.content {
  max-width: 1180px;
  margin: 0 auto;
  padding: 28px;
}
.card {
  background: var(--panel);
  border: 1px solid #eef2f7;
  border-radius: 18px;
  box-shadow: var(--shadow);
}
.hero {
  border-color: var(--primary-line);
  background: linear-gradient(180deg, #eff6ff, #ffffff);
  padding: 20px;
}
h1, h2, h3, h4, p { margin: 0; }
.hero h3 { color: #1e40af; font-size: 18px; }
.hero p { color: #1d4ed8; font-size: 14px; margin-top: 6px; }
.hero .price { color: #1d4ed8; opacity: .82; font-size: 12px; margin-top: 10px; }
.grid { display: grid; gap: 18px; }
.form { padding: 24px; margin-top: 22px; }
.section { margin-bottom: 24px; }
.section h4 { font-size: 16px; margin-bottom: 14px; }
.fields { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 16px; }
.fields.two { grid-template-columns: repeat(2, minmax(0, 1fr)); }
label { display: block; color: #374151; font-size: 13px; font-weight: 650; }
select, input {
  width: 100%;
  margin-top: 7px;
  border: 1px solid #dbe3ef;
  background: #fff;
  border-radius: 13px;
  height: 42px;
  padding: 0 13px;
  color: #111827;
  font-size: 14px;
  outline: none;
}
select:focus, input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, .14);
}
.actions { display: flex; gap: 12px; flex-wrap: wrap; }
.btn {
  border: 0;
  border-radius: 13px;
  padding: 10px 16px;
  font-weight: 750;
  font-size: 14px;
  cursor: default;
}
.primary { color: #fff; background: linear-gradient(90deg, #3b82f6, #2563eb); box-shadow: 0 8px 20px rgba(37,99,235,.22); }
.secondary { background: #fff; color: #374151; border: 1px solid #dbe3ef; }
.info-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); margin-top: 22px; }
.info { padding: 22px; }
.info h4 { font-size: 16px; margin-bottom: 12px; }
.info p { color: var(--muted); font-size: 14px; line-height: 1.65; margin-top: 8px; }
.admin-only {
  margin-top: 22px;
  padding: 36px;
  text-align: center;
}
.admin-only h3 { font-size: 18px; }
.admin-only p { color: var(--muted); font-size: 14px; margin-top: 8px; }
.toggle { display:flex; gap:8px; }
.toggle a {
  text-decoration:none;
  border-radius: 999px;
  padding: 7px 11px;
  font-size: 13px;
  border:1px solid var(--line);
  color:#475467;
  background:#fff;
}
.toggle a.on { color:#fff; background:var(--primary); border-color:var(--primary); }
@media (max-width: 900px) {
  .sidebar { width: 74px; flex-basis: 74px; }
  .brand-title, .version, .nav span, .section-title { display: none; }
  .brand { justify-content:center; padding:0; }
  .nav a { justify-content:center; padding:12px; }
  .fields, .fields.two, .info-grid { grid-template-columns: 1fr; }
  .content { padding: 18px; }
}
"""


def page(admin: bool) -> str:
    active_admin = "on" if admin else ""
    active_user = "" if admin else "on"
    form = """
      <form class="card form">
        <section class="section">
          <h4>流程参数</h4>
          <div class="fields">
            <label>模式
              <select><option>demo_planner_critic（无 Stylist）</option><option>demo_full（含 Stylist）</option></select>
            </label>
            <label>检索
              <select><option>auto</option><option>manual</option><option>random</option><option>none</option></select>
            </label>
            <label>宽高比
              <select><option>16:9</option><option>21:9</option><option>3:2</option></select>
            </label>
          </div>
        </section>
        <section class="section">
          <h4>生成参数</h4>
          <div class="fields">
            <label>候选数<input type="number" value="1" /></label>
            <label>Critic 轮数<input type="number" value="2" /></label>
            <label>精修分辨率上限<select><option>2K</option><option>4K</option></select></label>
          </div>
        </section>
        <section class="section">
          <h4>模型参数</h4>
          <div class="fields two">
            <label>主模型<input value="openrouter/google/gemini-2.5-flash" /></label>
            <label>生图模型<input value="openrouter/google/gemini-2.5-flash-image-preview" /></label>
          </div>
        </section>
        <div class="actions">
          <button class="btn primary" type="button">保存</button>
          <button class="btn secondary" type="button">恢复默认参数</button>
        </div>
      </form>
    """
    admin_only = """
      <div class="card admin-only">
        <h3>需要管理员权限</h3>
        <p>当前页面为系统级配置区域，请联系管理员修改参数。</p>
      </div>
    """
    return f"""<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Sub2API 科研绘图预览</title>
  <style>{CSS}</style>
</head>
<body>
  <div class="app">
    <aside class="sidebar">
      <div class="brand">
        <div class="logo">S</div>
        <div><div class="brand-title">Sub2API</div><div class="version">v1.0.0</div></div>
      </div>
      <nav class="nav">
        <div class="section-title">我的账户</div>
        <a><span class="ico">⌂</span><span>首页</span></a>
        <a><span class="ico">⌁</span><span>API Keys</span></a>
        <a><span class="ico">◌</span><span>可用渠道</span></a>
        <a class="active"><span class="ico">▧</span><span>科研绘图</span></a>
        <a><span class="ico">◍</span><span>我的订阅</span></a>
        <a><span class="ico">◎</span><span>个人资料</span></a>
      </nav>
    </aside>
    <main class="main">
      <header class="topbar">
        <div>
          <strong>科研绘图</strong>
          <div class="crumb">配置科研绘图内置生成流程的参数设置。</div>
        </div>
        <div class="toggle">
          <a class="{active_admin}" href="/admin">管理员视图</a>
          <a class="{active_user}" href="/user">用户视图</a>
        </div>
      </header>
      <div class="content">
        <section class="card hero">
          <h3>一键生成SCI配色的论文框架图和统计图</h3>
          <p>配置科研绘图内置生成流程的参数设置。</p>
          <div class="price">当前单次价格：3.00 元/次。</div>
        </section>
        {form if admin else admin_only}
        <section class="grid info-grid">
          <div class="card info">
            <h4>典型场景与版式参考</h4>
            <p>适用于论文方法框架图（方法、流程、架构）的快速草图生成。</p>
            <p>适用于实验流程图或算法步骤图，结构更清晰。</p>
            <p>适用于 SCI 风格统计图初稿，保证配色语言统一。</p>
          </div>
          <div class="card info">
            <h4>使用建议</h4>
            <p>1) 建议先用默认参数，再逐步提高候选数和 Critic 轮数做精修。</p>
            <p>2) 按任务类型区分主模型与生图模型，通常能提升稳定性与质量。</p>
            <p>3) 修改后先用小样本验证，再批量应用到正式任务。</p>
          </div>
        </section>
      </div>
    </main>
  </div>
</body>
</html>"""


class Handler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:
        body = page(admin=self.path != "/user").encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format: str, *args: object) -> None:
        return


if __name__ == "__main__":
    server = ThreadingHTTPServer(("127.0.0.1", 3010), Handler)
    print("Preview: http://127.0.0.1:3010/admin", flush=True)
    server.serve_forever()
