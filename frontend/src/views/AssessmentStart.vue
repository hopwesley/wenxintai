<template>
  <main class="page">
    <section class="card">
      <h1>两阶段评测流程</h1>
      <p class="lead">填写基本信息后即可开始 S1 题集作答，完成后系统将自动下发 S2 并生成报告。</p>

      <form class="form" @submit.prevent="handleSubmit">
        <label class="field">
          <span>评测模式</span>
          <select v-model="form.mode" required>
            <option value="standard">standard</option>
            <option value="advanced">advanced</option>
          </select>
        </label>

        <label class="field">
          <span>邀请码</span>
          <input v-model.trim="form.inviteCode" type="text" placeholder="可选" />
        </label>

        <label class="field">
          <span>微信 OpenID</span>
          <input v-model.trim="form.wechatOpenId" type="text" placeholder="可选" />
        </label>

        <p class="hint">邀请码与微信 OpenID 至少填写一项。</p>

        <button type="submit" :disabled="submitting">
          {{ submitting ? '创建中…' : '开始评测' }}
        </button>
      </form>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </section>

    <section v-if="resumeState.assessmentId" class="card">
      <h2>已存在的评测</h2>
      <p class="status">评测 ID：{{ resumeState.assessmentId }}</p>
      <p v-if="progress" class="status">当前状态：{{ progress.label }} ({{ progress.status }})</p>
      <div class="actions">
        <button type="button" @click="resume" :disabled="!canResume">
          继续答题
        </button>
        <button type="button" @click="viewReport" :disabled="!canViewReport">
          查看报告
        </button>
        <button type="button" class="secondary" @click="resetLocalState">
          清除本地缓存
        </button>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { createAssessment, getProgress, type ProgressResponse } from '@/api/assessment'
import {
  getAssessmentFlowState,
  setAssessmentId,
  setQuestionSet,
  clearAll,
  type StageKey
} from '@/store/assessmentFlow'

const router = useRouter()

const form = reactive({
  mode: 'standard',
  inviteCode: '',
  wechatOpenId: ''
})

const submitting = ref(false)
const errorMessage = ref('')
const resumeState = ref(getAssessmentFlowState())
const progress = ref<ProgressResponse | null>(null)

const canResume = computed(() => Boolean(resumeState.value.activeQuestionSetId))
const canViewReport = computed(() => progress.value?.label === 'REPORT_READY')

onMounted(async () => {
  if (resumeState.value.assessmentId) {
    try {
      progress.value = await getProgress(resumeState.value.assessmentId)
    } catch (error) {
      console.warn('[AssessmentStart] failed to load progress', error)
    }
  }
})

function normalizeOptional(value: string) {
  const trimmed = value.trim()
  return trimmed.length > 0 ? trimmed : undefined
}

async function handleSubmit() {
  errorMessage.value = ''
  const invite = normalizeOptional(form.inviteCode)
  const openId = normalizeOptional(form.wechatOpenId)
  if (!invite && !openId) {
    errorMessage.value = '请填写邀请码或微信 OpenID'
    return
  }
  submitting.value = true
  try {
    const response = await createAssessment({
      mode: form.mode,
      invite_code: invite,
      wechat_openid: openId
    })
    setAssessmentId(response.assessment_id)
    setQuestionSet(response.question_set_id, response.stage as StageKey, response.questions)
    resumeState.value = getAssessmentFlowState()
    progress.value = await getProgress(response.assessment_id).catch(() => null)
    router.push({ name: 'assessment-questions', params: { questionSetId: response.question_set_id } })
  } catch (error) {
    console.error('[AssessmentStart] failed to create assessment', error)
    if (error instanceof Error) {
      errorMessage.value = error.message
    } else {
      errorMessage.value = '创建评测失败，请稍后重试'
    }
  } finally {
    submitting.value = false
  }
}

function resume() {
  if (resumeState.value.activeQuestionSetId) {
    router.push({ name: 'assessment-questions', params: { questionSetId: resumeState.value.activeQuestionSetId } })
  }
}

function viewReport() {
  const id = resumeState.value.assessmentId
  if (id) {
    router.push({ name: 'assessment-report', params: { assessmentId: id } })
  }
}

function resetLocalState() {
  clearAll()
  resumeState.value = getAssessmentFlowState()
  progress.value = null
}
</script>

<style scoped>
.page {
  margin: 0 auto;
  max-width: 680px;
  padding: 32px 16px 64px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.card {
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.08);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.card h1,
.card h2 {
  margin: 0;
  color: #111827;
}

.lead {
  color: #4b5563;
  margin: 0;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  color: #1f2937;
  font-size: 14px;
}

input,
select,
button {
  font: inherit;
}

input,
select {
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.6);
}

button {
  padding: 10px 16px;
  border-radius: 8px;
  border: none;
  background: #2563eb;
  color: #fff;
  cursor: pointer;
}

button[disabled] {
  opacity: 0.6;
  cursor: not-allowed;
}

.secondary {
  background: #e5e7eb;
  color: #111827;
}

.hint {
  margin: 0;
  font-size: 12px;
  color: #6b7280;
}

.error {
  color: #dc2626;
  margin: 0;
}

.status {
  margin: 0;
  color: #374151;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
</style>
