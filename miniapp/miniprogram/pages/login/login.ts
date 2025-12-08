import { API_AUTH_STATUS, API_WECHAT_SIGNIN } from '../../utils/constants'
import { post, request } from '../../utils/request'
import { initSession, setAuthInfo } from '../../utils/store'

Page({
  data: {
    loading: false,
    status: '',
  },
  async onLoad() {
    await this.checkStatus()
  },
  async checkStatus() {
    this.setData({ loading: true })
    try {
      const res = await request<{ token?: string }>({ url: API_AUTH_STATUS, method: 'GET' })
      if (res.token) {
        setAuthInfo({ token: res.token, loggedIn: true })
        this.setData({ status: '已登录，即将跳转' })
        wx.navigateBack({ delta: 1 })
      }
    } catch (err) {
      // ignore
    } finally {
      this.setData({ loading: false })
    }
  },
  async handleLogin() {
    this.setData({ loading: true, status: '' })
    wx.login({
      success: async (loginRes) => {
        try {
          const signinRes = await post<{ token: string; cookie?: string }>(API_WECHAT_SIGNIN, {
            code: loginRes.code,
          })
          setAuthInfo({ token: signinRes.token, cookie: signinRes.cookie, loggedIn: true })
          initSession()
          this.setData({ status: '登录成功，返回首页' })
          wx.navigateBack({ delta: 1 })
        } catch (err: any) {
          this.setData({ status: err?.message || '登录失败，请重试' })
        } finally {
          this.setData({ loading: false })
        }
      },
      fail: () => {
        this.setData({ status: '无法获取微信登录态，请检查网络', loading: false })
      },
    })
  },
})
