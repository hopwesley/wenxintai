import { API_AUTH_STATUS } from './utils/constants'
import { initSession, setAuthInfo } from './utils/store'
import { request } from './utils/request'

App<IAppOption>({
  globalData: {
    session: initSession(),
    sessionReady: false,
  },
  async onLaunch() {
    this.globalData.session = initSession()
    try {
      if (this.globalData.session.token || this.globalData.session.cookie) {
        const status = await request<{ token?: string; userInfo?: WechatMiniprogram.UserInfo }>({
          url: API_AUTH_STATUS,
          method: 'GET',
        })
        if (status.token) {
          setAuthInfo({ token: status.token, loggedIn: true, userInfo: status.userInfo })
          this.globalData.session = initSession()
        }
      }
    } catch (err) {
      console.log('auth status skipped', err)
    }
    this.globalData.sessionReady = true
  },
})
