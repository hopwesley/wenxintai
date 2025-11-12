<template>
  <main class="config-page">
    <section class="config-card">
      <header class="config-header">
        <p class="config-badge">邀请码：{{ inviteCode }}</p>
        <h1>完善测试设置</h1>
        <p class="config-desc">请选择年级与测试模式，我们会基于您的选择生成专属题目。</p>
      </header>

      <form class="config-form" @submit.prevent="handleSubmit">
        <label class="config-field">
          <span>年级</span>
          <input
            v-model.trim="form.grade"
            type="text"
            placeholder="例如：高一"
            :disabled="submitting"
            required
          />
        </label>

        <fieldset class="config-field">
          <legend>测试模式</legend>
          <div class="mode-options">
            <label v-for="option in modeOptions" :key="option.value" class="mode-option">
              <input
                type="radio"
                name="mode"
                :value="option.value"
                v-model="form.mode"
                :disabled="submitting"
              />
              <span>{{ option.label }}</span>
            </label>
          </div>
        </fieldset>

        <label class="config-field">
          <span>兴趣偏好（可选）</span>
          <input
            v-model.trim="form.interest"
            type="text"
            :list="hobbyListId"
            placeholder="如：艺术、物理"
            :disabled="submitting"
          />
          <datalist v-if="hobbies.length" :id="hobbyListId">
            <option v-for="item in hobbies" :key="item" :value="item" />
          </datalist>
        </label>

        <p v-if="errorMessage" class="config-error">{{ errorMessage }}</p>

        <button class="config-submit" type="submit" :disabled="!canSubmit || submitting">
          <span v-if="submitting">创建中…</span>
          <span v-else>下一步</span>
        </button>
      </form>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { createAssessment } from '@/api/assessment'
import { getHobbies } from '@/api'
import { useTestSession, type ModeOption } from '@/store/testSession'
import { setAssessmentId, setQuestionSet, type StageKey } from '@/store/assessmentFlow'

interface TestConfigForm {
  grade: string
  mode: ModeOption | ''
  interest: string
}

const router = useRouter()
const { state, setTestConfig, setInviteCode } = useTestSession()

const form = reactive<TestConfigForm>({
  grade: state.grade ?? '',
  mode: state.mode ?? '',
  interest: state.interest ?? ''
})

const hobbies = ref<string[]>([])
const hobbyListId = 'hobby-options'
const errorMessage = ref('')
const submitting = ref(false)

const inviteCode = computed(() => state.inviteCode ?? '')

const selectedMode = computed<ModeOption | null>(() => {
  return form.mode === '3+3' || form.mode === '3+1+2' ? form.mode : null
})

const canSubmit = computed(() => {
  return Boolean(inviteCode.value && form.grade.trim() && selectedMode.value)
})

const modeOptions = [
  { label: '3+3', value: '3+3' as ModeOption },
  { label: '3+1+2', value: '3+1+2' as ModeOption }
]

onMounted(async () => {
  if (!inviteCode.value) {
    router.replace('/')
    return
  }

  if (!form.mode) {
    form.mode = '3+3'
  }

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
    router.replace('/')
    return
  }

  if (!selectedMode.value) {
    errorMessage.value = '请选择测试模式'
    return
  }

  if (!form.grade.trim()) {
    errorMessage.value = '请输入年级'
    return
  }

  errorMessage.value = ''
  submitting.value = true

  try {
    const grade = form.grade.trim()
    const interest = form.interest.trim()
    const response = await createAssessment({
      invite_code: inviteCode.value,
      mode: selectedMode.value,
      grade
    })

    const questionSetId = response.active_question_set_id ?? response.question_set_id
    if (!questionSetId) {
      throw new Error('未获取到题集信息')
    }

    setTestConfig({
      grade,
      mode: selectedMode.value,
      interest: interest || undefined
    })
    setInviteCode(inviteCode.value)

    setAssessmentId(response.assessment_id)
    setQuestionSet(questionSetId, response.stage as StageKey, response.questions)

    await router.push(`/questions/${encodeURIComponent(questionSetId)}`)
  } catch (error) {
    console.error('[StartTestConfig] failed to create assessment', error)
    if (error instanceof Error) {
      errorMessage.value = error.message
    } else {
      errorMessage.value = '创建评测失败，请稍后重试'
    }
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

.mode-options {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.mode-option {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.4);
  cursor: pointer;
  transition: border-color 0.2s ease, background-color 0.2s ease;
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
