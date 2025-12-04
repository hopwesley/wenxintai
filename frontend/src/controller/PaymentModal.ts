// src/controller/useNativePayment.ts
import {computed, ref} from 'vue'
import {API_PATHS, apiRequest, isApiErr} from '@/api'
import {router} from "@/controller/router_index";
import {useAlert} from "@/controller/useAlert";
import {useGlobalLoading} from "@/controller/useGlobalLoading";

export interface NativeCreateOrderResponse {
    order_id: string
    code_url: string
    amount: number
    description: string
}

export interface QueryOrderStatusResponse {
    paid: boolean
}

interface UseNativePaymentOptions {
    onSuccess: () => void
    onClose: () => void
    publicID: string
}

export function useNativePayment(opts: UseNativePaymentOptions) {
    const paying = ref(false)
    const payOrder = ref<NativeCreateOrderResponse | null>(null)
    const payPollingTimer = ref<number | null>(null)
    const paySucceeded = ref(false)
    const payLoading = ref(false)

    const code = ref('')
    const inviteLoading = ref(false)
    const errorMessage = ref('')
    const inputRef = ref<HTMLInputElement | null>(null)
    const trimmedCode = computed(() => code.value.trim())
    const {showAlert} = useAlert()
    const {showLoading, hideLoading} = useGlobalLoading()

    async function queryOrderStatus(orderId: string): Promise<QueryOrderStatusResponse> {
        return apiRequest<QueryOrderStatusResponse>(
            API_PATHS.WECHAT_NATIVE_ORDER_STATUS+`?order_id=${encodeURIComponent(orderId)}`
        )
    }

    function startPayPolling(orderId: string) {
        stopPayPolling()

        payPollingTimer.value = window.setInterval(async () => {
            try {
                const status = await queryOrderStatus(orderId)
                if (status.paid) {
                    paySucceeded.value = true
                    stopPayPolling()
                    opts.onSuccess()
                }
            } catch (err) {
                console.error('[Pay] queryOrderStatus error', err)
            }
        }, 2000)
    }

    function stopPayPolling() {
        if (payPollingTimer.value !== null) {
            window.clearInterval(payPollingTimer.value)
            payPollingTimer.value = null
        }
    }

    function resetAll() {
        code.value = ''
        inviteLoading.value = false
        payLoading.value = false
        errorMessage.value = ''
        paying.value = false
        payOrder.value = null
        paySucceeded.value = false
    }

    // 点击微信支付
    async function handleWeChatPayClick() {
        if (paying.value) return
        if (!opts.publicID) {
            console.warn('No product when trying to pay')
            return
        }

        try {
            paying.value = true
            const res = await apiRequest<NativeCreateOrderResponse>(API_PATHS.WECHAT_CREATE_NATIVE_ORDER, {
                method: 'POST',
                body: {
                    public_id: opts.publicID,
                },
            })

            payOrder.value = res
            startPayPolling(res.order_id)
        } catch (e) {
            console.error('[Pay] create native order error', e)
            if (isApiErr(e)) {
                showAlert(e.message)
                console.log('code:', e.code,'err:', e.err);
                return
            }

            showAlert('发生未知错误，请稍后重试')
        } finally {
            paying.value = false
        }
    }

    async function handleInviteConfirm() {
        if (inviteLoading.value) return

        inviteLoading.value = true
        errorMessage.value = ''
        try {
            const trimmed = code.value.trim()
            await verifyInviteWithMessage(trimmed)
            opts.onSuccess()
        } catch (e) {
            errorMessage.value = e instanceof Error ? e.message : String(e)
            inputRef.value?.focus()
            return
        } finally {
            inviteLoading.value = false
        }
    }

    async function verifyInviteWithMessage(rawCode: string) {
        await apiRequest(API_PATHS.INVITE_PAYMENT, {
            method: 'POST',
            body: {
                invite_code: rawCode,
                public_id: opts.publicID,
            },
        })
    }


    function handleCancel() {
        showAlert('您确定放弃本次测试报告吗？', () => {
            showLoading('结束报告')
            router
                .replace('/')
                .then(() => {
                    opts.onClose()
                })
                .finally(() => {
                    hideLoading()
                })
        })
    }

    return {
        // 状态
        paying,
        payOrder,
        payLoading,
        code,
        trimmedCode,
        inviteLoading,
        errorMessage,
        inputRef,
        // 方法
        handleWeChatPayClick,
        handleInviteConfirm,
        resetAll,
        stopPayPolling,
        handleCancel,
    }
}
