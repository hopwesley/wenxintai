<template>
  <div v-if="open" class="new-user-dialog-backdrop">
    <div class="new-user-dialog">

      <!-- 顶部标题 + 关闭按钮 -->
      <div class="dialog-header">
        <h2 class="dialog-title">完善资料</h2>
        <button class="dialog-close" @click="handleConfirm('skip')">✕</button>
      </div>

      <!-- 头像 / 昵称区域 -->
      <div class="profile-box">
        <div class="avatar">
          <img
              v-if="avatarUrl"
              :src="avatarUrl"
              alt="微信头像"
              class="avatar__img"
          />
        </div>

        <div class="profile-info">
          <div class="nickname">{{ nickName || '昵称' }}</div>
          <div class="wechat-id">微信账号</div>
        </div>
      </div>


      <!-- 表单区域 -->
      <div class="form-area">

        <!-- 所在地区：省 / 市（都可空） -->
        <div class="form-group">
          <label class="form-label required">所在地区</label>
          <div class="form-input-row">
            <!-- 省份选择 -->
            <select
                v-model="selectedProvince"
                class="form-select"
            >
              <option value="">请选择省份</option>
              <option
                  v-for="prov in provinces"
                  :key="prov.name"
                  :value="prov.name"
              >
                {{ prov.name }}
              </option>
            </select>

            <!-- 市级选择 -->
            <select
                v-model="selectedCity"
                class="form-select form-select--mini"
                :disabled="!selectedProvince"
            >
              <option value="">请选择市级</option>
              <option
                  v-for="city in currentCities"
                  :key="city"
                  :value="city"
              >
                {{ city }}
              </option>
            </select>
          </div>
          <p v-if="locationError" class="form-error">{{ locationError }}</p>
        </div>

        <!-- 学校名称 -->
        <div class="form-group">
          <label class="form-label">学校名称（非必填，建议填写）</label>
          <input
              type="text"
              class="form-input"
              placeholder="填写所在学校名称"
          />
        </div>

        <!-- 学号 -->
        <div class="form-group">
          <label class="form-label form-label--highlight">学号（非必填，建议填写）</label>
          <input
              type="text"
              class="form-input"
              placeholder="填写所属学号"
          />
        </div>

        <!-- 家长手机号：非必填 -->
        <div class="form-group">
          <label class="form-label">家长手机号（非必填，建议填写）</label>
          <input
              type="tel"
              class="form-input"
              placeholder="填写家长常用手机号"
          />
        </div>

        <!-- 不再提醒 -->
        <label class="checkbox-line">
          <input type="checkbox" v-model="dontRemind"/>
          <span>下次不再提醒</span>
        </label>
      </div>

      <!-- 底部按钮 -->
      <div class="dialog-footer">
        <button class="btn-confirm" @click="handleConfirm('confirm')">确定</button>
        <button class="btn-skip" @click="handleConfirm('skip')">跳过</button>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'
import {chinaProvinces} from '@/views/components/chinaRegions'   // ✅ 省市数据文件，下面给
import {useAuthStore} from '@/controller/wx_auth'
import {apiRequest} from "@/api";

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'never-remind'): void
}>()

const dontRemind = ref(false)
const authStore = useAuthStore()
const avatarUrl = computed(() => authStore.signInStatus.avatar_url || '')
const nickName = computed(() => authStore.signInStatus.nick_name || '')

// 省市选择状态（可为空）
const provinces = chinaProvinces
const selectedProvince = ref<string>('')
const selectedCity = ref<string>('')

// ✅ 额外表单字段
const schoolName = ref('')
const studentId = ref('')
const parentPhone = ref('')

// ✅ 所在地区的错误提示
const locationError = ref('')

watch(selectedProvince, () => {
  selectedCity.value = ''
})

const currentCities = computed(() => {
  const prov = provinces.find(p => p.name === selectedProvince.value)
  return prov ? prov.cities : []
})


