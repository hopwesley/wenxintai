import { reactive, watch } from 'vue'
import type { Variant } from '@/config/testSteps'

const STORAGE_KEY = 'wenxintai:test-session'
const SESSION_ID_KEY = 'session_id'

type AnswerValue = 1 | 2 | 3 | 4 | 5

type ModeOption = '3+3' | '3+1+2'

export interface TestSession {
  sessionId?: string
  variant: Variant
  currentStep: number
  mode?: ModeOption
  hobby?: string
  grade?: string
  inviteCode?: string
  answersStage1: Record<string, AnswerValue>
  answersStage2: Record<string, AnswerValue>
}

function readPersistedSessionId(): string | undefined {
  if (typeof window === 'undefined') {
    return undefined
  }
  const stored = window.localStorage.getItem(SESSION_ID_KEY)
  return stored ?? undefined
}

const defaultSession: TestSession = {
  sessionId: readPersistedSessionId(),
  variant: 'basic',
  currentStep: 1,
  mode: undefined,
  hobby: undefined,
  grade: undefined,
  inviteCode: undefined,
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
      sessionId: readPersistedSessionId() ?? parsed?.sessionId,
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

watch(
  () => state.sessionId,
  (value) => {
    if (typeof window === 'undefined') return
    if (typeof value === 'string' && value.trim()) {
      window.localStorage.setItem(SESSION_ID_KEY, value)
    } else {
      window.localStorage.removeItem(SESSION_ID_KEY)
    }
  }
)

export function useTestSession() {
  function ensureSessionId() {
    if (state.sessionId) {
      return state.sessionId
    }
    const persisted = readPersistedSessionId()
    if (persisted) {
      state.sessionId = persisted
      return persisted
    }
    throw new Error('NO_SESSION')
  }

  function getSessionId() {
    return state.sessionId ?? null
  }

  function setSessionId(id: string | null | undefined) {
    state.sessionId = id ?? undefined
    if (typeof window === 'undefined') {
      return
    }
    if (id && id.trim()) {
      window.localStorage.setItem(SESSION_ID_KEY, id)
    } else {
      window.localStorage.removeItem(SESSION_ID_KEY)
    }
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
    state.mode = undefined
    state.hobby = undefined
    state.grade = undefined
  }

  function setCurrentStep(step: number) {
    state.currentStep = step
  }


  function setTestConfig(payload: { grade: string; mode: ModeOption; hobby?: string }) {
    state.grade = payload.grade
    state.mode = payload.mode
    state.hobby = payload.hobby
  }

  function setInviteCode(code: string | null | undefined) {
    const normalized = typeof code === 'string' ? code.trim() : ''
    state.inviteCode = normalized ? normalized : undefined
  }

  function getInviteCode() {
    return state.inviteCode ?? null
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
      mode: state.mode,
      hobby: state.hobby,
      grade: state.grade,
      inviteCode: state.inviteCode,
      answersStage1: { ...state.answersStage1 },
      answersStage2: { ...state.answersStage2 },
    }
  }

  return {
    state,
    ensureSessionId,
    getSessionId,
    setSessionId,
    setVariant,
    setCurrentStep,
    setTestConfig,
    setInviteCode,
    getInviteCode,
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
