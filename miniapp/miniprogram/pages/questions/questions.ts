import { API_TEST_SUBMIT, buildQuestionWsUrl } from '../../utils/constants'
import { post } from '../../utils/request'
import { cacheAnswers, getCachedAnswers, setCurrentTest } from '../../utils/store'
import { connectWebSocket, ManagedSocket } from '../../utils/websocket'

interface QuestionOption {
  id: string
  text: string
}

interface Question {
  id: string
  title: string
  type: 'single' | 'multiple'
  options: QuestionOption[]
}

Page({
  data: {
    publicId: '',
    businessType: '',
    testType: '',
    questions: [] as Question[],
    answers: {} as Record<string, string | string[]>,
    logs: [] as string[],
    submitting: false,
  },

  socketTask: null as ManagedSocket | null,
  questionBuffer: '',

  onLoad(options: Record<string, string>) {
    const publicId =
      options && options.public_id ? options.public_id : ''
    const businessType =
      options && options.business_type ? options.business_type : ''
    const testType =
      options && options.test_type ? options.test_type : ''

    this.setData({ publicId, businessType, testType })
    this.restoreCache(publicId)
    this.openQuestionStream(publicId, businessType, testType)
  },

  onUnload() {
    if (this.socketTask) {
      this.socketTask.close()
    }
  },

  restoreCache(publicId: string) {
    const cache = getCachedAnswers(publicId)
    if (cache) {
      this.setData({
        answers: cache.answers,
      })
    }
  },

  openQuestionStream(publicId: string, businessType?: string, testType?: string) {
    if (!publicId) return
    const url = buildQuestionWsUrl(publicId, businessType, testType)
    const socket = connectWebSocket(url, {
      onData: (payload) => this.onQuestionChunk(payload),
      onDone: (payload) => this.onQuestionsReady(payload),
      onError: (message) => this.pushLog(message),
      onLog: (message) => this.pushLog(message),
    })
    socket.sendPing()
    this.socketTask = socket
  },

  onQuestionChunk(payload: any) {
    if (typeof payload === 'string') {
      this.questionBuffer += payload
      return
    }
    this.pushLog('收到题目分片')
  },

  onQuestionsReady(payload: any) {
    let dataSource: any = payload

    if (typeof payload === 'string') {
      try {
        dataSource = JSON.parse(this.questionBuffer + payload)
      } catch (err) {
        this.pushLog('题目解析失败')
        return
      }
    }

    if (Array.isArray(dataSource)) {
      this.setData({ questions: dataSource })
    } else if (dataSource && dataSource.questions) { // 替换 dataSource?.questions
      this.setData({ questions: dataSource.questions })
    }

    this.pushLog('题目加载完成')
  },

  pushLog(message: string) {
    const logs = this.data.logs.concat(
      `${new Date().toLocaleTimeString()} ${message}`,
    )
    this.setData({ logs })
  },

  onSingleChange(event: WechatMiniprogram.RadioGroupChange) {
    const questionId = event.currentTarget.dataset.qid as string
    this.setData({
      [`answers.${questionId}`]: event.detail.value,
    })
    this.persistAnswers()
  },

  onMultiChange(event: WechatMiniprogram.CheckboxGroupChange) {
    const questionId = event.currentTarget.dataset.qid as string
    this.setData({
      [`answers.${questionId}`]: event.detail.value as string[],
    })
    this.persistAnswers()
  },

  persistAnswers() {
    cacheAnswers(this.data.publicId, this.data.answers)
  },

  async submitPage() {
    if (!this.data.publicId) return
    this.setData({ submitting: true })

    try {
      const res = await post<{ next_route?: string; public_id?: string }>(
        API_TEST_SUBMIT,
        {
          public_id: this.data.publicId,
          business_type: this.data.businessType,
          test_type: this.data.testType,
          answers: this.data.answers,
        },
      )

      setCurrentTest({
        publicId: res.public_id || this.data.publicId,
        nextRoute: res.next_route,
      })

      if (res.next_route === 'report') {
        wx.navigateTo({
          url: `/pages/report/report?public_id=${
            res.public_id || this.data.publicId
          }`,
        })
      }
    } catch (err: any) {
      const msg =
        err && err.message ? String(err.message) : '提交失败'
      wx.showToast({
        title: msg,
        icon: 'none',
      })
    } finally {
      this.setData({ submitting: false })
    }
  },
})
