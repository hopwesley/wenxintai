import {apiRequest} from '@/api'
import {useTestSession} from '@/controller/testSession'
import type {PlanKey} from '@/controller/common'
import {ref} from "vue";

export interface NativeCreateOrderResponse {
    order_id: string
    code_url: string
    amount?: number
    description?: string
}

// 创建 native 订单：向后端要 code_url
export async function createNativeOrder(planKey: PlanKey): Promise<NativeCreateOrderResponse> {
    const {state} = useTestSession()

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


export function useNativePayment() {

    const paying = ref(false)              // 点击“微信支付”后，按钮 loading
    const payOrder = ref<NativeCreateOrderResponse | null>(null)  // 当前订单信息（包括 code_url）
    const payPollingTimer = ref<number | null>(null)              // 轮询定时器 id
    const paySucceeded = ref(false)

    async function handleWeChatPay(currentPlan: PlanKey) {
        if (!currentPlan) {
            return
        }

        if (paying.value) {
            return
        }

        try {
            paying.value = true
            paySucceeded.value = false

            // 1. 调后端创建 native 订单
            const order = await createNativeOrder(currentPlan)

            payOrder.value = order

            // 2. 不要立刻关弹窗，而是在弹窗里展示二维码
            //   -> InviteCodeModal 需要根据是否有 payOrder 来切换 UI（输入邀请码 vs 显示二维码）
            //   -> 这里只负责把 payOrder 设置好

            // 3. 开始轮询订单状态
            startPayPolling(order.order_id)
        } catch (err: any) {
            console.error('[Pay] handleWeChatPay failed', err)
        } finally {
            paying.value = false
        }
    }

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
                    // 例如：router.push(...) 或者调用 handlePaySuccess()
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
        handleWeChatPay,
    }
}

