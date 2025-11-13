import { computed, ref, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTestSession } from '@/store/testSession'
import { STEPS, type Variant, isVariant } from '@/config/testSteps'
import {getQuestions, redeemInvite, submitTestSession} from '@/api'

// 与后端/Store 的真实定义保持一致：仅本文件内使用，避免编译冲突
type ModeOption = '3+3' | '3+1+2'
type AnswerValue = 1 | 2 | 3 | 4 | 5

export interface UseQuestionsStageOptions { stage: 1 | 2; pageSize?: number }

interface Question { id: string; text: string }
interface StageQuestions { stage1: Question[]; stage2: Question[] }
type StepLite = { key: string; titleKey?: string }

export function useQuestionsStage(opts: UseQuestionsStageOptions) {
    const pageSize = opts.pageSize ?? 5
    const route = useRoute()
    const router = useRouter()

    const {
        state, getSessionId, setVariant, setCurrentStep,
        setAnswer, isPageComplete, nextStep, prevStep,
    } = useTestSession()

    const variant = ref<Variant>('basic')
    const currentStep = ref(2) // 第二步（第一阶段）
    const loading = ref(true)
    const submitting = ref(false)
    const errorMessage = ref('')

    const questions = ref<Question[]>([])
    const currentPage = ref(1)
    const highlightedId = ref<string | null>(null)
    const refs = new Map<string, HTMLElement>()

    // 路由参数 → 类型守卫，避免 oldVal 解构导致 undefined 崩溃
    watchEffect(() => {
        const v = String(route.params.variant ?? 'basic')
        if (isVariant(v)) {
            variant.value = v
            setVariant(v)
        } else {
            router.replace({ path: '/test/basic/step/1' })
            return
        }

        const stepNum = Number(route.params.step ?? '2')
        const stepsLen = ((STEPS as Record<Variant, readonly StepLite[]>)[variant.value] ?? []).length
        if (Number.isNaN(stepNum) || stepNum < 1 || stepNum > stepsLen) {
            router.replace({ path: `/test/${variant.value}/step/1` })
            return
        }
        currentStep.value = stepNum
        setCurrentStep(stepNum)
    })

    const stepItems = computed(() => {
        const arr = (STEPS as Record<Variant, readonly StepLite[]>)[variant.value] ?? []
        return arr.map(it => ({ key: it.key, title: it.titleKey ?? it.key }))
    })

    const cached = ref<StageQuestions | null>(null)

    watchEffect(() => { void loadQuestions() })

    async function loadQuestions() {
        // 基础资料缺失 → 回到 Step1
        if (!state.mode || !state.grade) {
            await router.replace({ path: `/test/${variant.value}/step/1` })
            return
        }
        // 第二阶段必须完成第一阶段
        if (opts.stage === 2 && (!state.answersStage1 || Object.keys(state.answersStage1).length === 0)) {
            await router.replace({ path: `/test/${variant.value}/step/2` })
            return
        }

        if (loading.value) return;

        loading.value = true
        errorMessage.value = ''
        try {
            if (!cached.value) {
                const sessionId = getSessionId()
                if (!sessionId) {
                    if (typeof window !== 'undefined') {
                        window.alert('需要邀请码或登录后访问')
                    }
                    await router.replace({ path: '/' })
                    return
                }
                const resp = await getQuestions({
                    session_id: sessionId,
                    mode: (state.mode as ModeOption) ?? '3+3',
                    grade: String(state.grade ?? ''),
                    hobby: state.hobby ?? '',
                })
                cached.value = normalize(resp)
            }
            const list = (opts.stage === 1 ? cached.value?.stage1 : cached.value?.stage2) ?? []
            questions.value = list
            if (!list.length) errorMessage.value = '题目拉取为空'
            if (currentPage.value > totalPages.value) currentPage.value = totalPages.value
        } catch (e) {
            console.error('[useQuestionsStage] loadQuestions error', e)
            if (e instanceof Error) {
                errorMessage.value = e.message
                if (e.name === 'NO_SESSION' || e.name === 'INVITE_REQUIRED') {
                    if (typeof window !== 'undefined') {
                        window.alert(e.message)
                    }
                    await router.replace({ path: '/' })
                    return
                }
            } else {
                errorMessage.value = '加载题目失败，请稍后再试'
            }
        } finally {
            loading.value = false
        }
    }

    function setRef(el: HTMLElement | null) {
        if (!el) return
        const id = el.querySelector('input')?.getAttribute('name') ?? ''
        if (id) refs.set(id, el)
    }
    function highlight(id: string) {
        highlightedId.value = id
        const el = refs.get(id); if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
        setTimeout(() => (highlightedId.value = null), 1200)
    }

    // 分页
    const totalPages = computed(() => Math.max(1, Math.ceil(questions.value.length / pageSize)))
    const currentPageQuestions = computed(() => {
        const start = (currentPage.value - 1) * pageSize
        return questions.value.slice(start, start + pageSize)
    })

    // 量表
    const scaleOptions = [
        { value: 1 as AnswerValue, label: '从不' },
        { value: 2 as AnswerValue, label: '较少' },
        { value: 3 as AnswerValue, label: '一般' },
        { value: 4 as AnswerValue, label: '经常' },
        { value: 5 as AnswerValue, label: '总是' },
    ]

    const isCurrentPageComplete = computed(() =>
        isPageComplete(opts.stage, currentPageQuestions.value.map(q => q.id))
    )

    function onSelect(id: string, v: AnswerValue) { setAnswer(opts.stage, id, v) }
    function getAnswer(id: string) {
        return (opts.stage === 1 ? state.answersStage1 : state.answersStage2)?.[id]
    }

    function scrollToFirstUnanswered() {
        for (const q of currentPageQuestions.value) {
            if (!getAnswer(q.id)) { highlight(q.id); break }
        }
    }

    async function handlePrev() {
        const p = prevStep()
        await router.push({ path: `/test/${variant.value}/step/${p}` })
    }

    const nextLabel = computed(() => {
        if (currentPage.value === totalPages.value && opts.stage === 2) return '提交'
        return '下一页'
    })

    async function handleNext() {
        if (!isCurrentPageComplete.value) { scrollToFirstUnanswered(); return }

        if (currentPage.value < totalPages.value) {
            currentPage.value += 1
            return
        }

        if (opts.stage === 1) {
            const n = nextStep(((STEPS as Record<Variant, readonly StepLite[]>)[variant.value] ?? []).length)
            await router.push({ path: `/test/${variant.value}/step/${n}` })
            return
        }

        // 第二阶段最后一页：提交
        submitting.value = true
        try {
            const payload = buildSubmitPayload()
            await submitTestSession(payload as unknown as any)

            const sid = getSessionId()
            if (sid) await redeemInvite(sid)

            const n = nextStep(((STEPS as Record<Variant, readonly StepLite[]>)[variant.value] ?? []).length)
            await router.push({ path: `/test/${variant.value}/step/${n}` })
        } catch (e) {
            console.error('[useQuestionsStage] submit error', e)
            if (e instanceof Error) {
                errorMessage.value = e.message
            } else {
                errorMessage.value = '提交失败，请稍后再试'
            }
        } finally {
            submitting.value = false
        }
    }

    const currentStepTitle = computed(() => {
        const arr = (STEPS as Record<Variant, readonly StepLite[]>)[variant.value] ?? []
        const item = arr[currentStep.value - 1]
        return item ? (item.titleKey ?? item.key) : `阶段 ${currentStep.value}`
    })

    return {
        // 状态
        loading, submitting, errorMessage,
        currentStep, currentStepTitle, stepItems,
        // 分页
        currentPage, totalPages, currentPageQuestions, isCurrentPageComplete, nextLabel,
        // 题目与作答
        scaleOptions, getAnswer, onSelect, handlePrev, handleNext,
        // 高亮
        highlightedId, setRef,
    }

    /** 构建提交载荷：将 stage1/2 的答案映射到后端需要的字段（riasec / asc） */
    function buildSubmitPayload() {
        const session_id = getSessionId()
        if (!session_id) {
            throw new Error('会话已失效，请重新开始测试')
        }
        const variantValue: Variant = variant.value
        const modeValue: ModeOption = (state.mode as ModeOption) ?? '3+3'

        // 尽量按后端常见结构构造（若后端签名更严格，这里 as unknown as X 兜底）
        const riasec_answers = mapAnswers(state.answersStage1)
        const asc_answers = mapAnswers(state.answersStage2)

        return {
            session_id,
            variant: variantValue,
            grade: state.grade ?? undefined,
            mode: modeValue,
            hobby: state.hobby ?? '',
            riasec_answers,
            asc_answers,
        }
    }

    function mapAnswers(rec: Record<string, AnswerValue> | undefined) {
        if (!rec) return []
        return Object.entries(rec).map(([id, value]) => ({ id, value }))
    }
}

