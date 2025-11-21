import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', () => {
    // 全局微信登录弹窗状态
    const wechatLoginOpen = ref(false)

    function openWeChatLogin() {
        wechatLoginOpen.value = true
    }

    function closeWeChatLogin() {
        wechatLoginOpen.value = false
    }

    function toggleWeChatLogin(val?: boolean) {
        wechatLoginOpen.value = val ?? !wechatLoginOpen.value
    }

    return {
        wechatLoginOpen,
        openWeChatLogin,
        closeWeChatLogin,
        toggleWeChatLogin,
    }
})
