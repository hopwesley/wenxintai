<template>
  <TestLayout>
    <template #header>
      <StepIndicator :steps="stepItems" :current="currentStepIndex"/>
    </template>
    <main class="config-page">
      <section class="config-card">
        <header class="config-header">
          <p class="config-badge">邀请码：{{ inviteCode }}</p>
          <h1>完善测试设置</h1>
          <p class="config-desc">请选择年级与测试模式，我们会基于您的选择生成专属题目。</p>
        </header>

        <form class="config-form" @submit.prevent="handleSubmit">
          <!-- 年级：下拉 -->
          <label class="config-field">
            <span>年级</span>
            <select v-model="form.grade" :disabled="submitting" required>
              <option value="">请选择年级</option>
              <option value="初二">初二</option>
              <option value="初三">初三</option>
              <option value="高一">高一</option>
            </select>
          </label>

          <!-- 模式：下拉（必选） -->
          <label class="config-field">
            <span>测试模式</span>
            <select v-model="form.mode" :disabled="submitting" required>
              <option value="">请选择模式</option>
              <option value="3+3">3+3 模式</option>
              <option value="3+1+2">3+1+2 模式</option>
            </select>
          </label>

          <!-- 兴趣：下拉（可选，来自后端） -->
          <label class="config-field">
            <span>兴趣偏好（可选）</span>
            <select v-model="form.hobby" :disabled="submitting">
              <option value="">请选择爱好</option>
              <option v-for="h in hobbies" :key="h" :value="h">{{ h }}</option>
            </select>
          </label>

          <p v-if="errorMessage" class="config-error">{{ errorMessage }}</p>

          <button class="config-submit" type="submit" :disabled="!canSubmit || submitting">
            <span v-if="submitting">创建中…</span>
            <span v-else>下一步</span>
          </button>
        </form>
      </section>
    </main>
  </TestLayout>
</template>

<script setup lang="ts">
import {computed, onMounted, reactive, ref} from 'vue'
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/components/StepIndicator.vue'
import {useRouter} from 'vue-router'
import {useTestSession, type ModeOption} from '@/store/testSession'
import {getHobbies} from "@/api";
import {useAlert} from '@/logic/useAlert'

interface TestConfigForm {
  grade: string
  mode: ModeOption | ''
  hobby: string
}

const {showAlert} = useAlert()
const router = useRouter()
const {state, setTestConfig, setInviteCode} = useTestSession()

function handleFlowError(msg?: string) {
  showAlert(msg ?? '测试流程异常，请返回首页重新开始', () => {
    router.replace('/')
  })
}


const stepItems = computed(() => {
  const routes = state.testRoutes ?? []
  return routes.map(r => ({
    key: r.router,   // 英文路由名
    title: r.desc,   // 中文描述
  }))
})


const currentStepIndex = computed(() => {
  const routes = state.testRoutes ?? []
  const idx = routes.findIndex(r => r.router === 'basic-info')
  return idx >= 0 ? idx + 1 : 0
})


const form = reactive<TestConfigForm>({
  grade: state.grade ?? '',
  mode: state.mode ?? '',         // 默认空，强制用户选择
  hobby: state.hobby ?? ''
})

const hobbies = ref<string[]>([]) // ← 仅保留这一处定义
const errorMessage = ref('')
const submitting = ref(false)
const inviteCode = computed(() => state.inviteCode ?? '')

const selectedMode = computed<ModeOption | null>(() => {
  return form.mode === '3+3' || form.mode === '3+1+2' ? form.mode : null
})

const canSubmit = computed(() => {
  return Boolean(inviteCode.value && form.grade.trim() && selectedMode.value)
})

