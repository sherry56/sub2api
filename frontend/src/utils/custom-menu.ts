export function shouldOpenCustomMenuInTopWindow(baseUrl: string): boolean {
  if (!baseUrl) return false

  try {
    const url = new URL(baseUrl)
    return url.pathname === '/menu/paperbanana-launch.html'
  } catch {
    return false
  }
}
