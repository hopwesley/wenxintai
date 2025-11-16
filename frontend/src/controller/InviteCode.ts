import {apiRequest} from "@/api";
const api_verify='/api/invites/verify';
const api_redeem='/api/invites/redeem';

export interface VerifyInviteResponse {
    ok: boolean
    reason: string
    public_id?: string | null
}

export async function verifyInvite(code: string): Promise<VerifyInviteResponse> {
    return apiRequest<VerifyInviteResponse>(api_verify, {
        method: 'POST',
        body: { invite_code:code }
    })
}

export async function redeemInvite(sessionId?: string) {
    return apiRequest(api_redeem, {
        method: 'POST',
        body: sessionId ? { session_id: sessionId } : {}
    })
}

// 给组件用的包装结果：同时带是否成功、响应结果、错误提示文案
export interface VerifyInviteResult {
    ok: boolean
    response?: VerifyInviteResponse
    errorMessage?: string
}

/**
 * 统一处理：
 * - 去掉前后空格
 * - 邀请码为空
 * - 后端业务错误（res.ok === false）
 * - 网络异常 / 其它异常
 *
 * 组件只需要关心 ok / errorMessage / response
 */
export async function verifyInviteWithMessage(rawCode: string): Promise<VerifyInviteResult> {
    const code = rawCode.trim()

    if (!code) {
        return {
            ok: false,
            errorMessage: '请输入邀请码',
        }
    }

    try {
        const res = await verifyInvite(code)

        if (!res.ok) {
            return {
                ok: false,
                errorMessage: res.reason || '邀请码无效',
            }
        }

        return {
            ok: true,
            response: res,
        }
    } catch (error) {
        console.error('[InviteCode] verifyInviteWithMessage failed', error)

        if (error instanceof Error) {
            if (error.message === 'Failed to fetch') {
                return {
                    ok: false,
                    errorMessage: '网络异常，请检查网络后重试',
                }
            }
            return {
                ok: false,
                errorMessage: error.message,
            }
        }

        return {
            ok: false,
            errorMessage: '验证失败，请稍后再试',
        }
    }
}
