export interface BasicInfoPayload {
  province?: string
  city?: string
  grade?: string
  gender?: string
  phone?: string
  hobbies?: string[]
  interests?: string[]
}

export interface AnswerCache {
  answers: Record<string, string | string[]>
  updatedAt: number
}

const SESSION_STORAGE_KEY = 'app_session_state'
const ANSWER_CACHE_PREFIX = 'qstage:'
const FORM_CACHE_PREFIX = 'basicinfo:'

let session: IAppSession = {
  loggedIn: false,
  token: undefined,
  cookie: undefined,
  userInfo: undefined,
  currentTest: {},
}

const persistSession = () => {
  wx.setStorageSync(SESSION_STORAGE_KEY, session)
}

export const initSession = (): IAppSession => {
  const stored = wx.getStorageSync(SESSION_STORAGE_KEY)
  if (stored) {
    session = { ...session, ...stored }
  }
  return session
}

export const getSession = (): IAppSession => session

export const setAuthInfo = (auth: Partial<IAppSession>) => {
  session = { ...session, ...auth, loggedIn: Boolean(auth.token || session.token) }
  persistSession()
}

export const setUserInfo = (userInfo?: WechatMiniprogram.UserInfo) => {
  session = { ...session, userInfo }
  persistSession()
}

export const setCurrentTest = (test?: IAppSession['currentTest']) => {
  session = { ...session, currentTest: test }
  persistSession()
}

export const clearSession = () => {
  session = { loggedIn: false, token: undefined, cookie: undefined, userInfo: undefined, currentTest: {} }
  persistSession()
}

export const cacheAnswers = (key: string, answers: Record<string, string | string[]>) => {
  const cache: AnswerCache = {
    answers,
    updatedAt: Date.now(),
  }
  wx.setStorageSync(`${ANSWER_CACHE_PREFIX}${key}`, cache)
}

export const getCachedAnswers = (key: string): AnswerCache | undefined => {
  return wx.getStorageSync(`${ANSWER_CACHE_PREFIX}${key}`)
}

export const cacheBasicInfo = (key: string, payload: BasicInfoPayload) => {
  wx.setStorageSync(`${FORM_CACHE_PREFIX}${key}`, payload)
}

export const getCachedBasicInfo = (key: string): BasicInfoPayload | undefined => {
  return wx.getStorageSync(`${FORM_CACHE_PREFIX}${key}`)
}
