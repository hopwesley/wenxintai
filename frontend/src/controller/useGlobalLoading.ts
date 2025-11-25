// src/controller/useGlobalLoading.ts
import {ref} from 'vue'

const visible = ref(false)
const message = ref<string>('正在加载...')
let hideTimer: number | null = null

export function useGlobalLoading() {
    function showLoading(
        msg?: string,
        durationMs?: number,  // 可选：自动隐藏的时长（毫秒）
    ) {
        message.value = msg ?? '正在加载...'
        visible.value = true

        // 先清理旧的定时器
        if (hideTimer !== null) {
            window.clearTimeout(hideTimer)
            hideTimer = null
        }

        // 如果传了时长，就自动关
        if (durationMs && durationMs > 0) {
            hideTimer = window.setTimeout(() => {
                visible.value = false
                hideTimer = null
            }, durationMs)
        }
    }

    function hideLoading() {
        visible.value = false
        if (hideTimer !== null) {
            window.clearTimeout(hideTimer)
            hideTimer = null
        }
    }

    return {
        // 状态：给全局容器组件用
        visible,
        message,
        // 操作：给业务代码调用
        showLoading,
        hideLoading,
    }
}
