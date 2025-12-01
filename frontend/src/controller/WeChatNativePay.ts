// src/controller/WeChatNativePay.ts
import { apiRequest } from '@/api'
import { useTestSession } from '@/controller/testSession'
import type { PlanKey } from '@/controller/common'

export interface NativeCreateOrderResponse {
    order_id: string
    code_url: string
    amount?: number
    description?: string
}

// 创建 native 订单：向后端要 code_url
export async function createNativeOrder(planKey: PlanKey): Promise<NativeCreateOrderResponse> {
    const { state } = useTestSession()

    const body = {
        business_type: state.businessType,
        plan_key: planKey,
    }

    return apiRequest<NativeCreateOrderResponse>('/api/pay/wechat/native/create', {
        method: 'POST',
        body,
    })
}

// 轮询支付状态的结果结构（根据你后端约定来定）
export interface QueryOrderStatusResponse {
    paid: boolean
    // 也可以带状态码：pending/paid/failed
}

export async function queryOrderStatus(orderId: string): Promise<QueryOrderStatusResponse> {
    return apiRequest<QueryOrderStatusResponse>(`/api/pay/wechat/order-status?order_id=${encodeURIComponent(orderId)}`)
}
