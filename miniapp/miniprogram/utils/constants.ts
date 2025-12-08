export const API_BASE = 'https://wenxintai.cn'

// REST 接口
export const API_PRODUCTS        = `${API_BASE}/api/products`
export const API_HOBBIES         = `${API_BASE}/api/hobbies`
export const API_WECHAT_SIGNIN   = `${API_BASE}/api/wechat_signin`
export const API_AUTH_STATUS     = `${API_BASE}/api/auth/wx/status`
export const API_BASIC_INFO      = `${API_BASE}/api/tests/basic_info`
export const API_TEST_SUBMIT     = `${API_BASE}/api/test_submit`
export const API_GENERATE_REPORT = `${API_BASE}/api/generate_report`
export const API_PREPARE_PAY     = `${API_BASE}/api/prepare_pay`
export const API_FINISH_REPORT   = `${API_BASE}/api/finish_report`
export const API_INVITE_CODE     = `${API_BASE}/api/pay/use_invite`
export const API_PAY_ORDER       = `${API_BASE}/api/pay/wechat/order_create`
export const API_PAY_STATUS      = `${API_BASE}/api/pay/wechat/order_status`
export const API_LOGIN           = `${API_BASE}/api/auth/miniapp_login`

// WS / SSE 之类的 URL 构造
export const buildQuestionWsUrl = (
  publicId: string,
  businessType?: string,
  testType?: string,
) => {
  const searchParams: string[] = []
  if (businessType) searchParams.push(`business_type=${encodeURIComponent(businessType)}`)
  if (testType)     searchParams.push(`test_type=${encodeURIComponent(testType)}`)
  const query = searchParams.length ? `?${searchParams.join('&')}` : ''
  return `${API_BASE}/api/ws/question/${publicId}${query}`
}

export const buildReportWsUrl = (publicId: string) =>
  `${API_BASE}/api/ws/report/${publicId}`
