const STORAGE_KEY = 'assessment-flow-state'

export type StageKey = 'S1' | 'S2'

export interface StoredQuestionSet {
  stage: StageKey
  questions: any
}

export interface AssessmentFlowState {
  assessmentId?: string
  activeQuestionSetId?: string
  questionSets: Record<string, StoredQuestionSet>
  latestReportId?: string
}

function readStorage(): AssessmentFlowState {
  if (typeof window === 'undefined') {
    return { questionSets: {} }
  }
  const raw = window.localStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return { questionSets: {} }
  }
  try {
    const parsed = JSON.parse(raw) as AssessmentFlowState
    if (!parsed || typeof parsed !== 'object') {
      return { questionSets: {} }
    }
    return {
      assessmentId: parsed.assessmentId,
      activeQuestionSetId: parsed.activeQuestionSetId,
      latestReportId: parsed.latestReportId,
      questionSets: parsed.questionSets ?? {}
    }
  } catch (error) {
    console.warn('[assessmentFlow] failed to parse storage', error)
    return { questionSets: {} }
  }
}

function writeStorage(state: AssessmentFlowState) {
  if (typeof window === 'undefined') {
    return
  }
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
}

function updateState(mutator: (draft: AssessmentFlowState) => void): AssessmentFlowState {
  const current = readStorage()
  const draft: AssessmentFlowState = {
    assessmentId: current.assessmentId,
    activeQuestionSetId: current.activeQuestionSetId,
    latestReportId: current.latestReportId,
    questionSets: { ...current.questionSets }
  }
  mutator(draft)
  writeStorage(draft)
  return draft
}

export function getAssessmentFlowState(): AssessmentFlowState {
  return readStorage()
}

export function setAssessmentId(assessmentId: string): AssessmentFlowState {
  return updateState((state) => {
    state.assessmentId = assessmentId
  })
}

export function setQuestionSet(questionSetId: string, stage: StageKey, questions: any): AssessmentFlowState {
  return updateState((state) => {
    state.questionSets[questionSetId] = { stage, questions }
    state.activeQuestionSetId = questionSetId
  })
}

export function setActiveQuestionSet(questionSetId: string | undefined): AssessmentFlowState {
  return updateState((state) => {
    state.activeQuestionSetId = questionSetId
  })
}

export function clearQuestionSets(): AssessmentFlowState {
  return updateState((state) => {
    state.questionSets = {}
    state.activeQuestionSetId = undefined
  })
}

export function recordReport(reportId: string): AssessmentFlowState {
  return updateState((state) => {
    state.latestReportId = reportId
    state.activeQuestionSetId = undefined
  })
}

export function getQuestionSet(questionSetId: string): StoredQuestionSet | undefined {
  const state = readStorage()
  return state.questionSets[questionSetId]
}

export function clearAll(): void {
  if (typeof window === 'undefined') {
    return
  }
  window.localStorage.removeItem(STORAGE_KEY)
}
