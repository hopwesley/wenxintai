import { reactive, watch } from 'vue'
import type { Variant } from '@/config/testSteps'

const STORAGE_KEY = 'wenxintai:test-session'

type AnswerValue = 1 | 2 | 3 | 4 | 5

type ModeOption = '3+3' | '3+1+2'

export interface TestSession {
  sessionId?: string
  variant: Variant
  currentStep: number
  age?: number
  mode?: ModeOption
  hobby?: string
  answersStage1: Record<string, AnswerValue>
  answersStage2: Record<string, AnswerValue>
}

const defaultSession: TestSession = {
  sessionId: undefined,
  variant: 'basic',
  currentStep: 1,
  age: undefined,
  mode: undefined,
  hobby: undefined,
  answersStage1: {},
  answersStage2: {},
}

type SerializableTestSession = Omit<TestSession, 'answersStage1' | 'answersStage2'> & {
  answersStage1: Record<string, AnswerValue>
  answersStage2: Record<string, AnswerValue>
}

function loadFromStorage(): TestSession {
  if (typeof window === 'undefined') {
    return { ...defaultSession, answersStage1: {}, answersStage2: {} }
  }

  const raw = window.sessionStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return { ...defaultSession, answersStage1: {}, answersStage2: {} }
  }

  try {
    const parsed = JSON.parse(raw) as Partial<SerializableTestSession>
    return {
      ...defaultSession,
      ...parsed,
      answersStage1: parsed?.answersStage1 ?? {},
      answersStage2: parsed?.answersStage2 ?? {},
    }
  } catch (error) {
    console.warn('[testSession] Failed to parse session storage', error)
    return { ...defaultSession, answersStage1: {}, answersStage2: {} }
  }
}

const state = reactive<TestSession>(loadFromStorage())

function persist(session: TestSession) {
  if (typeof window === 'undefined') return
  const payload: SerializableTestSession = {
    ...session,
    answersStage1: { ...session.answersStage1 },
    answersStage2: { ...session.answersStage2 },
  }
  window.sessionStorage.setItem(STORAGE_KEY, JSON.stringify(payload))
}

watch(
  () => ({ ...state, answersStage1: { ...state.answersStage1 }, answersStage2: { ...state.answersStage2 } }),
  (value) => {
    persist(value as TestSession)
  },
  { deep: true }
)

function generateSessionId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `session-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export function useTestSession() {
  function ensureSessionId() {
    if (!state.sessionId) {
      state.sessionId = generateSessionId()
    }
    return state.sessionId
  }

  function setVariant(variant: Variant) {
    if (state.variant !== variant) {
      state.variant = variant
      resetForVariant()
    }
  }

  function resetForVariant() {
    state.currentStep = 1
    state.answersStage1 = {}
    state.answersStage2 = {}
    state.age = undefined
    state.mode = undefined
    state.hobby = undefined
  }

  function setCurrentStep(step: number) {
    state.currentStep = step
  }

  function setBasicInfo(payload: { age: number; mode: ModeOption; hobby: string }) {
    ensureSessionId()
    state.age = payload.age
    state.mode = payload.mode
    state.hobby = payload.hobby
  }

  function setAnswer(stage: 1 | 2, questionId: string, value: AnswerValue) {
    const target = stage === 1 ? state.answersStage1 : state.answersStage2
    target[questionId] = value
  }

  function getAnswer(stage: 1 | 2, questionId: string) {
    return (stage === 1 ? state.answersStage1 : state.answersStage2)[questionId]
  }

  function clearAnswers(stage: 1 | 2) {
    if (stage === 1) {
      state.answersStage1 = {}
    } else {
      state.answersStage2 = {}
    }
  }

  function isPageComplete(stage: 1 | 2, questionIds: string[]): boolean {
    const answers = stage === 1 ? state.answersStage1 : state.answersStage2
    return questionIds.every((id) => Boolean(answers[id]))
  }

  function nextStep(maxStep?: number) {
    const limit = maxStep ?? Number.MAX_SAFE_INTEGER
    state.currentStep = Math.min(limit, state.currentStep + 1)
    return state.currentStep
  }

  function prevStep() {
    state.currentStep = Math.max(1, state.currentStep - 1)
    return state.currentStep
  }

  function toPayload() {
    return {
      sessionId: state.sessionId,
      variant: state.variant,
      age: state.age,
      mode: state.mode,
      hobby: state.hobby,
      answersStage1: { ...state.answersStage1 },
      answersStage2: { ...state.answersStage2 },
    }
  }

  return {
    state,
    ensureSessionId,
    setVariant,
    setCurrentStep,
    setBasicInfo,
    setAnswer,
    getAnswer,
    clearAnswers,
    isPageComplete,
    nextStep,
    prevStep,
    toPayload,
  }
}

export type { ModeOption, AnswerValue }
