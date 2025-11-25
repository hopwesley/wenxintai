interface RequestOptions {
  method?: string
  body?: unknown
  headers?: Record<string, string>
}

type ApiError = Error & { code?: string; body?: any }

export async function apiRequest<T = any>(path: string, options: RequestOptions = {}): Promise<T> {
    const init: RequestInit = {
        method: options.method ?? 'GET',
        headers: { ...options.headers },
        credentials: 'include',
    }

    if (options.body !== undefined) {
        init.body = JSON.stringify(options.body)
        init.headers = { 'Content-Type': 'application/json', ...init.headers }
    }

    const resp = await fetch(path, init)

    if (resp.ok) {
        try {
            return (await resp.json()) as T
        } catch {
            return undefined as T
        }
    }

    let body: any
    try {
        body = await resp.json()
    } catch {
        body = null
    }

    const code = (body?.code && String(body.code)) || ''
    const message = (body?.message && String(body.message)) || '请求失败，请稍后重试'

    // 只对外暴露 code + message（UI 显示 message，必要时可读 err.code）
    const err = new Error(message) as ApiError
    if (code) {
        err.name = code
        err.code = code
    }
    err.body = body // 便于上层需要时做额外处理/埋点

    console.debug('[API ERROR]', resp.status, path, body)
    throw err
}

export interface LoginPayload {
  wechat_id: string
  nickname: string
  avatar_url: string
}

export async function login(payload: LoginPayload) {
  return apiRequest('/api/login', { method: 'POST', body: payload })
}

export async function getHobbies(): Promise<string[]> {
  const data = await apiRequest<{ hobbies: string[] }>('/api/hobbies')
  return data?.hobbies ?? []
}

export interface ReportRequest {
  session_id?: string
  mode: string
  api_key?: string
}

export async function getReport(payload: ReportRequest) {
  return apiRequest('/api/report', { method: 'POST', body: payload })
}
