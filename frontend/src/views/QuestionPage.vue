<template>
  <main class="page">
    <section class="card" v-if="questionSet">
      <header class="header">
        <div>
          <h1>当前阶段：{{ questionSet.stage }}</h1>
          <p v-if="progress" class="muted">评测状态：{{ progress.label }} ({{ progress.status }})</p>
        </div>
        <button type="button" class="secondary" @click="goHome">返回首页</button>
      </header>

      <div class="questions">
        <article v-for="item in questionItems" :key="item.id" class="question">
          <h2>{{ item.title }}</h2>
          <p class="desc">{{ item.description }}</p>
          <textarea v-model.trim="answers[item.id]" rows="3" placeholder="请输入答案"></textarea>
        </article>
      </div>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

      <footer class="actions">
        <button type="button" @click="submit" :disabled="submitting">
          {{ submitting ? '提交中…' : '提交答案' }}
        </button>
      </footer>
    </section>

    <section v-else class="card">
      <h1>未找到题集</h1>
      <p>当前题集已过期或本地缓存被清除，请返回首页重新创建评测。</p>
      <button type="button" @click="goHome">返回首页</button>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getProgress, submitAnswers, type ProgressResponse, type SubmitAnswersResponse, type SubmitAnswersResponseStage2 } from '@/api/assessment'
import {
  getAssessmentFlowState,
  getQuestionSet,
  setQuestionSet,
  recordReport,
  setActiveQuestionSet,
  type StageKey
} from '@/store/assessmentFlow'

interface QuestionItem {
  id: string
  title: string
  description: string
}

const route = useRoute()
const router = useRouter()

const state = ref(getAssessmentFlowState())
const currentQuestionSetId = ref<string | null>(null)
const progress = ref<ProgressResponse | null>(null)
const errorMessage = ref('')
const submitting = ref(false)
const answers = reactive<Record<string, string>>({})

const questionSet = computed(() => {
  if (!currentQuestionSetId.value) return undefined
  return getQuestionSet(currentQuestionSetId.value)
})

const questionItems = computed<QuestionItem[]>(() => {
  const qs = questionSet.value
  if (!qs) return []
  const raw = Array.isArray(qs.questions) ? qs.questions : []
  return raw.map((item: any, index: number) => {
    const id = typeof item?.id === 'string' && item.id.trim() ? item.id : `q${index + 1}`
    const title = item?.prompt ?? item?.title ?? `问题 ${index + 1}`
    const description = item?.description ?? item?.text ?? ''
    return { id, title, description }
  })
})

watch(
  () => route.params.questionSetId,
  (value) => {
    if (typeof value === 'string' && value) {
      currentQuestionSetId.value = value
      setActiveQuestionSet(value)
      resetAnswers()
    }
  },
  { immediate: true }
)

onMounted(async () => {
  const assessmentId = state.value.assessmentId
  if (!assessmentId) {
    return
  }
  try {
    progress.value = await getProgress(assessmentId)
  } catch (error) {
    console.warn('[QuestionPage] failed to fetch progress', error)
  }
})

function resetAnswers() {
  Object.keys(answers).forEach((key) => delete answers[key])
  questionItems.value.forEach((item) => {
    answers[item.id] = ''
  })
  errorMessage.value = ''
}

async function submit() {
  if (!currentQuestionSetId.value) {
    return
  }
  errorMessage.value = ''
  const payload = questionItems.value.map((item) => ({
    question_id: item.id,
    answer: answers[item.id]?.trim() ?? ''
  }))

  const incomplete = payload.some((entry) => !entry.answer)
  if (incomplete) {
    errorMessage.value = '请完成所有题目后再提交'
    return
  }

  submitting.value = true
  try {
    const response = await submitAnswers(currentQuestionSetId.value, payload)
    await handleSubmitResult(response)
  } catch (error) {
    console.error('[QuestionPage] submit error', error)
    if (error instanceof Error) {
      errorMessage.value = error.message
    } else {
      errorMessage.value = '提交失败，请稍后重试'
    }
  } finally {
    submitting.value = false
  }
}

async function handleSubmitResult(result: SubmitAnswersResponse) {
  if ('next_question_set_id' in result) {
    setQuestionSet(result.next_question_set_id, result.stage as StageKey, result.questions)
    state.value = getAssessmentFlowState()
    router.replace({ name: 'assessment-questions', params: { questionSetId: result.next_question_set_id } })
    return
  }

  const stage2 = result as SubmitAnswersResponseStage2
  if (stage2.report_id) {
    recordReport(stage2.report_id)
  }
  state.value = getAssessmentFlowState()
  router.push({ name: 'assessment-report', params: { assessmentId: stage2.assessment_id } })
}

function goHome() {
  router.push({ name: 'assessment-start' })
}
</script>

<style scoped>
.page {
  margin: 0 auto;
  max-width: 800px;
  padding: 32px 16px 64px;
}

.card {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.08);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.header h1 {
  margin: 0;
  color: #111827;
}

.muted {
  margin: 4px 0 0;
  color: #6b7280;
}

.questions {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.question {
  border: 1px solid rgba(148, 163, 184, 0.4);
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.question h2 {
  margin: 0;
  font-size: 18px;
  color: #1f2937;
}

.question .desc {
  margin: 0;
  color: #4b5563;
  font-size: 14px;
}

textarea {
  font: inherit;
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.6);
  resize: vertical;
}

.actions {
  display: flex;
  justify-content: flex-end;
}

button {
  font: inherit;
  padding: 10px 18px;
  border-radius: 8px;
  border: none;
  background: #2563eb;
  color: #fff;
  cursor: pointer;
}

button.secondary {
  background: #e5e7eb;
  color: #111827;
}

button[disabled] {
  opacity: 0.6;
  cursor: not-allowed;
}

.error {
  color: #dc2626;
  margin: 0;
}
</style>
