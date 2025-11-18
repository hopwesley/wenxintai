import {ref} from 'vue'

const visible = ref(false)
const title = ref<string>('提示')
const message = ref<string>('')
let confirmCallback: (() => void) | null = null

export function useAlert() {
    function showAlert(
        msg: string,
        onConfirm?: () => void,
        customTitle?: string,
    ) {
        message.value = msg
        title.value = customTitle ?? '提示'
        confirmCallback = onConfirm ?? null
        visible.value = true
    }

    function handleConfirm() {
        const cb = confirmCallback
        confirmCallback = null
        visible.value = false
        if (cb) cb()
    }

    function closeAlert() {
        visible.value = false
        confirmCallback = null
    }

    return {
        // 状态（给全局容器用）
        visible,
        title,
        message,
        // 操作
        showAlert,
        handleConfirm,
        closeAlert,
    }
}
