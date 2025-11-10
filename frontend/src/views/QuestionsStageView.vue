<template>
  <TestLayout>
    <template #header>
      <StepIndicator :steps="stepItems" :current="state.currentStep" />
    </template>

    <section class="questions">
      <header class="questions__header">
        <h1>{{ currentStepTitle }}</h1>
        <p class="questions__progress" v-if="totalPages > 1">
          {{ currentPage }} / {{ totalPages }}
        </p>
      </header>

      <div v-if="loading" class="questions__loading">{{ t('loading.default') }}</div>
      <div v-else-if="errorMessage" class="questions__error">{{ errorMessage }}</div>
      <div v-else>
        <form @submit.prevent="handleNext">
          <div
            v-for="question in currentPageQuestions"
            :key="question.id"
            :class="['question', { 'question--highlight': highlightedQuestionId === question.id }]"
            :ref="(el) => registerQuestionRef(question.id, el)"
          >
            <p class="question__text">{{ question.text }}</p>
            <div class="question__options" role="radiogroup">
              <label
                v-for="option in scaleOptions"
                :key="option.value"
                class="question__option"
              >
                <input
                  type="radio"
                  :name="`question-${question.id}`"
                  :value="option.value"
                  :checked="getAnswer(question.id) === option.value"
                  @change="onSelect(question.id, option.value)"
                />
                <span>{{ option.label }}</span>
              </label>
            </div>
          </div>
        </form>
      </div>

      <footer class="questions__footer" v-if="!loading && !errorMessage">
        <button type="button" class="questions__nav questions__nav--prev" @click="handlePrev">
          {{ t('btn.prev') }}
        </button>
        <button
          type="button"
          class="questions__nav questions__nav--next"
          :disabled="!isCurrentPageComplete || submitting"
          @click="handleNext"
        >
          <span v-if="submitting">{{ t('loading.submitting') }}</span>
          <span v-else>{{ currentNextLabel }}</span>
        </button>
      </footer>
    </section>

    <template #footer>
      <p>{{ t('disclaimer') }}</p>
    </template>
  </TestLayout>
</template>

<script setup lang="ts">
import { computed, ref, watch, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import TestLayout from '@/layouts/TestLayout.vue'
import StepIndicator from '@/components/StepIndicator.vue'
import { useI18n } from '@/i18n'
import { getQuestions, submitTestSession } from '@/api'
import { useTestSession } from '@/store/testSession'
import { STEPS, isVariant, type Variant } from '@/config/testSteps'

interface Question {
  id: string
  text: string
}

interface StageQuestions {
  stage1: Question[]
  stage2: Question[]
}

const props = withDefaults(defineProps<{ stage: 1 | 2; pageSize?: number }>(), {
  pageSize: 5,
})

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { state, ensureSessionId, setVariant, setCurrentStep, setAnswer, isPageComplete, nextStep, prevStep, toPayload } =
  useTestSession()

const variant = ref<Variant>('basic')
const loading = ref(true)
const submitting = ref(false)
const errorMessage = ref('')
const questions = ref<Question[]>([])
const currentPage = ref(1)
const highlightedQuestionId = ref<string | null>(null)

const questionRefs = new Map<string, HTMLElement>()

const scaleOptions = computed(() => [
  { value: 1 as const, label: t('scale.never') },
  { value: 2 as const, label: t('scale.rare') },
  { value: 3 as const, label: t('scale.normal') },
  { value: 4 as const, label: t('scale.often') },
  { value: 5 as const, label: t('scale.alot') },
])

const stepItems = computed(() =>
  STEPS[variant.value].map((item) => ({ key: item.key, title: t(item.titleKey) }))
)

const totalPages = computed(() => Math.max(1, Math.ceil(questions.value.length / props.pageSize)))

const currentPageQuestions = computed(() => {
  const start = (currentPage.value - 1) * props.pageSize
  return questions.value.slice(start, start + props.pageSize)
})

const currentPageQuestionIds = computed(() => currentPageQuestions.value.map((q) => q.id))

const isCurrentPageComplete = computed(() =>
  isPageComplete(props.stage, currentPageQuestionIds.value)
)

const currentStepTitle = computed(() => {
  const currentStepIndex = state.currentStep - 1
  const item = STEPS[variant.value][currentStepIndex]
  return item ? t(item.titleKey) : ''
})

const currentNextLabel = computed(() => {
  if (currentPage.value === totalPages.value && props.stage === 2) {
    return t('btn.submit')
  }
  return t('btn.next')
})

watchEffect(() => {
  const variantParam = String(route.params.variant ?? 'basic')
  if (!isVariant(variantParam)) {
    router.replace({ path: '/test/basic/step/1' })
    return
  }
  variant.value = variantParam
  setVariant(variant.value)
})

watchEffect(() => {
  const stepNumber = Number(route.params.step ?? '2')
  if (Number.isNaN(stepNumber) || stepNumber < 1 || stepNumber > STEPS[variant.value].length) {
    router.replace({ path: `/test/${variant.value}/step/1` })
    return
  }
  setCurrentStep(stepNumber)
})

watch(
  [variant, () => props.stage],
  ([nextVariant, nextStage], [prevVariant, prevStage]) => {
    if (!prevVariant || nextVariant !== prevVariant) {
      cachedQuestions.value = null
      currentPage.value = 1
    }
    if (nextStage !== prevStage) {
      currentPage.value = 1
    }
    highlightedQuestionId.value = null
    loadQuestions()
  },
  { immediate: true }
)

function registerQuestionRef(id: string, element: Element | null) {
  if (!element) {
    questionRefs.delete(id)
    return
  }
  questionRefs.set(id, element as HTMLElement)
}

function highlightQuestion(questionId: string) {
  highlightedQuestionId.value = questionId
  const el = questionRefs.get(questionId)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }
  setTimeout(() => {
    if (highlightedQuestionId.value === questionId) {
      highlightedQuestionId.value = null
    }
  }, 1500)
}

