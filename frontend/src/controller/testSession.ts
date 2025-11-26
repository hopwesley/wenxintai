import { reactive, watch } from 'vue'
import {
    AnswerValue,
    ModeOption, PlanKey,
    type TestFlowStep,
} from '@/controller/common'

const STORAGE_KEY = 'wenxintai:test-session'

export interface TestSession {
    recordPublicID?: string
    businessType?: PlanKey
    testFlowSteps?: TestFlowStep[]
    testRoutes?: string[]
    nextRouteItem: Record<string, number>
    stageAnswers: Record<string, Record<number, AnswerValue>>

    mode?: ModeOption
    hobby?: string
    grade?: string

    inviteCode?: string
    wechatOpenId?: string

    currentStep?: number
}

// 默认值：这里可以按需要补一些默认 businessType 等
const defaultSession: TestSession = {
    recordPublicID: undefined,
    businessType: undefined,
    testFlowSteps: undefined,
    testRoutes: undefined,
    nextRouteItem: {},
    stageAnswers: {},
    mode: undefined,
    hobby: undefined,
    grade: undefined,
    inviteCode: undefined,
    wechatOpenId: undefined,
    currentStep: undefined,
}

/**
 * 从 localStorage 读取 TestSession
 */
function loadFromStorage(): TestSession {
    if (typeof window === 'undefined') {
        return { ...defaultSession }
    }

    const raw = window.localStorage.getItem(STORAGE_KEY)
    if (!raw) {
        return { ...defaultSession }
    }

    try {
        const parsed = JSON.parse(raw) as Partial<TestSession>
        return {
            ...defaultSession,
            ...parsed,
        }
    } catch (error) {
        console.warn('[testSession] Failed to parse localStorage', error)
        return { ...defaultSession }
    }
}

/**
 * 持久化到 localStorage
 */
function persist(session: TestSession) {
    if (typeof window === 'undefined') return
    try {
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
    } catch (error) {
        console.warn('[testSession] Failed to save localStorage', error)
    }
}

/**
 * 清空会话（例如测试完成 / 主动退出时可以调用）
 */
function clearStorage() {
    if (typeof window === 'undefined') return
    try {
        window.localStorage.removeItem(STORAGE_KEY)
    } catch (error) {
        console.warn('[testSession] Failed to clear localStorage', error)
    }
}

// 全局唯一的 reactive state
const state = reactive<TestSession>(loadFromStorage())

// 任何字段变化时自动写回 localStorage
watch(
    () => ({ ...state }),
    (value) => {
        persist(value as TestSession)
    },
    { deep: true }
)

export function useTestSession() {
    function setTestConfig(payload: { grade: string; mode: ModeOption; hobby?: string }) {
        state.grade = payload.grade
        state.mode = payload.mode
        state.hobby = payload.hobby
    }

    function setInviteCode(code: string | null | undefined) {
        const normalized = typeof code === 'string' ? code.trim() : ''
        state.inviteCode = normalized || undefined
    }

    function setBusinessType(typ: PlanKey) {
        if (!typ) return
        state.businessType = typ
    }

    function setTestFlow(steps: TestFlowStep[]) {
        const safeSteps = steps ?? []
        state.testFlowSteps = safeSteps
        state.testRoutes = safeSteps.map(step => step.title)

        if (!state.nextRouteItem) {
            state.nextRouteItem = {}
        }
    }
    function setNextRouteItem(route:string, rid:number){
        if (!route) return

        if (!state.nextRouteItem) {
            state.nextRouteItem = {}
        }
        state.nextRouteItem[route] =  rid
    }

    function setPublicID(pid: string) {
        state.recordPublicID = pid
    }

    function getPublicID() {
        return state.recordPublicID
    }

    // 保存某一阶段的答案
    function saveStageAnswers(stageKey: string, answers: Record<number, AnswerValue>) {
        if (!stageKey) return
        if (!state.stageAnswers) {
            state.stageAnswers = {}
        }
        state.stageAnswers[stageKey] = { ...answers }
    }

    // 读取某一阶段的答案
    function loadStageAnswers(stageKey: string): Record<number, AnswerValue> | undefined {
        if (!stageKey || !state.stageAnswers) return undefined
        return state.stageAnswers[stageKey]
    }

    function resetSession() {
        state.recordPublicID = undefined
        state.businessType = undefined
        state.testRoutes = undefined
        state.nextRouteItem = {}
        state.stageAnswers = {}
        Object.assign(state, { ...defaultSession })
        clearStorage()
    }

    return {
        state,
        setTestConfig,
        setInviteCode,
        setBusinessType,
        setTestFlow,
        setPublicID,
        getPublicID,
        setNextRouteItem,
        saveStageAnswers,
        loadStageAnswers,
        resetSession,
    }
}
