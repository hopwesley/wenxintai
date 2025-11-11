interface RequestOptions {
  method?: string
  body?: unknown
  headers?: Record<string, string>
}

async function parseResponseBody(resp: Response) {
  if (resp.status === 204) {
    return null
  }
  const contentType = resp.headers.get('Content-Type') ?? ''
  if (contentType.includes('application/json')) {
    try {
      return await resp.json()
    } catch (error) {
      console.warn('[api] failed to parse JSON response', error)
      return null
    }
  }
  try {
    return await resp.text()
  } catch (error) {
    console.warn('[api] failed to read text response', error)
    return null
  }
}

async function request<T = any>(path: string, options: RequestOptions = {}): Promise<T> {
  const init: RequestInit = {
    method: options.method ?? 'GET',
    headers: { ...options.headers },
    credentials: 'include'
  }

  if (options.body !== undefined) {
    init.body = JSON.stringify(options.body)
    init.headers = {
      'Content-Type': 'application/json',
      ...init.headers
    }
  }

  const resp = await fetch(path, init)

  if (!resp.ok) {
    const data = await parseResponseBody(resp)
    let message = `请求失败 (${resp.status})`
    let code = ''
    if (data && typeof data === 'object') {
      const maybeObj = data as Record<string, unknown>
      if (typeof maybeObj.error === 'string' && maybeObj.error.trim()) {
        message = maybeObj.error
      }
      if (typeof maybeObj.code === 'string' && maybeObj.code.trim()) {
        code = maybeObj.code
      }
    } else if (typeof data === 'string' && data.trim()) {
      message = data
    }
    const error = new Error(message)
    if (code) {
      error.name = code
    }
    ;(error as Error & { status?: number }).status = resp.status
    throw error
  }

  return (await parseResponseBody(resp)) as T
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
  gender: string
  grade: string
  hobby: string
  api_key?: string
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