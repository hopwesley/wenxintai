import { getSession, setAuthInfo } from './store'

export interface RequestOptions<TData = any> {
  url: string
  method?: WechatMiniprogram.RequestOption['method']
  data?: TData
  header?: Record<string, string>
}

interface RequestError {
  message: string
  status?: number
  data?: unknown
}

const DEFAULT_HEADERS = {
  'Content-Type': 'application/json',
}

const mergeHeaders = (headers?: Record<string, string>) => {
  const session = getSession()
  const authHeaders: Record<string, string> = {}
  if (session.token) {
    authHeaders['Authorization'] = `Bearer ${session.token}`
  }
  if (session.cookie) {
    authHeaders['Cookie'] = session.cookie
  }
  return { ...DEFAULT_HEADERS, ...headers, ...authHeaders }
}

export const request = <T = any, U = any>(
  options: RequestOptions<U>,
): Promise<T> =>
  new Promise<T>((resolve, reject) => {
    const headers = mergeHeaders(options.header)

    wx.request<T>({
      url: options.url,
      data: options.data,
      method: options.method || 'GET',
      header: headers,
      success: (res) => {
        const { statusCode, data, header } = res

        // 这里 header 类型一般是 Record<string, string> | undefined
        const cookie =
          (header && (header['Set-Cookie'] as string)) || undefined
        if (cookie) {
          setAuthInfo({ cookie })
        }

        if (statusCode && statusCode >= 200 && statusCode < 300) {
          resolve(data)
          return
        }
        const err: RequestError = {
          message: `Request failed with status ${statusCode}`,
          status: statusCode,
          data,
        }
        reject(err)
      },
      fail: (err) => {
        const error: RequestError = {
          message: err.errMsg || 'Network error',
        }
        reject(error)
      },
    })
  })

export const get = <T = any>(url: string, data?: Record<string, any>) =>
  request<T>({
    url,
    method: 'GET',
    data,
  })

export const post = <T = any, U = any>(url: string, data?: U) =>
  request<T, U>({
    url,
    method: 'POST',
    data,
  })
