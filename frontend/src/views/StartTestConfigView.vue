<template>
  <TestLayout>
    <template #header>
      <StepIndicator :steps="stepItems" :current="1"/>
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
import { computed, onMounted, reactive, ref } from 'vue'
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/components/StepIndicator.vue'
import { STEPS, type Variant } from '@/config/testSteps'
import {useRouter} from 'vue-router'
import {createAssessment} from '@/api/assessment'   // ← 修正导入路径
import {useTestSession, type ModeOption} from '@/store/testSession'
import {setAssessmentId, setQuestionSet, type StageKey} from '@/store/assessmentFlow'
import {getHobbies} from "@/api";


interface TestConfigForm {
  grade: string
  mode: ModeOption | ''
  hobby: string
}

const router = useRouter()
const {state, setTestConfig, setInviteCode} = useTestSession()

const stepItems = computed(() => {
  const arr = (STEPS as Record<Variant, readonly { key: string; titleKey?: string }[]>)[state.variant] ?? []
  return arr.map(it => ({ key: it.key, title: it.titleKey ?? it.key }))
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
  if (!inviteCode.value) {
    router.replace('/')
    return
  }

  // 不要给默认模式，保持必选
  try {
    const list = await getHobbies()      // 约定返回 string[]
    hobbies.value = Array.isArray(list) ? list.map(String) : []
  } catch (error) {
    console.warn('[StartTestConfig] failed to load hobbies', error)
    hobbies.value = []
  }
})

async function handleSubmit() {
  if (!inviteCode.value) {
    router.replace('/')
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
    const interest = form.hobby.trim()
    const response = await createAssessment({
      invite_code: inviteCode.value,
      mode: selectedMode.value,
      grade
    })

    const questionSetId = response.active_question_set_id ?? response.question_set_id
    if (!questionSetId) {
      errorMessage.value = '未获取到题集信息'
      return
    }

    setTestConfig({grade, mode: selectedMode.value, hobby: interest || undefined})
    setInviteCode(inviteCode.value)

    setAssessmentId(response.assessment_id)
    setQuestionSet(questionSetId, response.stage as StageKey, response.questions)

    await router.push({
      name: 'test-stage',
      params: { variant: state.variant || 'basic', step: 2 }
    })
  } catch (error) {
    console.error('[StartTestConfig] failed to create assessment', error)
    errorMessage.value = error instanceof Error ? error.message : '创建评测失败，请稍后重试'
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
