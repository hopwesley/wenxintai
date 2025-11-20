import {computed, ref, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useTestSession} from '@/store/testSession'
import {
    AnswerValue,
    CommonResponse,
    StageAsc,
    StageMotivation,
    StageOcean,
    StageRiasec,
    useSubscriptBySSE
} from "@/controller/common";
import {useAlert} from "@/controller/useAlert";
import {useGlobalLoading} from "@/controller/useGlobalLoading";
import {apiRequest} from "@/api";

export interface Question {
    id: number
    text: string

    // RIASEC 专用
    dimension?: string

    // ASC 专用
    subject?: string        // "PHY"
    subject_label?: string  // "物理"
    reverse?: boolean
    subtype?: string        // "Comparison" | "Efficacy" | ...

    // 之后 OCEAN / SDT / MI 也可以继续往这里挂可选字段
}

interface StageSnapshot {
    questions: Question[]
    answers: Record<number, AnswerValue>
    currentPage: number
}


export interface RiasecAnswerPayload {
    id: number
    dimension: string  // R / I / A / S / E / C
    value: AnswerValue      // 1~5
}

export interface AscAnswerPayload {
    id: number
    subject: string        // "PHY"
    subject_label: string  // "物理"（可选：看你是否真的要存）
    value: AnswerValue          // 1~5
    reverse: boolean
    subtype: string        // "Comparison" | "Efficacy" ...
}

export interface OceanAnswerPayload {
    id: number
    value: AnswerValue      // 1~5
    dimension: string      // "O" / "C" / "E" / "A" / "N"
    reverse: boolean
}

export type AnyAnswerPayload =
    | RiasecAnswerPayload
    | AscAnswerPayload
    | OceanAnswerPayload

