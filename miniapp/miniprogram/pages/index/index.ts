import { API_AUTH_STATUS, API_HOBBIES, API_PRODUCTS } from '../../utils/constants'
import { get, request } from '../../utils/request'
import { cacheBasicInfo, initSession, setAuthInfo, setCurrentTest } from '../../utils/store'

interface Product {
  id: string
  name: string
  description?: string
  business_type?: string
  test_type?: string
}

interface Hobby {
  id: string
  name: string
}

Page({
  data: {
    products: [] as Product[],
    hobbies: [] as Hobby[],
    selectedProductId: '',
    selectedHobbies: [] as string[],
    loading: false,
    statusMessage: '',
    hasSession: false,
  },
  async onLoad() {
    initSession()
    await this.checkSession()
    this.loadCatalogs()
  },
  async checkSession() {
    try {
      const status = await request<{ token?: string; current_test?: IAppSession['currentTest'] }>({
        url: API_AUTH_STATUS,
        method: 'GET',
      })
      if (status.token) {
        setAuthInfo({ token: status.token, loggedIn: true })
        setCurrentTest(status.current_test)
      }
      this.setData({ hasSession: Boolean(status.token) })
    } catch (err) {
      this.setData({ statusMessage: '尚未登录或读取状态失败，请先登录' })
    }
  },
  async loadCatalogs() {
    this.setData({ loading: true })
    try {
      const [products, hobbies] = await Promise.all([
        get<Product[]>(API_PRODUCTS),
        get<Hobby[]>(API_HOBBIES),
      ])
      this.setData({ products, hobbies, statusMessage: '' })
    } catch (err: any) {
      this.setData({ statusMessage: err?.message || '加载数据失败' })
    } finally {
      this.setData({ loading: false })
    }
  },
  onProductChange(event: WechatMiniprogram.RadioGroupChange) {
    const selectedProductId = event.detail.value as string
    this.setData({ selectedProductId })
  },
  onHobbyChange(event: WechatMiniprogram.CheckboxGroupChange) {
    this.setData({ selectedHobbies: event.detail.value as string[] })
  },
  goToLogin() {
    wx.navigateTo({ url: '/pages/login/login' })
  },
  goToBasicInfo() {
    if (!this.data.selectedProductId) {
      wx.showToast({ title: '请选择要测评的产品', icon: 'none' })
      return
    }
    cacheBasicInfo(this.data.selectedProductId, { hobbies: this.data.selectedHobbies })
    const selectedProduct = this.data.products.find((p) => p.id === this.data.selectedProductId)
    setCurrentTest({
      publicId: undefined,
      businessType: selectedProduct?.business_type,
      testType: selectedProduct?.test_type,
    })
    wx.navigateTo({
      url: `/pages/basicinfo/basicinfo?product_id=${this.data.selectedProductId}`,
    })
  },
  continueTest() {
    const current = initSession().currentTest
    if (current?.nextRoute === 'report') {
      wx.navigateTo({ url: `/pages/report/report?public_id=${current.publicId || ''}` })
      return
    }
    wx.navigateTo({
      url: '/pages/questions/questions',
    })
  },
})
