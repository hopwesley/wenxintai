interface RequestOptions {
  method?: string
  body?: unknown
}

interface ApiError extends Error {
  code?: string
  status?: number
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const init: RequestInit = {
    method: options.method ?? 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }
  if (options.body !== undefined) {
    init.body = JSON.stringify(options.body)
  }

  const resp = await fetch(path, init)
  const contentType = resp.headers.get('Content-Type') ?? ''
  const isJSON = contentType.includes('application/json')
  const payload = isJSON ? await resp.json().catch(() => null) : await resp.text().catch(() => null)

  if (!resp.ok) {
    const error: ApiError = new Error('请求失败')
    error.status = resp.status
    if (payload && typeof payload === 'object' && 'message' in payload) {
      error.message = String((payload as any).message)
    } else if (typeof payload === 'string' && payload.trim()) {
      error.message = payload
    }
    if (payload && typeof payload === 'object' && 'code' in payload) {
      error.code = String((payload as any).code)
    }
    throw error
  }

  return payload as T
}

export interface CreateAssessmentRequest {
  mode: string
  invite_code?: string
  wechat_openid?: string
}

export interface CreateAssessmentResponse {
  assessment_id: string
  question_set_id: string
  stage: 'S1'
  questions: any
}

export function createAssessment(body: CreateAssessmentRequest) {
  return request<CreateAssessmentResponse>('/api/assessments', {
    method: 'POST',
    body
  })
}

export interface SubmitAnswersResponseStage1 {
  next_question_set_id: string
  stage: 'S2'
  questions: any
}

export interface SubmitAnswersResponseStage2 {
  assessment_id: string
  report_id: string
}

export type SubmitAnswersResponse = SubmitAnswersResponseStage1 | SubmitAnswersResponseStage2

export function submitAnswers(questionSetId: string, answers: any[]): Promise<SubmitAnswersResponse> {
  return request<SubmitAnswersResponse>(`/api/question_sets/${encodeURIComponent(questionSetId)}/answers`, {
    method: 'POST',
    body: { answers }
  })
}

export interface ReportResponse {
  report_id: string
  summary?: string
  full: any
}

export function getReport(assessmentId: string) {
  return request<ReportResponse>(`/api/assessments/${encodeURIComponent(assessmentId)}/report`)
}

export interface ProgressResponse {
  status: number
  label: string
}

export function getProgress(assessmentId: string) {
  return request<ProgressResponse>(`/api/assessments/${encodeURIComponent(assessmentId)}/progress`)
}