export function useQuestionsStagePage() {

    const route = useRoute()
    const router = useRouter()
    const {state, getPublicID, setNextRouteItem, saveStageAnswers, loadStageAnswers} = useTestSession()
    const {showAlert} = useAlert()
    const aiLoading = ref(true)
    const {showLoading, hideLoading} = useGlobalLoading()

    const pageSize = 5
    const currentPage = ref(1)
    const questions = ref<Question[]>([])
    const answers = ref<Record<number, AnswerValue>>({})
    const highlightedQuestions = ref<Record<number, boolean>>({})
    const logLines = ref<string[]>([])
    const MAX_LOG_LINES = 8
    const truncatedLatestMessage = computed(() => logLines.value)
    const isSubmitting = ref(false)
    const totalCount = computed(() => questions.value.length)
    const totalPages = computed(() =>
        totalCount.value > 0 ? Math.ceil(totalCount.value / pageSize) : 1
    )
    const pageStartIndex = computed(() => (currentPage.value - 1) * pageSize)
    const pageEndIndex = computed(() =>
        Math.min(pageStartIndex.value + pageSize, totalCount.value)
    )
    const pagedQuestions = computed(() =>
        questions.value.slice(pageStartIndex.value, pageEndIndex.value)
    )
    const isFirstPage = computed(() => currentPage.value <= 1)
    const isLastPage = computed(() => currentPage.value >= totalPages.value)
    const rawMessage = ref('')

    const public_id: string | undefined = getPublicID()
    const routes = state.testRoutes ?? []
    const businessType = computed(() =>
        String(route.params.businessType ?? state.businessType ?? '')
    )

    const testStage = computed(() =>
        String(route.params.testStage ?? '')
    )

    watch(
        () => [businessType.value, testStage.value],
        () => {
            initStageForCurrentRoute()
        },
        {immediate: true},
    )

    // 唯一标识当前阶段，用于在 store 中找到对应答案
    const stageKey = computed(() => {
        const pid = public_id
        const stage = testStage.value
        const biz = businessType.value
        if (!pid || !stage || !biz) return ''
        return `qstage:${biz}:${stage}:${pid}`
    })

    // 只要题目已经加载出来，且答案有变动，就把当前阶段的答案写回 store
    watch(
        () => ({
            key: stageKey.value,
            answers: answers.value,
            hasQuestions: questions.value.length > 0,
        }),
        ({ key, answers, hasQuestions }) => {
            if (!key || !hasQuestions) return
            // 注意：这里不会缓存题目，只把答案 map 写到 store.stageAnswers[key]
            saveStageAnswers(key, answers)
        },
        { deep: true }
    )


    function showAIProcess() {
        aiLoading.value = true
    }

    function hideAIProcess() {
        aiLoading.value = false
    }

    function validateTestStage(testStage: string): boolean {
        const validStages = [
            StageRiasec,
            StageAsc,
            StageOcean,
            StageMotivation,
        ]

        if (!validStages.includes(testStage)) {
            showAlert('测试流程异常，请返回首页重新开始', () => {
                router.replace('/').then()
            })
            return false
        }
        return true
    }

    function isQuestionHighlighted(id: number): boolean {
        return highlightedQuestions.value[id]
    }

    function handleSseError(err: Error) {
        console.log('------>>> sse channel error:', err)
        showAlert('获取测试流程失败，请稍后再试:' + err)
        hideAIProcess()
    }

    function handleSseMsg(chunk: string) {
        rawMessage.value += chunk
        if (rawMessage.value.length < 20) {
            return
        }
        logLines.value.push(rawMessage.value)
        if (logLines.value.length > MAX_LOG_LINES) {
            logLines.value.splice(0, logLines.value.length - MAX_LOG_LINES)
        }
        rawMessage.value = ''
    }

    function resetStageState() {
        currentPage.value = 1
        questions.value = []
        answers.value = {}
        highlightedQuestions.value = {}
        logLines.value = []
        rawMessage.value = ''
        isSubmitting.value = false
    }

    function handleSseDone(questionStr: string) {
        console.log('------>>> go questions:', questionStr)
        try {
            const parsed = JSON.parse(questionStr) as Question[]
            if (!Array.isArray(parsed) || parsed.length === 0) {
                showAlert('获取 AI 题目失败,请刷新页面进行重试')
                return
            }

            questions.value = parsed
            currentPage.value = 1
            highlightedQuestions.value = {}
            //TODO::applyAnswersForCurrentStage

        } catch (e) {
            console.error('[QuestionsStagePage] 解析题目失败:', e)
            showAlert('获取测试题目失败，请稍后再试')
        } finally {
            hideAIProcess()
        }
    }

    /**
     * 统一处理“本阶段应该展示哪些答案”
     * @param rawAnswers 从后端返回的 raw.answers（可能没有、可能是一个对象）
     *
     * 优先级：
     * 1. 如果 rawAnswers 有值，认为是服务器从数据库加载出来的标准答案 → 以它为准
     * 2. 否则，如果本地有缓存（stageAnswers[stageKey]），用本地缓存
     * 3. 如果都没有，就保持当前 answers 不变（通常是空）
     */
    function applyAnswersForCurrentStage(rawAnswers: unknown) {
        const key = stageKey.value
        let finalAnswers: Record<number, AnswerValue> | undefined

        // 1. 优先使用服务器返回的答案（raw.answers）
        if (rawAnswers && typeof rawAnswers === 'object') {
            finalAnswers = rawAnswers as Record<number, AnswerValue>
        } else if (key) {
            // 2. 没有服务器答案时，尝试用本地缓存
            const cached = loadStageAnswers(key)
            if (cached) {
                finalAnswers = cached
            }
        }

        // 3. 有最终确定的答案，就写回本地 state + 缓存
        if (finalAnswers) {
            // 更新当前页面上的答案
            answers.value = { ...finalAnswers }

            // 同步写回全局缓存，保证下次进这个阶段还能恢复
            if (key) {
                saveStageAnswers(key, finalAnswers)
            }
        }
    }


    function initStageForCurrentRoute() {
        resetStageState()

        const key = stageKey.value
        if (key) {
            const cached = loadStageAnswers(key)
            if (cached) {
                answers.value = { ...cached }
            }
        }

        const stage = testStage.value
        const bizType = businessType.value

        console.log('[QuestionsStagePage] init stage:', stage, 'aiLoading=', aiLoading.value, 'key=', key)

        // 先尝试恢复本阶段的本地答案（题目仍然交给 SSE 来加载）


        // 2. 没有缓存 -> 走原有 AI 拉题逻辑
        showAIProcess()

        if (!public_id || !routes.length || !validateTestStage(stage)) {
            showAlert('测试流程异常，请返回首页重新开始', () => {
                router.replace('/').then()
            })
            return
        }


        const sseCtrl = useSubscriptBySSE(public_id, bizType, stage, {
            autoStart: false,
            onOpen: showAIProcess,
            onError: handleSseError,
            onMsg: handleSseMsg,
            onClose: hideAIProcess,
            onDone: handleSseDone,
        })

        sseCtrl.start()
    }

    async function submitCurrentStageAnswers() {
        const payload = {
            public_id,
            business_type: businessType.value,
            test_type: testStage.value,
            answers: buildAnswersPayloadForCurrentStage(),
        }
        const resp = await apiRequest<CommonResponse>('/api/test_submit', {method: 'POST', body: payload})

        if (!resp.ok) {
            throw new Error(resp.msg || '提交失败，请稍后重试')
        }

        if(!resp.next_route){
            throw new Error(resp.msg || '未找到下一步处理逻辑')
        }

        let next_route_index = resp.next_route_index ?? 0
        if (next_route_index < 0) {
            next_route_index = 0
        }

        setNextRouteItem(resp.next_route, next_route_index)
        return resp.next_route
    }

    function gotoNextStageAfterSubmit(next_route: string | null) {
        if (!next_route) {
            showAlert('测试流程异常，未找到下一步，请返回首页重新开始', () => {
                router.replace('/').then()
            })
            return
        }

        const currentBusinessType = state.businessType || businessType.value
        if (!currentBusinessType) {
            showAlert('测试流程异常，未找到测评类型，请返回首页重新开始', () => {
                router.replace('/').then()
            })
            return
        }
        router.push(`/assessment/${currentBusinessType}/${next_route}`).then()
    }


    function handlePrev() {
        if (isFirstPage.value || isSubmitting.value) return
        currentPage.value -= 1
        highlightedQuestions.value = {}
    }

    async function handleNext() {
        if (!questions.value.length) {
            return
        }

        const pageQs = pagedQuestions.value
        const missingIds: number[] = []

        for (const q of pageQs) {
            const v = answers.value[q.id]
            if (v == null) {
                missingIds.push(q.id)
            }
        }

        if (missingIds.length > 0) {
            const map: Record<number, boolean> = {}
            missingIds.forEach(id => {
                map[id] = true
            })
            highlightedQuestions.value = map
            showAlert('请先完成本页所有题目')
            return
        }

        highlightedQuestions.value = {}

        if (currentPage.value < totalPages.value) {
            currentPage.value += 1
            return
        }

        isSubmitting.value = true
        try {
            showLoading('正在提交答案，请稍候…', 15000)
            const next_route = await submitCurrentStageAnswers()
            gotoNextStageAfterSubmit(next_route)
        } catch (err) {
            console.error('[QuestionsStagePage] 提交失败:', err)
            const msg = err instanceof Error ? err.message : '提交失败，请稍后重试'
            showAlert(msg)
        } finally {
            isSubmitting.value = false
            hideLoading()
        }
    }

    function buildRiasecAnswers(
        questions: Question[],
        answersMap: Record<number, number>,
    ): RiasecAnswerPayload[] {
        return questions
            .filter(q => answersMap[q.id] != null && q.dimension)
            .map(q => ({
                id: q.id,
                dimension: q.dimension as string,
                value: answersMap[q.id] as AnswerValue,
            }))
    }

    function buildAscAnswers(
        questions: Question[],
        answersMap: Record<number, number>,
    ): AscAnswerPayload[] {
        return questions
            .filter(q => answersMap[q.id] != null && q.subject)
            .map(q => ({
                id: q.id,
                subject: q.subject as string,             // "PHY"
                subject_label: q.subject_label || '',     // 视需求决定要不要存
                value: answersMap[q.id] as AnswerValue,        // 1~5
                reverse: !!q.reverse,
                subtype: q.subtype || '',
            }))
    }

    function buildOceanAnswers(
        questions: Question[],
        answersMap: Record<number, number>,
    ): OceanAnswerPayload[] {
        return questions
            .filter(q => answersMap[q.id] != null && q.dimension)
            .map(q => ({
                id: q.id,
                dimension: q.dimension as string,
                value: answersMap[q.id] as AnswerValue,        // 1~5
                reverse: !!q.reverse,
            }))
    }

    function buildAnswersPayloadForCurrentStage(): AnyAnswerPayload[] {
        const stage = testStage.value   // RIASEC / ASC / ...

        const map = answers.value
        const qs = questions.value
        switch (stage) {
            case StageRiasec:
                return buildRiasecAnswers(qs, map)
            case StageAsc:
                return buildAscAnswers(qs, map)
            case StageOcean:
                return buildOceanAnswers(qs, map)
            default:
                return qs
                    .filter(q => map[q.id] != null)
                    .map(q => ({
                        id: q.id,
                        value: map[q.id] as number,
                    })) as AnyAnswerPayload[]
        }
    }

    const currentStepTitle = computed(() => {
        const routes = state.testRoutes ?? []
        const stageKey = String(route.params.testStage ?? '')
        const idx = state.nextRouteItem?.[stageKey] ?? 0
        if (idx >= 0 && idx < routes.length) {
            return routes[idx] || '正在加载…'
        }
        return '正在加载…'
    })

    return {
        // 布局 & 步骤条
        route,
        aiLoading,
        // 分页 & 题目 & 答案
        totalCount,
        totalPages,
        pageStartIndex,
        pageEndIndex,
        pagedQuestions,
        answers,
        isFirstPage,
        isLastPage,
        isSubmitting,

        // 高亮 & 日志 & 行为
        truncatedLatestMessage,
        isQuestionHighlighted,
        handlePrev,
        handleNext,
        currentStepTitle,
    }
}
