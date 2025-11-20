import { reactive, watch } from 'vue'
import { ModeOption, TestTypeBasic, TestTypePro, TestTypeSchool } from '@/controller/common'

const STORAGE_KEY = 'wenxintai:test-session'

export interface TestSession {
    // 当前测试记录在后端 tests_record 表中的 public_id
    recordPublicID?: string

    // basic / pro / school 之一，或者未来扩展的字符串
    businessType?: typeof TestTypeBasic | typeof TestTypePro | typeof TestTypeSchool | string

    // 测试流程的路由列表
    testRoutes?: string[]
    nextRouteItem: Record<string, number>

    // BasicInfo / AssessmentBasicInfo 收集到的配置
    mode?: ModeOption
    hobby?: string
    grade?: string

    // 入口信息
    inviteCode?: string
    wechatOpenId?: string

    // 当前步骤（如果不需要持久化，也可以以后删掉）
    currentStep?: number
}

// 默认值：这里可以按需要补一些默认 businessType 等
const defaultSession: TestSession = {
    recordPublicID: undefined,
    businessType: undefined,
    testRoutes: undefined,
    nextRouteItem: {},
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
        state.inviteCode = normalized || undefined
    }

    function setBusinessType(type: typeof TestTypeBasic | typeof TestTypePro | typeof TestTypeSchool | string) {
        state.businessType = type
    }

    function setTestRoutes(routes: string[]) {
        state.testRoutes = routes
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

    function resetSession() {
        // 重置内存里的 state
        Object.assign(state, { ...defaultSession })
        // 清理 localStorage
        clearStorage()
    }

    return {
        state,

        setCurrentStep,
        setTestConfig,
        setInviteCode,
        setBusinessType,
        setTestRoutes,
        setPublicID,
        getPublicID,
        setNextRouteItem,
    }
}
