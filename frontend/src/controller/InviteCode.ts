import {apiRequest} from "@/api";
const api_verify='/api/invites/verify';
const api_redeem='/api/invites/redeem';
export interface VerifyInviteResponse {
    session_id: string
    status: string
    reserved_until?: string
}

export async function verifyInvite(code: string, sessionId?: string): Promise<VerifyInviteResponse> {
    return apiRequest<VerifyInviteResponse>(api_verify, {
        method: 'POST',
        body: { code, session_id: sessionId }
    })
}

export async function redeemInvite(sessionId?: string) {
    return apiRequest(api_redeem, {
        method: 'POST',
        body: sessionId ? { session_id: sessionId } : {}
    })
}
