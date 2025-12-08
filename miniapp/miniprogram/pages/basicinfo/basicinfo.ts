import { API_BASIC_INFO, API_HOBBIES } from '../../utils/constants'
import { get, post } from '../../utils/request'
import { cacheBasicInfo, getCachedBasicInfo, setCurrentTest } from '../../utils/store'

interface BasicInfoForm {
  province: string
  city: string
  grade: string
  gender: string
  phone: string
  hobbies: string[]
  interests: string[]
}

Page({
  data: {
    productId: '',
    form: {
      province: '',
      city: '',
      grade: '',
      gender: '',
      phone: '',
      hobbies: [] as string[],
      interests: [] as string[],
    } as BasicInfoForm,
    hobbyOptions: [] as { id: string; name: string }[],
    gradeOptions: ['小学', '初中', '高中', '大学'],
    genderOptions: ['男', '女'],
    loading: false,
    status: '',
  },

  async onLoad(options: Record<string, string>) {
    // 去掉 ?.
    const productId = options && options.product_id ? options.product_id : ''
    this.setData({ productId })
    await this.loadHobbies()
    this.restoreCache(productId)
  },

  async loadHobbies() {
    try {
      const hobbyOptions = await get<{ id: string; name: string }[]>(API_HOBBIES)
      this.setData({ hobbyOptions })
    } catch (err) {
      // ignore
    }
  },

  restoreCache(productId: string) {
    const cache = getCachedBasicInfo(productId)
    if (cache) {
      this.setData({
        form: {
          ...this.data.form,
          ...cache,
        },
      })
    }
  },

  onRegionChange(event: any) {
    const value = event.detail && event.detail.value ? event.detail.value as string[] : []
    const province = value[0] || ''
    const city = value[1] || ''
    this.setData({
      'form.province': province,
      'form.city': city,
    })
  },

  onGradeChange(event: WechatMiniprogram.PickerChange) {
    const index = Number(event.detail.value)
    const grade = this.data.gradeOptions[index]
    this.setData({ 'form.grade': grade })
  },

  onGenderChange(event: WechatMiniprogram.PickerChange) {
    const index = Number(event.detail.value)
    const gender = this.data.genderOptions[index]
    this.setData({ 'form.gender': gender })
  },

  onPhoneInput(event: WechatMiniprogram.Input) {
    this.setData({ 'form.phone': event.detail.value })
  },

  onInterestInput(event: WechatMiniprogram.Input) {
    const val = event.detail.value || ''
    this.setData({ 'form.interests': val ? [val] : [] })
  },

  onHobbyChange(event: WechatMiniprogram.CheckboxGroupChange) {
    this.setData({ 'form.hobbies': event.detail.value as string[] })
  },

  validateForm(form: BasicInfoForm) {
    if (!form.province || !form.city) return '请选择所在地区'
    if (!form.grade) return '请选择年级'
    if (!form.gender) return '请选择性别'
    if (!/^1[3-9]\d{9}$/.test(form.phone)) return '请输入有效手机号'
    return ''
  },

  async submitForm() {
    const message = this.validateForm(this.data.form)
    if (message) {
      wx.showToast({ title: message, icon: 'none' })
      return
    }

    this.setData({ loading: true, status: '' })

    cacheBasicInfo(this.data.productId, this.data.form)

    try {
      const payload = {
        ...this.data.form,
        product_id: this.data.productId,
      }

      const res = await post<{ next_route?: string; public_id?: string }>(
        API_BASIC_INFO,
        payload,
      )

      setCurrentTest({
        publicId: res.public_id,
        nextRoute: res.next_route,
      })

      if (res.next_route === 'questions') {
        wx.navigateTo({
          url: `/pages/questions/questions?public_id=${res.public_id}`,
        })
      } else {
        wx.navigateTo({ url: '/pages/report/report' })
      }
    } catch (err: any) {
      const msg =
        err && err.message
          ? String(err.message)
          : '提交失败，请稍后再试'
      this.setData({ status: msg })
    } finally {
      this.setData({ loading: false })
    }
  },
})
