interface RequestOptions {
  method?: string
  body?: unknown
  headers?: Record<string, string>
}

type ApiError = Error & { code?: string; body?: any }

async function request<T = any>(path: string, options: RequestOptions = {}): Promise<T> {
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

    let body: any = null
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
  return request('/api/login', { method: 'POST', body: payload })
}

export async function getHobbies(): Promise<string[]> {
  const data = await request<{ hobbies: string[] }>('/api/hobbies')
  return data?.hobbies ?? []
}

export interface QuestionsRequest {
  session_id?: string
  mode: string
  grade: string
  hobby: string
}

export async function getQuestions(payload: QuestionsRequest) {
  return request('/api/questions', { method: 'POST', body: payload })
}

export interface AnswersRequest {
  session_id?: string
  mode: string
  riasec_answers: any[]
  asc_answers: any[]
  alpha?: number
  beta?: number
  gamma?: number
}

export async function sendAnswers(payload: AnswersRequest) {
  return request('/api/answers', { method: 'POST', body: payload })
}

export interface SubmitTestSessionRequest {
  sessionId: string
  variant: 'basic' | 'pro' | 'campus'
  age: number
  mode: string
  hobby: string
  riasec_answers: Record<string, number>
  asc_answers: Record<string, number>
}

export async function submitTestSession(payload: SubmitTestSessionRequest) {
  return request('/api/test/submit', { method: 'POST', body: payload })
}

export interface ReportRequest {
  session_id?: string
  mode: string
  api_key?: string
}

export async function getReport(payload: ReportRequest) {
  return request('/api/report', { method: 'POST', body: payload })
}

export interface VerifyInviteResponse {
  session_id: string
  status: string
  reserved_until?: string
}

export async function verifyInvite(code: string, sessionId?: string): Promise<VerifyInviteResponse> {
  return request<VerifyInviteResponse>('/api/invites/verify', {
    method: 'POST',
    body: { code, session_id: sessionId }
  })
}

export async function redeemInvite(sessionId?: string) {
  return request('/api/invites/redeem', {
    method: 'POST',
    body: sessionId ? { session_id: sessionId } : {}
  })
}