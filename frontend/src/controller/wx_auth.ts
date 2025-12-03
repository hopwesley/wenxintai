import {ref, nextTick, computed, watch} from 'vue'
import {defineStore} from 'pinia'
import {API_PATHS, apiRequest} from '@/api'

export type WxLoginStatus = 'idle' | 'pending' | 'success' | 'error' | 'expired'
const NEW_USER_HINT_KEY_PREFIX = 'wenxintai:newUserInfoDismissed:'
/**
 * 对应 Go 里的 wxStatusResponse：
 * type wxStatusResponse struct {
 *   Status    string `json:"status"`
 *   IsNew     *bool  `json:"is_new,omitempty"`
 *   NickName  string `json:"nick_name,omitempty"`
 *   AvatarURL string `json:"avatar_url,omitempty"`
 * }
 *
 * 额外增加一个本地状态 'signOut'，表示“明确退出登录”。
 */
export interface WxLoginStatusResponse {
    status: 'pending' | 'ok' | 'expired' | 'signOut'
    uid?:string
    is_new?: boolean
    nick_name?: string
    avatar_url?: string
    appid?: string
    redirect_uri?: string
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

    // 当前扫码登录状态（只管“这一次扫码流程”）
    const loginStatus = ref<WxLoginStatus>('idle')

    // 全局登录态（是否有已登录用户）
    const isLoggedIn = ref(false)

    // 当前登录用户的状态（给 HomeView 用来显示头像 / 昵称）
    const signInStatus = ref<WxLoginStatusResponse>({
        status: 'signOut',
    })

    const newUserInfoDismissed = ref(false)
    // 当前用户的本地存储 key
    const currentUserKey = computed(() => {
        const id = signInStatus.value.uid || ''
        return id ? NEW_USER_HINT_KEY_PREFIX + id : ''
    })

// 每当登录用户变了，就重新从 localStorage 读取这一位用户的设置
    watch(currentUserKey, (key) => {
        if (!key || typeof window === 'undefined') {
            newUserInfoDismissed.value = false
            return
        }
        const stored = window.localStorage.getItem(key)
        newUserInfoDismissed.value = stored === '1'
    })

// 勾选“下次不再提醒”时，只对当前 userKey 写入
    function dismissNewUserInfoHint() {
        newUserInfoDismissed.value = true
        if (!currentUserKey.value || typeof window === 'undefined') return
        window.localStorage.setItem(currentUserKey.value, '1')
    }

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
    }

    function cancelWeChatLogin() {
        resetLoginState()
        wechatLoginOpen.value = false
    }

    /**
     * 开始一次新的微信扫码登录流程：
     * 1) 生成 state
     * 2) 打开弹窗
     * 3) 等待 DOM 更新后，用 WxLogin.js 渲染二维码
     * 4) 轮询后端登录状态
     */
    async function startWeChatLogin() {

        await ensureWxConfig()

        const appid = signInStatus.value.appid
        const rawRedirect = signInStatus.value.redirect_uri

        if (!appid || !rawRedirect) {
            console.error('[auth] 微信登录配置缺失：appid 或 redirect_uri 为空')
            loginStatus.value = 'error'
            return
        }

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

        const redirectUri = encodeURIComponent(rawRedirect)

        // 注意：这个容器在 HomeView.vue 里
        new wxLoginCtor({
            id: 'wx-login-qrcode',
            appid: appid,
            scope: 'snsapi_login',
            redirect_uri: redirectUri,
            state,
            style: '',
            href: '',
        })

        startPolling()
    }

    /**
     * 轮询一次 /api/auth/wx/status
     * 注意：后端现在是“只看 cookie，不看 state”，所以 ?state=... 只是兼容保留。
     */
    async function pollOnce() {
        if (!loginState.value) return

        try {
            const res = await apiRequest<WxLoginStatusResponse>(
                `/api/auth/wx/status?state=${encodeURIComponent(loginState.value)}`,
                {method: 'GET'},
            )

            // 把返回结果直接同步到 signInStatus
            signInStatus.value = res

            if (res.status === 'ok') {
                loginStatus.value = 'success'
                isLoggedIn.value = true
                wechatLoginOpen.value = false
                clearTimer()
            } else if (res.status === 'expired') {
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

    /**
     * 页面加载时，根据 cookie 静默检查一次登录状态
     * HomeView 之后可以在 onMounted 里调用：
     *   const auth = useAuthStore()
     *   auth.fetchSignInStatus()
     */
    async function fetchSignInStatus() {
        try {
            const res = await apiRequest<WxLoginStatusResponse>(API_PATHS.WECHAT_SIGN_IN, {
                method: 'GET',
            })

            signInStatus.value = res

            if (res.status === 'ok') {
                isLoggedIn.value = true
                loginStatus.value = 'success'
            } else {
                isLoggedIn.value = false
                if (res.status === 'expired') {
                    loginStatus.value = 'expired'
                } else {
                    loginStatus.value = 'idle'
                }
            }
        } catch (e) {
            console.error('[auth] fetchSignInStatus failed', e)
            // 出错就保持现状，不强制改状态
        }
    }

    async function ensureWxConfig() {
        if (signInStatus.value.appid && signInStatus.value.redirect_uri) {
            return
        }
        await fetchSignInStatus()
    }
    /**
     * 退出登录：清服务端 cookie + 清本地状态
     */
    async function logout() {
        try {
            await apiRequest('/api/auth/logout', {method: 'POST'})
        } catch (e) {
            console.error('[auth] logout failed', e)
            // 即使接口失败，本地也可以先清状态
        }
        clearTimer()
        loginState.value = null
        loginStatus.value = 'idle'
        isLoggedIn.value = false
        signInStatus.value = {status: 'signOut'}
        wechatLoginOpen.value = false
    }

    return {
        // 状态
        wechatLoginOpen,
        loginState,
        loginStatus,
        isLoggedIn,
        signInStatus,
        // 行为
        startWeChatLogin,
        cancelWeChatLogin,
        fetchSignInStatus,
        logout,
        newUserInfoDismissed,
        dismissNewUserInfoHint,
    }
})
