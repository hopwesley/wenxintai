import {
  API_FINISH_REPORT,
  API_GENERATE_REPORT,
  API_INVITE_CODE,
  API_PAY_ORDER,
  API_PAY_STATUS,
  API_PREPARE_PAY,
  buildReportWsUrl,
} from '../../utils/constants'
import { get, post } from '../../utils/request'
import { setCurrentTest } from '../../utils/store'
import { connectWebSocket, ManagedSocket } from '../../utils/websocket'

Page({
  data: {
    publicId: '',
    report: null as any,
    logs: [] as string[],
    loading: false,
    inviteCode: '',
    awaitingPay: false,
    dimensionText: '',
  },
  socket: null as ManagedSocket | null,
  onLoad(options: Record<string, string>) {
    const publicId = options?.public_id || ''
    this.setData({ publicId })
    this.loadOrGenerate(publicId)
  },
  onUnload() {
    if (this.socket) this.socket.close()
  },
  async loadOrGenerate(publicId: string) {
    this.setData({ loading: true })
    try {
      const res = await get<{ report?: any; need_pay?: boolean }>(API_GENERATE_REPORT)
      if (res.report) {
        this.handleReportReady(res.report)
      } else {
        this.startReportStream(publicId)
      }
    } catch (err: any) {
      this.pushLog(err?.message || '加载报告失败')
    } finally {
      this.setData({ loading: false })
    }
  },
  startReportStream(publicId: string) {
    if (!publicId) return
    const socket = connectWebSocket(buildReportWsUrl(publicId), {
      onData: (payload) => this.pushLog(payload),
      onDone: (payload) => this.handleReportReady(payload),
      onError: (message) => this.pushLog(message),
    })
    socket.sendPing()
    this.socket = socket
  },
  handleReportReady(payload: any) {
    let reportData = payload
    if (typeof payload === 'string') {
      try {
        reportData = JSON.parse(payload)
      } catch (err) {
        this.pushLog('报告解析失败')
        return
      }
    }
  
    // 这里安全地计算维度字符串
    let dimensionText = ''
    const dims = (reportData && reportData.dimensions) as any
    if (Array.isArray(dims)) {
      // 确保每个元素都是字符串，避免奇怪类型
      dimensionText = dims.map((d) => String(d)).join('、')
    } else if (typeof dims === 'string') {
      // 后端如果直接给了字符串，也兜一下
      dimensionText = dims
    }
  
    this.setData({
      report: reportData,
      dimensionText,
    })
  
    setCurrentTest({ publicId: this.data.publicId, nextRoute: 'report' })
  },
  pushLog(message: string) {
    const logs = this.data.logs.concat(`${new Date().toLocaleTimeString()} ${message}`)
    this.setData({ logs })
  },
  async preparePay() {
    this.setData({ awaitingPay: true })
    try {
      const prepare = await post<{ order_token: string }>(API_PREPARE_PAY, { public_id: this.data.publicId })
      await this.createPayOrder(prepare.order_token)
    } catch (err: any) {
      wx.showToast({ title: err?.message || '预下单失败', icon: 'none' })
    } finally {
      this.setData({ awaitingPay: false })
    }
  },
  async createPayOrder(orderToken: string) {
    try {
      const payRes = await post<any>(API_PAY_ORDER, { order_token: orderToken })
      await wx.requestPayment({
        timeStamp: payRes.timeStamp,
        nonceStr: payRes.nonceStr,
        package: payRes.package,
        signType: payRes.signType,
        paySign: payRes.paySign,
        success: async () => {
          await this.confirmReport()
        },
        fail: async () => {
          await this.pollPayStatus(orderToken)
        },
      })
    } catch (err: any) {
      this.pushLog(err?.message || '下单失败')
    }
  },
  async pollPayStatus(orderToken: string) {
    try {
      const status = await get<{ paid: boolean }>(`${API_PAY_STATUS}?order_token=${orderToken}`)
      if (status.paid) {
        await this.confirmReport()
      }
    } catch (err) {
      this.pushLog('支付状态查询失败')
    }
  },
  async confirmReport() {
    try {
      await post(API_FINISH_REPORT, { public_id: this.data.publicId })
      this.pushLog('支付完成，报告生成中...')
      this.startReportStream(this.data.publicId)
    } catch (err) {
      this.pushLog('确认报告失败')
    }
  },
  onInviteInput(event: WechatMiniprogram.Input) {
    this.setData({ inviteCode: event.detail.value })
  },
  async useInvite() {
    if (!this.data.inviteCode) return
    try {
      await post(API_INVITE_CODE, { invite_code: this.data.inviteCode, public_id: this.data.publicId })
      this.pushLog('邀请码验证成功，直接生成报告')
      await this.confirmReport()
    } catch (err: any) {
      wx.showToast({ title: err?.message || '邀请码无效', icon: 'none' })
    }
  },
})