async function handleConfirm(action: 'confirm' | 'skip' = 'confirm') {
  // 如果是“确定”，需要校验 + 提交
  if (action === 'confirm') {
    // 1) 校验省市必选
    if (!selectedProvince.value || !selectedCity.value) {
      locationError.value = '请选择所在地区和市级'
      return
    }
    locationError.value = ''

    try {
      await apiRequest('/api/user/basic_info', {
        method: 'POST',
        body: {
          province: selectedProvince.value,
          city: selectedCity.value,
          school_name: schoolName.value || undefined,
          student_id: studentId.value || undefined,
          parent_phone: parentPhone.value || undefined,
        },
      })
    } catch (e) {
      console.error('[NewUserInfoDialog] 提交基础信息失败', e)
      return
    }
  }

  // 无论“确定”还是“跳过”，都按勾选状态决定是否不再提醒
  if (dontRemind.value) {
    emit('never-remind')
  }
  emit('update:open', false)
}

</script>

<style scoped>
/* 背景蒙层 */
.new-user-dialog-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2100;
}

/* 主体区域 */
.new-user-dialog {
  width: 480px;
  border-radius: 18px;
  background: #fff;
  padding: 28px 32px 30px;
  box-shadow: 0 20px 55px rgba(15, 23, 42, 0.25);
  position: relative;
}

/* 顶部标题栏 */
.dialog-header {
  display: flex;
  justify-content: center;
  position: relative;
  margin-bottom: 10px;
}

.dialog-title {
  font-size: 20px;
  font-weight: 600;
  color: #111;
}

.dialog-close {
  position: absolute;
  right: 0;
  top: -4px;
  border: none;
  background: none;
  font-size: 20px;
  cursor: pointer;
  color: #777;
}

/* 头像区域 */
.profile-box {
  background: #f5f3fb;
  border-radius: 16px;
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 22px;
}

.avatar {
  width: 52px;
  height: 52px;
  background: #e0dcf4;
  border-radius: 50%;
}

.profile-info {
  display: flex;
  flex-direction: column;
}

.nickname {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.wechat-id {
  font-size: 13px;
  color: #999;
}

/* 输入区域 */
.form-area {
  display: flex;
  flex-direction: column;
  gap: 18px;
  margin-bottom: 26px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 14px;
  color: #555;
}

.form-label--highlight {
  color: #d9534f;
}

/* 输入框 / 选择框样式 */
.form-input {
  width: 100%;
  height: 40px;
  border-radius: 10px;
  border: 1px solid #ddd;
  padding: 0 12px;
  font-size: 14px;
  outline: none;
}

.form-input-row {
  display: flex;
  gap: 10px;
}

.form-select {
  width: 100%;
  height: 40px;
  border-radius: 10px;
  border: 1px solid #ddd;
  padding: 0 10px;
  font-size: 14px;
  outline: none;
  appearance: none;
  background-size: 6px 6px, 6px 6px;
  background: #fff linear-gradient(45deg, transparent 50%, #c0c4cc 50%),
  linear-gradient(135deg, #c0c4cc 50%, transparent 50%) no-repeat calc(100% - 15px) 50%, calc(100% - 10px) 50%;
}

.form-select:disabled {
  background-color: #f5f5f5;
  color: #aaa;
  cursor: not-allowed;
}

.form-select--mini {
  width: 120px;
}

.form-input:focus,
.form-select:focus {
  border-color: #6c5ce7;
}

/* 不再提醒 */
.checkbox-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
  font-size: 13px;
  color: #666;
}

/* 底部按钮 */
.dialog-footer {
  display: flex;
  justify-content: center;
  gap: 16px;
}

.btn-confirm {
  background: #7167ff;
  color: white;
  padding: 8px 26px;
  border-radius: 30px;
  border: none;
  cursor: pointer;
  font-size: 15px;
}

.btn-skip {
  background: none;
  color: #999;
  padding: 8px 26px;
  border-radius: 30px;
  border: 1px solid #ccc;
  cursor: pointer;
  font-size: 15px;
}

.btn-confirm:hover {
  background: #6358ff;
}

.btn-skip:hover {
  border-color: #aaa;
  color: #666;
}

.avatar {
  width: 52px;
  height: 52px;
  background: #e0dcf4;
  border-radius: 50%;
}

/* ✅ 新增：让头像图铺满圆形区域 */
.avatar__img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  display: block;
}

</style>
