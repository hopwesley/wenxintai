import {apiRequest} from "@/api";
import {useTestSession} from "@/controller/testSession";
import {ref} from "vue";

export async function verifyInviteWithMessage(rawCode: string){
}


export interface NativeCreateOrderResponse {
    order_id: string
    code_url: string
    amount: number
    description: string
}


// 轮询支付状态的结果结构（根据你后端约定来定）
export interface QueryOrderStatusResponse {
    paid: boolean
}

export async function queryOrderStatus(orderId: string): Promise<QueryOrderStatusResponse> {
    return apiRequest<QueryOrderStatusResponse>(`/api/pay/wechat/order-status?order_id=${encodeURIComponent(orderId)}`)
}


export function useNativePayment() {
    const paying = ref(false)              // 点击“微信支付”后，按钮 loading
    const payOrder = ref<NativeCreateOrderResponse | null>(null)  // 当前订单信息（包括 code_url）
    const payPollingTimer = ref<number | null>(null)              // 轮询定时器 id
    const paySucceeded = ref(false)

    function startPayPolling(orderId: string) {
        stopPayPolling() // 防止重复

        // 每 2 秒查一次支付状态，视情况调节
        payPollingTimer.value = window.setInterval(async () => {
            try {
                const status = await queryOrderStatus(orderId)
                if (status.paid) {
                    paySucceeded.value = true
                    stopPayPolling()
                    // TODO: 在这里触发你原来进入测试的逻辑
                }
            } catch (err) {
                console.error('[Pay] queryOrderStatus error', err)
                // 轮询失败可以先不打断，下次继续查
            }
        }, 2000)
    }

    function stopPayPolling() {
        if (payPollingTimer.value !== null) {
            window.clearInterval(payPollingTimer.value)
            payPollingTimer.value = null
        }
    }

    return {
        payOrder,
        paying,
    }
}