async function loadQuestions() {
  if (!state.mode || !state.hobby || state.age == null) {
    router.replace({ path: `/test/${variant.value}/step/1` })
    return
  }

  if (props.stage === 2 && Object.keys(state.answersStage1).length === 0) {
    router.replace({ path: `/test/${variant.value}/step/2` })
    return
  }

  loading.value = true
  errorMessage.value = ''

  try {
    questions.value = await ensureStageQuestions(props.stage)
    if (!questions.value.length) {
      errorMessage.value = t('error.noQuestions')
    }
    questionRefs.clear()
    if (currentPage.value > totalPages.value) {
      currentPage.value = totalPages.value
    }
  } catch (error) {
    console.error('[QuestionsStageView] failed to load questions', error)
    errorMessage.value = t('error.network')
  } finally {
    loading.value = false
  }
}

function onSelect(questionId: string, value: 1 | 2 | 3 | 4 | 5) {
  setAnswer(props.stage, questionId, value)
}

function getAnswer(questionId: string) {
  return getAnswerFromStore(props.stage, questionId)
}

function getAnswerFromStore(stage: 1 | 2, questionId: string) {
  return state[stage === 1 ? 'answersStage1' : 'answersStage2'][questionId]
}

function scrollToFirstUnanswered() {
  for (const question of currentPageQuestions.value) {
    if (!getAnswerFromStore(props.stage, question.id)) {
      highlightQuestion(question.id)
      break
    }
  }
}

async function handleNext() {
  if (!isCurrentPageComplete.value) {
    scrollToFirstUnanswered()
    return
  }

  if (currentPage.value < totalPages.value) {
    currentPage.value += 1
    return
  }

  if (props.stage === 1) {
    const limit = STEPS[variant.value].length
    const nextStepNumber = nextStep(limit)
    await router.push({ path: `/test/${variant.value}/step/${nextStepNumber}` })
    return
  }

  submitting.value = true
  try {
    const payload = toPayload()
    if (!payload.sessionId) {
      payload.sessionId = ensureSessionId()
    }
    if (payload.age == null || !payload.mode || !payload.hobby) {
      submitting.value = false
      await router.replace({ path: `/test/${variant.value}/step/1` })
      return
    }
    await submitTestSession({
      sessionId: payload.sessionId!,
      variant: payload.variant,
      age: payload.age!,
      mode: payload.mode!,
      hobby: payload.hobby!,
      riasec_answers: payload.answersStage1,
      asc_answers: payload.answersStage2,
    })
    const limit = STEPS[variant.value].length
    const nextStepNumber = nextStep(limit)
    await router.push({ path: `/test/${variant.value}/step/${nextStepNumber}` })
  } catch (error) {
    console.error('[QuestionsStageView] submit error', error)
    errorMessage.value = t('error.network')
  } finally {
    submitting.value = false
  }
}

