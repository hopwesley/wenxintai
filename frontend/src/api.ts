interface RequestOptions {
    method?: string
    body?: unknown
    headers?: Record<string, string>
}

export const API_PATHS = {
    HEALTH: '/api/health',
    LOAD_HOBBIES: '/api/hobbies',
    LOAD_PRODUCTS: '/api/products',
    LOAD_CUR_PRODUCT: '/api/prepare_pay',

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
    INVITE_PAYMENT: '/api/pay/use_invite',
    WECHAT_CREATE_NATIVE_ORDER: '/api/pay/wechat/order_create',
    WECHAT_NATIVE_ORDER_STATUS: '/api/pay/wechat/order_status',
} as const

interface ApiErr {
    code: string;
    message: string;
    err?: any;
}

export async function apiRequest<T = any>(
    path: string,
    options: RequestOptions = {}
): Promise<T> {
    const init: RequestInit = {
        method: options.method ?? 'GET',
        headers: {...options.headers},
        credentials: 'include',
    };

    if (options.body !== undefined) {
        init.body = JSON.stringify(options.body);
        init.headers = {'Content-Type': 'application/json', ...init.headers};
    }

    let resp: Response;
    try {
        resp = await fetch(path, init);
        if (resp.ok) {
            const text = await resp.text();
            return text ? JSON.parse(text) as T : undefined as T;
        }
    } catch (networkErr) {
        console.error('Network error:', networkErr);
        const error = new Error("网络异常，请检查网络后重试:" + (networkErr as Error).message) as Error & ApiErr;
        error.code = 'NETWORK_ERROR';
        error.err = networkErr;
        throw error;
    }

    const contentType = resp.headers.get('content-type') || '';
    let body: any;
    try {
        body = contentType.includes('json') ? await resp.json() : await resp.text();
    } catch (e) {
        body = null;
    }

    if (isApiErr(body)) {
        throw body;
    }

    const error = new Error("系统错误，请稍后重试:"+body) as Error & ApiErr;
    error.code = 'SYSTEM_ERROR';
    error.err = {status: resp.status, raw: body};
    throw error;
}

export function isApiErr(error: unknown): error is (Error & ApiErr) {
    return error !== null &&
        typeof error === 'object' &&
        'code' in error &&
        'message' in error &&
        typeof (error as any).code === 'string' &&
        typeof (error as any).message === 'string';
}



