import {computed, onMounted, ref} from 'vue'
import { useRoute,useRouter } from 'vue-router'
import { useTestSession } from '@/store/testSession'
import {StageAsc, StageMotivation, StageOcean, StageRiasec, useSubscriptBySSE} from "@/controller/common";
import {useAlert} from "@/controller/useAlert";

export interface Question {
    id: number;
    text: string;
    dimension: string;
}

export interface ScaleOption {
    value: number;
    label: string;
}

export function useQuestionsStagePage() {
    const route = useRoute()
    const router = useRouter()
    const {state, getPublicID} = useTestSession()
    const {showAlert} = useAlert()

    // ------- 步骤条 & 标题 & loading -------
    const loading = ref(true)

    function showLoading() {
        loading.value = true
    }

    function hideLoading() {
        loading.value = false
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

    const latestMessage = ref('')
    const logLines = ref<string[]>([])
    const MAX_LOG_LINES = 8
    const truncatedLatestMessage = computed(() => logLines.value)

    const isSubmitting = ref(false)

    const scaleOptions = ref<ScaleOption[]>([
        {value: 1, label: '从不'},
        {value: 2, label: '很少'},
        {value: 3, label: '一般'},
        {value: 4, label: '经常'},
        {value: 5, label: '非常多'},
    ])

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
    const {businessType, testStage} = route.params as { businessType: string; testStage: string }

    console.log('[QuestionsStagePage] apply_test resp:', testStage, businessType, public_id, routes)

    // ------- SSE 拉题目 -------
    onMounted(() => {
        errorMessage.value = ''

        if (!public_id || !routes.length || !validateTestStage(testStage)) {
            showAlert('测试流程异常，请返回首页重新开始', () => {
                router.replace('/').then()
            })
            return
        }

        const idx = routes.findIndex(r => r.router === String(testStage || ''))
        if (idx === -1) {
            showAlert('测试流程异常，未能识别当前步骤，请返回首页重新开始', () => {
                router.replace('/').then()
            })
            return
        }

        let message = ''

        const sseCtrl = useSubscriptBySSE(public_id, businessType, testStage, {
            autoStart: false,

            onOpen() {
                showLoading()
            },

            onError(err) {
                console.log('------>>> sse channel error:', err)
                showAlert('获取测试流程失败，请稍后再试:' + err)
                hideLoading()
            },

            onMsg(chunk) {
                message += chunk
                if (message.length < 20) {
                    return
                }
                logLines.value.push(message)
                if (logLines.value.length > MAX_LOG_LINES) {
                    logLines.value.splice(0, logLines.value.length - MAX_LOG_LINES)
                }
                message='';
            },

            onClose() {
                console.log('------>>> sse closed:')
                hideLoading()
            },

            onDone(questionStr) {
                const raw = (questionStr && questionStr.trim().length > 0) ? questionStr : message
                console.log('------>>> go questions:', raw)

                try {
                    const parsed = JSON.parse(raw) as Question[]
                    if (!Array.isArray(parsed) || parsed.length === 0) {
                        throw new Error('empty questions')
                    }

                    questions.value = parsed
                    currentPage.value = 1
                    highlightedQuestions.value = {}
                } catch (e) {
                    console.error('[QuestionsStagePage] 解析题目失败:', e)
                    errorMessage.value = '获取测试题目失败，请稍后再试'
                    showAlert('获取测试题目失败，请稍后再试')
                } finally {
                    hideLoading()
                }
            },
        })

        sseCtrl.start()
    })

    // ------- 翻页逻辑 -------
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
            console.log('[QuestionsStagePage] 当前阶段答题结果:', answers.value)
            showAlert('本阶段所有题目已完成（提交逻辑待接入）')
        } finally {
            isSubmitting.value = false
        }
    }

    return {
        // 布局 & 步骤条
        route,
        loading,
        stepItems,
        currentStep,
        currentStepTitle,

        // 分页 & 题目 & 答案
        totalCount,
        totalPages,
        pageStartIndex,
        pageEndIndex,
        pagedQuestions,
        scaleOptions,
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
    }
}