async function handlePrev() {
  if (currentPage.value > 1) {
    currentPage.value -= 1
    return
  }

  const previous = prevStep()
  await router.push({ path: `/test/${variant.value}/step/${previous}` })
}

const cachedQuestions = ref<StageQuestions | null>(null)

async function ensureStageQuestions(stage: 1 | 2) {
  if (!cachedQuestions.value) {
    const sessionId = ensureSessionId()
    const response = await getQuestions({
      session_id: sessionId,
      mode: state.mode ?? '',
      gender: '',
      grade: '',
      hobby: state.hobby ?? '',
    })
    cachedQuestions.value = normalizeQuestions(response)
  }
  const key = stage === 1 ? 'stage1' : 'stage2'
  return cachedQuestions.value?.[key] ?? []
}

function normalizeQuestions(response: any): StageQuestions {
  const container = response?.questions ?? response ?? {}

  const stage1Candidates = resolveQuestionArray(container.stage1 ?? container.stage1Questions)
  const stage2Candidates = resolveQuestionArray(container.stage2 ?? container.stage2Questions)

  if (stage1Candidates.length || stage2Candidates.length) {
    return {
      stage1: stage1Candidates,
      stage2: stage2Candidates,
    }
  }

  if (Array.isArray(container.questions)) {
    const combined = normalizeQuestionList(container.questions)
    const half = Math.ceil(combined.length / 2)
    return {
      stage1: combined.slice(0, half),
      stage2: combined.slice(half),
    }
  }

  if (Array.isArray(container)) {
    const combined = normalizeQuestionList(container)
    const half = Math.ceil(combined.length / 2)
    return {
      stage1: combined.slice(0, half),
      stage2: combined.slice(half),
    }
  }

  return { stage1: [], stage2: [] }
}

function resolveQuestionArray(input: unknown): Question[] {
  if (!Array.isArray(input)) return []
  return normalizeQuestionList(input)
}

function normalizeQuestionList(list: any[]): Question[] {
  return list.map((item, index) => ({
    id: String(item?.id ?? item?.question_id ?? item?.qid ?? item?.key ?? `q${index}`),
    text: String(item?.text ?? item?.title ?? item?.question ?? ''),
  }))
}
</script>

<style scoped>
.questions {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.questions__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.questions__header h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #111827;
}

.questions__progress {
  margin: 0;
  font-size: 14px;
  color: rgba(30, 41, 59, 0.6);
}

.questions__loading,
.questions__error {
  text-align: center;
  padding: 40px 0;
  color: rgba(30, 41, 59, 0.7);
}

.questions__error {
  color: #dc2626;
}

.question {
  padding: 20px;
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.4);
  background-color: #f9fafb;
  display: flex;
  flex-direction: column;
  gap: 16px;
  transition: box-shadow 0.2s ease;
}

.question--highlight {
  box-shadow: 0 0 0 4px rgba(239, 68, 68, 0.2);
}

.question__text {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
  color: #0f172a;
}

.question__options {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.question__option {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border-radius: 999px;
  border: 1px solid rgba(99, 102, 241, 0.4);
  background: white;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
}

.question__option input {
  appearance: none;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid rgba(148, 163, 184, 0.6);
  position: relative;
}

.question__option input:checked {
  border-color: #6366f1;
}

.question__option input:checked::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #6366f1;
}

.question__option:hover,
.question__option:focus-within {
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
}

.questions__footer {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.questions__nav {
  min-width: 160px;
  padding: 12px 20px;
  border-radius: 999px;
  border: none;
  cursor: pointer;
  font-size: 16px;
  font-weight: 600;
  transition: opacity 0.2s ease;
}

.questions__nav--prev {
  background: rgba(148, 163, 184, 0.2);
  color: #1f2937;
}

.questions__nav--next {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
}

.questions__nav:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .question {
    padding: 16px;
  }

  .questions__footer {
    flex-direction: column;
  }

  .questions__nav {
    width: 100%;
  }
}
</style>