onMounted(async () => {
  // 没有邀请码：直接弹窗 + 回首页
  if (!inviteCode.value) {
    handleFlowError('未找到邀请码，请返回首页重新开始')
    return
  }

  // 检查 testRoutes 是否存在、且包含 basic-info
  const routes = state.testRoutes ?? []
  if (!routes.length) {
    handleFlowError('测试流程异常，未找到测试流程，请返回首页重新开始')
    return
  }
  const idx = routes.findIndex(r => r.router === 'basic-info')
  if (idx < 0) {
    handleFlowError('测试流程异常，未找到 basic-info 步骤，请返回首页重新开始')
    return
  }

  // 正常情况：拉兴趣爱好列表
  try {
    const list = await getHobbies()
    hobbies.value = Array.isArray(list) ? list.map(String) : []
  } catch (error) {
    console.warn('[StartTestConfig] failed to load hobbies', error)
    hobbies.value = []
  }
})


async function handleSubmit() {
  if (!inviteCode.value) {
    // 理论上不会走到这里，兜底一下
    handleFlowError('未找到邀请码，请返回首页重新开始')
    return
  }
  if (!selectedMode.value) {
    errorMessage.value = '请选择测试模式'
    return
  }
  if (!form.grade.trim()) {
    errorMessage.value = '请选择年级'
    return
  }

  errorMessage.value = ''
  submitting.value = true

  try {
    const grade = form.grade.trim()
    const hobby = form.hobby.trim()

    // 写入测试配置到 TestSession
    setTestConfig({
      grade,
      mode: selectedMode.value as ModeOption,
      hobby: hobby || undefined,
    })
    setInviteCode(inviteCode.value)

    // 从 testRoutes 中找到 basic-info 的下一步
    const routes = state.testRoutes ?? []
    const idx = routes.findIndex(r => r.router === 'basic-info')
    if (idx < 0 || idx === routes.length - 1) {
      handleFlowError('测试流程异常，未找到下一步，请返回首页重新开始')
      return
    }

    const next = routes[idx + 1]
    const typ = state.testType || 'basic'

    await router.push(`/test/${typ}/${next.router}`)
  } catch (err) {
    console.error('[StartTestConfig] handleSubmit error', err)
    handleFlowError(
        (err as Error)?.message || '测试流程异常，请返回首页重新开始'
    )
  } finally {
    submitting.value = false
  }
}

</script>

<style scoped>
.config-page {
  margin: 0 auto;
  max-width: 640px;
  padding: 48px 16px 64px;
}

.config-card {
  background: #ffffff;
  border-radius: 20px;
  box-shadow: 0 16px 40px rgba(15, 23, 42, 0.08);
  padding: 32px 28px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.config-header {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.config-badge {
  align-self: flex-start;
  background: rgba(91, 124, 255, 0.12);
  color: #3949ab;
  border-radius: 999px;
  padding: 4px 12px;
  font-size: 13px;
  font-weight: 600;
  margin: 0;
}

.config-header h1 {
  margin: 0;
  font-size: 24px;
  color: #111827;
}

.config-desc {
  margin: 0;
  color: #4b5563;
  line-height: 1.6;
}

.config-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.config-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  color: #1f2937;
  font-size: 14px;
}

.config-field input,
.config-field select {
  font: inherit;
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.4);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.config-field input:focus,
.config-field select:focus {
  outline: none;
  border-color: #5b7cff;
  box-shadow: 0 0 0 3px rgba(91, 124, 255, 0.2);
}

.mode-option input {
  margin: 0;
}

.mode-option input:checked + span,
.mode-option:hover span {
  color: #1f2937;
  font-weight: 600;
}

.mode-option input:checked ~ span,
.mode-option input:checked ~ span {
  color: #1f2937;
}

.config-error {
  margin: -4px 0 0;
  color: #ef4444;
  font-size: 13px;
}

.config-submit {
  align-self: flex-end;
  min-width: 120px;
  padding: 10px 24px;
  border-radius: 999px;
  border: none;
  background: linear-gradient(135deg, #5b7cff, #6366f1);
  color: #fff;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.config-submit:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
</style>