/** 将后端返回归一化为 {stage1, stage2}；修复 (Question|null)[] 的类型 */
function normalize(resp: any): StageQuestions {
    const box = resp?.data ?? resp ?? {}

    function toList(arr: any[]): Question[] {
        return (arr ?? [])
            .map((x: any) => {
                if (typeof x === 'string') return { id: x, text: x }
                if (x && typeof x === 'object') {
                    return {
                        id: String(x.id ?? x.qid ?? x.text ?? Math.random()),
                        text: String(x.text ?? x.title ?? ''),
                    }
                }
                return null
            })
            .filter((v): v is Question => v !== null) // 关键：类型守卫，避免 (Question|null)[]
    }

    const s1 = Array.isArray(box.stage1)
        ? toList(box.stage1)
        : Array.isArray(box.stage1Questions)
            ? toList(box.stage1Questions)
            : null

    const s2 = Array.isArray(box.stage2)
        ? toList(box.stage2)
        : Array.isArray(box.stage2Questions)
            ? toList(box.stage2Questions)
            : null

    if (s1 || s2) return { stage1: s1 ?? [], stage2: s2 ?? [] }

    if (Array.isArray(box.questions)) {
        const all = toList(box.questions)
        const half = Math.ceil(all.length / 2)
        return { stage1: all.slice(0, half), stage2: all.slice(half) }
    }

    if (Array.isArray(box)) {
        const all = toList(box)
        const half = Math.ceil(all.length / 2)
        return { stage1: all.slice(0, half), stage2: all.slice(half) }
    }

    return { stage1: [], stage2: [] }
}
