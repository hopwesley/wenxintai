import { reactive, watch } from 'vue'
import {
    AnswerValue,
    ModeOption,
    PlanKey,
    TestFlowStep,
} from '@/controller/common'

const STORAGE_KEY = 'wenxintai:test-session'

/** 后端返回的完整 TestRecordDTO */
export interface TestRecordDTO {
    public_id: string
    business_type: PlanKey
    pay_order_id?: string
    wechat_id?: string
    grade?: string
    mode: ModeOption|''
    hobby?: string
    status: number
    created_at: string
}

/** Session 结构 */
export interface TestSession {
    record?: TestRecordDTO                              // ★ 用 record 取代所有散落字段

    testFlowSteps?: TestFlowStep[]
    testRoutes?: string[]
    nextRouteItem: Record<string, number>

    stageAnswers: Record<string, Record<number, AnswerValue>>

    currentStep?: number
}

/** 默认值 */
const defaultSession: TestSession = {
    record: undefined,
    testFlowSteps: undefined,
    testRoutes: undefined,
    nextRouteItem: {},
    stageAnswers: {},
    currentStep: undefined,
}

/** 从 localStorage 读取 */
function loadFromStorage(): TestSession {
    if (typeof window === 'undefined') {
        return { ...defaultSession }
    }

    const raw = window.localStorage.getItem(STORAGE_KEY)
    if (!raw) return { ...defaultSession }

    try {
        const parsed = JSON.parse(raw) as Partial<TestSession>
        return {
            ...defaultSession,
            ...parsed,
        }
    } catch (e) {
        console.warn('[testSession] Failed to parse storage', e)
        return { ...defaultSession }
    }
}

/** 保存到 localStorage */
function persist(session: TestSession) {
    if (typeof window === 'undefined') return
    try {
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
    } catch (e) {
        console.warn('[testSession] Failed to save storage', e)
    }
}

/** 清空 localStorage */
function clearStorage() {
    if (typeof window === 'undefined') return
    try {
        window.localStorage.removeItem(STORAGE_KEY)
    } catch (e) {
        console.warn('[testSession] Failed to clear storage', e)
    }
}

/** 全局唯一 reactive state */
const state = reactive<TestSession>(loadFromStorage())

/** 任意字段变化 → 自动写回 storage */
watch(
    () => ({ ...state }),
    val => persist(val as TestSession),
    { deep: true }
)

/** 对外 API */
export function useTestSession() {

    /** ★ 设置 record（来自 /api/test_flow 的 resp.record） */
    function setRecord(rec: TestRecordDTO) {
        if (!rec) return
        state.record = { ...rec }
    }

    /** 流程步骤 */
    function setTestFlow(steps: TestFlowStep[]) {
        const safe = steps ?? []
        state.testFlowSteps = safe
        state.testRoutes = safe.map(s => s.title)

        if (!state.nextRouteItem) {
            state.nextRouteItem = {}
        }
    }

    /** 设置下一跳路由索引 */
    function setNextRouteItem(route: string, idx: number) {
        if (!route) return
        if (!state.nextRouteItem) state.nextRouteItem = {}
        state.nextRouteItem[route] = idx
    }

    /** 保存某阶段答案 */
    function saveStageAnswers(stage: string, answers: Record<number, AnswerValue>) {
        if (!stage) return
        if (!state.stageAnswers) state.stageAnswers = {}
        state.stageAnswers[stage] = { ...answers }
    }

    /** 读取某阶段答案 */
    function loadStageAnswers(stage: string) {
        return state.stageAnswers?.[stage]
    }

    /** 重置整个 session */
    function resetSession() {
        Object.assign(state, { ...defaultSession })
        clearStorage()
    }

    return {
        state,
        setRecord,
        setTestFlow,
        setNextRouteItem,
        saveStageAnswers,
        loadStageAnswers,
        resetSession,
    }
}
