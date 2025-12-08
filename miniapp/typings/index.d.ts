/// <reference path="./types/index.d.ts" />

interface IAppSession {
  token?: string
  cookie?: string
  loggedIn: boolean
  userInfo?: WechatMiniprogram.UserInfo
  currentTest?: {
    publicId?: string
    businessType?: string
    testType?: string
    nextRoute?: string
  }
}

interface IAppOption {
  globalData: {
    session: IAppSession
    sessionReady?: boolean
  }
  userInfoReadyCallback?: WechatMiniprogram.GetUserInfoSuccessCallback
}
