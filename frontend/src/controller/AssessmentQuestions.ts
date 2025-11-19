import {computed, onMounted, ref, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useTestSession} from '@/store/testSession'
import {
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
    id: number;
    text: string;
    dimension: string;
}

export interface ScaleOption {
    value: number;
    label: string;
}

export interface AnswerTriple {
    id: number;         // 题目编号
    dimension: string;  // R / I / A / S / E / C
    value: number;      // 用户选择的 1~5
}


export function useQuestionsStagePage() {
    const route = useRoute()
    const router = useRouter()
    const {state, getPublicID} = useTestSession()
    const {showAlert} = useAlert()

    // ------- 步骤条 & 标题 & loading -------
    const aiLoading = ref(true)

    const {showLoading, hideLoading} = useGlobalLoading()

    function showAIProcess() {
        aiLoading.value = true
    }

    function hideAIProcess() {
        aiLoading.value = false
    }

    const stepItems = computed(() => {
        const routes = state.testRoutes ?? []
        return routes.map(r => ({
            key: r.router,
            title: r.desc,
        }))
    })

    const currentStep = computed(() => {
        const routes = state.testRoutes ?? []
        const testStage = String(route.params.testStage ?? '')
        const idx = routes.findIndex(r => r.router === testStage)
        return idx >= 0 ? idx + 1 : 0
    })

    const currentStepTitle = computed(() => {
        const routes = state.testRoutes ?? []
        const idx = currentStep.value - 1
        if (idx >= 0 && idx < routes.length) {
            return routes[idx].desc || '正在加载…'
        }
        return '正在加载…'
    })

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

    // ------- 本页面自己的状态（分页 / 题目 / 日志 / 提交） -------
    const pageSize = 5
    const currentPage = ref(1)
    const questions = ref<Question[]>([])
    const answers = ref<Record<number, number>>({})
    const highlightedQuestions = ref<Record<number, boolean>>({})
    const errorMessage = ref('')

    const logLines = ref<string[]>([])
    const MAX_LOG_LINES = 8
    const truncatedLatestMessage = computed(() => logLines.value)

    const isSubmitting = ref(false)

    const answerTriples = computed<AnswerTriple[]>(() => {
        const map = answers.value
        return questions.value
            .filter(q => map[q.id] != null)
            .map(q => ({
                id: q.id,
                dimension: q.dimension,
                value: map[q.id] as number,
            }))
    })

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

    function isQuestionHighlighted(id: number): boolean {
        return highlightedQuestions.value[id]
    }

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
        { immediate: true },
    )

    console.log('[QuestionsStagePage] apply_test resp:', testStage.value, businessType.value, public_id, routes)

    const rawMessage = ref('')


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
        errorMessage.value = ''
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
        } catch (e) {
            console.error('[QuestionsStagePage] 解析题目失败:', e)
            errorMessage.value = '获取测试题目失败，请稍后再试'
            showAlert('获取测试题目失败，请稍后再试')
        } finally {
            hideAIProcess()
        }
    }

    function initStageForCurrentRoute() {
        resetStageState()
        errorMessage.value = ''

        const stage = testStage.value
        const bizType = businessType.value

        if (!public_id || !routes.length || !validateTestStage(stage)) {
            showAlert('测试流程异常，请返回首页重新开始', () => {
                router.replace('/').then()
            })
            return
        }

        const idx = routes.findIndex(r => r.router === String(stage || ''))
        if (idx === -1) {
            showAlert('测试流程异常，未能识别当前步骤，请返回首页重新开始', () => {
                router.replace('/').then()
            })
            return
        }

        const sseCtrl = useSubscriptBySSE(public_id, bizType, stage, {
            autoStart: false,
            onOpen: showLoading,
            onError: handleSseError,
            onMsg: handleSseMsg,
            onClose: hideLoading,
            onDone: handleSseDone,
        })

        sseCtrl.start()
    }

    async function submitCurrentStageAnswers() {
        const payload = {
            public_id,
            business_type: businessType.value,
            test_type: testStage.value,
            answers: answerTriples.value,
        }
        return apiRequest<CommonResponse>('/api/test_submit', {
            method: 'POST',
            body: payload,
        })
    }

    function gotoNextStageAfterSubmit() {
        const routes = state.testRoutes ?? []
        if (!routes.length) {
            showAlert('测试流程异常，未找到下一步，请返回首页重新开始', () => {
                router.replace('/').then()
            })
            return
        }

        // 当前阶段（比如 riasec / asc / ocean / motivation）
        const currentStage = String(testStage.value || '')
        const idx = routes.findIndex((r) => r.router === currentStage)

        // 找不到自己，或者已经是最后一个步骤了：都算流程错误
        if (idx < 0 || idx === routes.length - 1) {
            showAlert('测试流程异常，未找到下一步，请返回首页重新开始', () => {
                router.replace('/').then()
            })
            return
        }

        const next = routes[idx + 1]

        // businessType 从全局状态优先，退回到当前路由参数
        const currentBusinessType = state.businessType || businessType
        if (!currentBusinessType) {
            showAlert('测试流程异常，未找到测评类型，请返回首页重新开始', () => {
                router.replace('/').then()
            })
            return
        }

        // 和 AssessmentBasicInfo 一样，直接用路径拼接：
        //  /assessment/${businessType}/${next.router}
        // 如果 next.router 是 report，会自动匹配到报告页路由
        router.push(`/assessment/${currentBusinessType}/${next.router}`).then()
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
            await submitCurrentStageAnswers()
            gotoNextStageAfterSubmit()
        } catch (err) {
            console.error('[QuestionsStagePage] 提交失败:', err)
            const msg = err instanceof Error ? err.message : '提交失败，请稍后重试'
            showAlert(msg)
        } finally {
            isSubmitting.value = false
            hideLoading()
        }
    }

    return {
        // 布局 & 步骤条
        route,
        loading: aiLoading,
        stepItems,
        currentStep,
        currentStepTitle,

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
        errorMessage,

        // 高亮 & 日志 & 行为
        truncatedLatestMessage,
        isQuestionHighlighted,
        handlePrev,
        handleNext,

        // 新增：带 dimension 的答案结构
        answerTriples,
    }
}
