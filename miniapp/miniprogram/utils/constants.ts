export const API_PRODUCTS = '/api/products'
export const API_HOBBIES = '/api/hobbies'
export const API_WECHAT_SIGNIN = '/api/wechat_signin'
export const API_AUTH_STATUS = '/api/auth/wx/status'
export const API_BASIC_INFO = '/api/tests/basic_info'
export const API_TEST_SUBMIT = '/api/test_submit'
export const API_GENERATE_REPORT = '/api/generate_report'
export const API_PREPARE_PAY = '/api/prepare_pay'
export const API_FINISH_REPORT = '/api/finish_report'
export const API_INVITE_CODE = '/api/pay/use_invite'
export const API_PAY_ORDER = '/api/pay/wechat/order_create'
export const API_PAY_STATUS = '/api/pay/wechat/order_status'

export const buildQuestionWsUrl = (publicId: string, businessType?: string, testType?: string) => {
  const searchParams = [] as string[]
  if (businessType) searchParams.push(`business_type=${encodeURIComponent(businessType)}`)
  if (testType) searchParams.push(`test_type=${encodeURIComponent(testType)}`)
  const query = searchParams.length ? `?${searchParams.join('&')}` : ''
  return `/api/ws/question/${publicId}${query}`
}

export const buildReportWsUrl = (publicId: string) => `/api/ws/report/${publicId}`
