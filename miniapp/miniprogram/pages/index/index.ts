import { ensureLoginWithProfile } from "../../utils/auth"
import { API_PRODUCTS } from "../../utils/constants"

interface PlanInfoDTO {
  key: string
  name: string
  price: number
  desc: string
  tag?: string
  has_paid: boolean
}

Page({
  data: {
    plans: [] as PlanInfoDTO[],
    showProfileDialog: false,
  },

  onLoad() {
    const authorized = wx.getStorageSync('profile_authorized')
    if (!authorized) {
      this.setData({ showProfileDialog: true })
    }

    this.loadProducts()
  },


  // 拉产品列表
  loadProducts() {
    wx.request<PlanInfoDTO[]>({
      url: API_PRODUCTS,
      method: 'GET',
      success: (res) => {
        const data = res.data || []
        this.setData({ plans: data })
      },
      fail: (err) => {
        console.error('loadProducts error', err)
        wx.showToast({
          title: '获取产品列表失败，请稍后重试',
          icon: 'none',
        })
      },
    })
  },

  // 点击「暂不授权」
  async onProfileDialogCancel() {
    try {
      // 只登录，不要头像昵称
      await ensureLoginWithProfile(false)
      wx.setStorageSync('miniapp_login_done', true)
    } catch (err: any) {
      wx.showToast({
        title: err?.message || '登录失败，请稍后重试',
        icon: 'none',
      })
    }
    this.setData({ showProfileDialog: false })
  },

  async onAuthorizeAndLogin() {
    try {
      await ensureLoginWithProfile(true)
      wx.setStorageSync('miniapp_login_done', true)
      this.setData({ showProfileDialog: false })
  
      wx.showToast({
        title: '已授权头像昵称',
        icon: 'success',
      })
    } catch (err: any) {
      wx.showToast({
        title: err?.message || '授权失败，请稍后重试',
        icon: 'none',
      })
    }
  },
  
  async onPlanTap(e: WechatMiniprogram.TouchEvent) {
    const id = e.currentTarget.dataset.id as string
    try {
      await ensureLoginWithProfile(false)
      wx.navigateTo({
        url: `/pages/basicinfo/basicinfo?plan_id=${id}`,
      })
    } catch (err: any) {
      wx.showToast({
        title: err?.message || '登录失败，请稍后重试',
        icon: 'none',
      })
    }
  },
})