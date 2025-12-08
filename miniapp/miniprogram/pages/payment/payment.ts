import { API_INVITE_CODE, API_PAY_ORDER, API_PAY_STATUS } from '../../utils/constants'
import { get, post } from '../../utils/request'

Page({
  data: {
    orderToken: '',
    inviteCode: '',
    publicId: '',
    status: '',
    paying: false,
  },
  onLoad(options: Record<string, string>) {
    this.setData({ publicId: options?.public_id || '' })
  },
  onInviteInput(event: WechatMiniprogram.Input) {
    this.setData({ inviteCode: event.detail.value })
  },
  async useInvite() {
    if (!this.data.inviteCode) return
    try {
      await post(API_INVITE_CODE, { invite_code: this.data.inviteCode, public_id: this.data.publicId })
      this.setData({ status: '邀请码验证成功，可直接查看报告' })
    } catch (err: any) {
      this.setData({ status: err?.message || '邀请码无效' })
    }
  },
  async createOrder() {
    this.setData({ paying: true, status: '' })
    try {
      const payRes = await post<any>(API_PAY_ORDER, { public_id: this.data.publicId })
      this.setData({ orderToken: payRes.order_token || '' })
      await wx.requestPayment({
        timeStamp: payRes.timeStamp,
        nonceStr: payRes.nonceStr,
        package: payRes.package,
        signType: payRes.signType,
        paySign: payRes.paySign,
        success: () => this.setData({ status: '支付完成' }),
        fail: () => this.pollStatus(),
      })
    } catch (err: any) {
      this.setData({ status: err?.message || '支付发起失败' })
    } finally {
      this.setData({ paying: false })
    }
  },
  async pollStatus() {
    if (!this.data.orderToken) return
    try {
      const res = await get<{ paid: boolean }>(`${API_PAY_STATUS}?order_token=${this.data.orderToken}`)
      if (res.paid) this.setData({ status: '支付成功' })
    } catch (err) {
      this.setData({ status: '查询支付状态失败' })
    }
  },
})
