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

    const code = (body?.code && String(body?.code)) || ''
    const message = (body?.message && String(body?.message)) || '请求失败，请稍后重试'

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

export const API_PATHS = {
    HEALTH: '/api/health',
    LOAD_HOBBIES: '/api/hobbies',

    TEST_FLOW: '/api/test_flow',

    TEST_BASIC_INFO: '/api/tests/basic_info',

    SSE_QUESTION_SUB: '/api/sub/question/',
    SSE_REPORT_SUB: '/api/sub/report/',
    SUBMIT_TEST: '/api/test_submit',
    GENERATE_REPORT: '/api/generate_report',
    FINISH_REPORT: '/api/finish_report',

    WECHAT_SIGN_IN: '/api/auth/wx/status',
    WECHAT_SIGN_IN_CALLBACK: '/api/wechat_signin',
    WECHAT_LOGOUT: '/api/auth/logout',
    WECHAT_UPDATE_PROFILE: '/api/user/update_profile',
    WECHAT_MY_PROFILE: '/api/auth/profile',

    WECHAT_PAYMENT: '/api/pay/',
    WECHAT_CREATE_NATIVE_ORDER: '/api/pay/wechat/native/create',
    WECHAT_NATIVE_ORDER_STATUS: '/api/pay/wechat/order-status',
} as const
