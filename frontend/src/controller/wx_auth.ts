import { ref, nextTick } from 'vue'
import { defineStore } from 'pinia'
import { apiRequest } from '@/api'

export type WxLoginStatus = 'idle' | 'pending' | 'success' | 'error' | 'expired'
const VITE_WX_APPID="wx51cf75df014d41e8"
const VITE_WX_REDIRECT_URI="https://sharp-happy-grouse.ngrok-free.app/api/wechat_signin"
interface WxLoginStatusResponse {
    status: 'pending' | 'ok' | 'expired'
    is_new?: boolean
}

let pollTimer: number | null = null

function genState(): string {
    return 'wx_' + Date.now().toString(36) + '_' + Math.random().toString(36).slice(2, 10)
}

export const useAuthStore = defineStore('auth', () => {
    // 弹窗是否打开
    const wechatLoginOpen = ref(false)

    // 当前这次扫码登录的 state
    const loginState = ref<string | null>(null)

    // 当前扫码登录状态
    const loginStatus = ref<WxLoginStatus>('idle')

    // 是否新用户（由后端判断）
    const isNewUser = ref<boolean | null>(null)

    // 全局登录态（你后面可以在别的页面用它来控制“仅登录可见”按钮）
    const isLoggedIn = ref(false)

    function clearTimer() {
        if (pollTimer !== null) {
            window.clearInterval(pollTimer)
            pollTimer = null
        }
    }

    function resetLoginState() {
        clearTimer()
        loginState.value = null
        loginStatus.value = 'idle'
        isNewUser.value = null
    }
    /**
     * 开始一次新的微信扫码登录流程：
     * 1) 生成 state
     * 2) 打开弹窗
     * 3) 等待 DOM 更新后，用 WxLogin.js 渲染二维码
     * 4) 轮询后端登录状态
     */
    async function startWeChatLogin() {
        resetLoginState()
        const state = genState()
        loginState.value = state
        loginStatus.value = 'pending'
        wechatLoginOpen.value = true

        if (typeof window === 'undefined') {
            console.warn('[auth] window is undefined, skip wx login init')
            return
        }

        await nextTick()

        const wxLoginCtor = (window as any).WxLogin
        if (!wxLoginCtor) {
            console.error('[auth] WxLogin script is not loaded')
            loginStatus.value = 'error'
            return
        }

        if (!VITE_WX_APPID || !VITE_WX_REDIRECT_URI) {
            console.error('[auth] VITE_WX_APPID / VITE_WX_REDIRECT_URI 未配置')
            loginStatus.value = 'error'
            return
        }

        const redirectUri = encodeURIComponent(VITE_WX_REDIRECT_URI)

        // 注意：这个容器在 HomeView.vue 里
        new wxLoginCtor({
            id: 'wx-login-qrcode',
            appid: VITE_WX_APPID,
            scope: 'snsapi_login',
            redirect_uri: redirectUri,
            state,
            style: '',
            href: ''
        })

        startPolling()
    }

    async function pollOnce() {
        if (!loginState.value) return

        try {
            const res = await apiRequest<WxLoginStatusResponse>(
                `/api/auth/wx/status?state=${encodeURIComponent(loginState.value)}`,
                { method: 'GET' },
            )

            if (res.status === 'ok') {
                // 后端确认登录成功
                loginStatus.value = 'success'
                isNewUser.value = !!res.is_new
                isLoggedIn.value = true
                wechatLoginOpen.value = false
                clearTimer()
            } else if (res.status === 'expired') {
                // 后端认为二维码 / 登录会话过期
                loginStatus.value = 'expired'
                clearTimer()
            } else {
                // pending -> 继续轮询
            }
        } catch (e) {
            console.error('[auth] poll wx login status failed', e)
            // 短暂网络错误先不把状态标记为 error，避免闪断
        }
    }

    function startPolling() {
        clearTimer()
        const maxMs = 2 * 60 * 1000 // 最多轮询 2 分钟
        const start = Date.now()

        pollTimer = window.setInterval(() => {
            if (Date.now() - start > maxMs) {
                loginStatus.value = 'expired'
                clearTimer()
                return
            }
            void pollOnce()
        }, 1500)
    }

    return {
        wechatLoginOpen,
        loginState,
        loginStatus,
        isNewUser,
        isLoggedIn,
        startWeChatLogin,
    }
})
