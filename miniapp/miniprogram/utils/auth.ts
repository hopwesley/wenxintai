import { post } from "./request"
import { API_LOGIN } from "./constants"

// needProfile = true -> 带头像昵称
// needProfile = false -> 只拿 unionid，不传头像昵称
export async function ensureLoginWithProfile(needProfile: boolean): Promise<void> {
  const token = wx.getStorageSync('auth_token')
  if (token) {
    return
  }

  // 1. wx.login 拿 code
  const loginRes = await wx.login()
  if (!loginRes.code) {
    throw new Error('微信登录失败')
  }

  let body: any = { code: loginRes.code }

  // 2. 如果需要头像昵称，再调 getUserProfile
  if (needProfile) {
    const profileRes = await wx.getUserProfile({
      desc: '用于展示头像和昵称',
      lang: 'zh_CN',
    })
    const { nickName, avatarUrl } = profileRes.userInfo
    body.nick_name = nickName
    body.avatar_url = avatarUrl
  }
  console.log('API_LOGIN =', API_LOGIN)
  // 3. 一次 POST 到后端
  const resp = await post<{ token: string }>(API_LOGIN, body)
  wx.setStorageSync('auth_token', resp.token)
}
