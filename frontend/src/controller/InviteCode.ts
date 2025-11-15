import {apiRequest} from "@/api";
const api_verify='/api/invites/verify';
const api_redeem='/api/invites/redeem';

export interface VerifyInviteResponse {
    ok: boolean
    reason: string
    has_record: boolean
    test_id?: number | null
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
