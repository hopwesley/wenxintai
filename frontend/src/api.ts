/*
 * Simple wrapper around the Wenxintai HTTP API. Each function performs
 * a fetch call to the corresponding endpoint. The server expects and
 * returns JSON. Adjust error handling as needed.
 */
export interface LoginPayload {
  wechat_id: string
  nickname: string
  avatar_url: string
}

export async function login(payload: LoginPayload) {
  const resp = await fetch('/api/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  })
  if (!resp.ok) {
    throw new Error('登录失败')
  }
  return resp.json()
}

export async function getHobbies(): Promise<string[]> {
  const resp = await fetch('/api/hobbies')
  if (!resp.ok) {
    throw new Error('无法获取爱好列表')
  }
  const data = await resp.json()
  return data.hobbies
}

export interface QuestionsRequest {
  session_id: string
  mode: string
  gender: string
  grade: string
  hobby: string
  api_key?: string
}

export async function getQuestions(payload: QuestionsRequest) {
  const resp = await fetch('/api/questions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  })
  if (!resp.ok) {
    throw new Error('无法获取题目')
  }
  return resp.json()
}

export interface AnswersRequest {
  session_id: string
  mode: string
  riasec_answers: any[]
  asc_answers: any[]
  alpha?: number
  beta?: number
  gamma?: number
}

export async function sendAnswers(payload: AnswersRequest) {
  const resp = await fetch('/api/answers', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  })
  if (!resp.ok) {
    throw new Error('评分失败')
  }
  return resp.json()
}

export interface ReportRequest {
  session_id: string
  mode: string
  api_key?: string
}

export async function getReport(payload: ReportRequest) {
  const resp = await fetch('/api/report', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  })
  if (!resp.ok) {
    throw new Error('生成报告失败')
  }
  return resp.json()
}

export interface VerifyInviteRequest {
  code: string
}

export type InviteFailureReason = 'used' | 'expired' | 'not_found'

export interface VerifyInviteResponse {
  ok: boolean
  reason?: InviteFailureReason
}

export async function verifyInviteCode(payload: VerifyInviteRequest): Promise<VerifyInviteResponse> {
  const resp = await fetch('/api/invites/verify-and-redeem', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  })

  if (!resp.ok) {
    throw new Error('邀请码校验失败')
  }

  return resp.json()
}