import {onMounted, onBeforeUnmount, getCurrentInstance, computed, ref, reactive} from 'vue'
import type {Router} from 'vue-router'
import {API_PATHS, apiRequest} from "@/api";

export const TestTypeBasic = "basic"
export const TestTypePro = "pro"
export const TestTypeAdv = "adv"
export const TestTypeSchool = "school"
export type PlanKey = typeof TestTypeBasic | typeof TestTypePro | typeof TestTypeSchool | typeof TestTypeAdv

export const StageBasic = "basic-info"
export const StageReport = "report"
export const StageRiasec = "RIASEC"
export const StageAsc = "ASC"
export const StageOcean = "OCEAN"
export const StageMotivation = "MOTIVATION"

export interface TestFlowStep {
    stage: string      // "basic-info" / "riasec" / "asc" / ...
    title: string      // 展示给用户看的中文文案，如“基础信息”“兴趣测试”
}

export const Mode33 = '3+3'
export const Mode312 = '3+1+2'
export type ModeOption = '3+3' | '3+1+2'
export type AnswerValue = 1 | 2 | 3 | 4 | 5
export const scaleOptions = [
    {value: 1 as AnswerValue, label: '从不'},
    {value: 2 as AnswerValue, label: '较少'},
    {value: 3 as AnswerValue, label: '一般'},
    {value: 4 as AnswerValue, label: '经常'},
    {value: 5 as AnswerValue, label: '总是'},
]

export const subjectLabelMap: Record<string, string> = {
    PHY: '物理',
    CHE: '化学',
    BIO: '生物',
    GEO: '地理',
    HIS: '历史',
    POL: '政治',
}

export interface CommonResponse {
    ok: boolean
    msg: string | null
    next_route: string | null
    next_route_index: number
}

export interface UseSSEOptions {
    onMsg?: (data: any) => void
    onOpen?: () => void
    onError?: (event: Error) => void
    onClose?: () => void
    onDone?: (question: string) => void
    autoStart?: boolean
}

export interface PlanInfo {
    key: PlanKey
    name: string
    price: number
    desc: string
    tag?: string
    has_paid?: boolean
}

function eventToError(ev: Event, message = '[SSE] connection error'): Error {
    console.log(ev)
    const err = new Error(message)
    ;(err as any).cause = ev      // 挂在 cause 上，方便调试
    ;(err as any).rawEvent = ev   // 你也可以自定义属性
    return err
}

export function isValidChinaMobile(input: string): boolean {
    if (!input) return false

    // 去掉空格和连字符等
    let phone = input.replace(/[\s-]/g, '')

    // 处理前缀：+86 / 0086 / 86
    if (phone.startsWith('+86')) {
        phone = phone.slice(3)
    } else if (phone.startsWith('0086')) {
        phone = phone.slice(4)
    } else if (phone.startsWith('86') && phone.length > 11) {
        phone = phone.slice(2)
    }

    // 现在应该只剩下 11 位本地号码
    if (!/^\d{11}$/.test(phone)) {
        return false
    }

    // 中国大陆手机号段：1 开头，第 2 位 3-9
    return /^1[3-9]\d{9}$/.test(phone)
}

export function useSubscriptBySSE(
    url: string,
    options: UseSSEOptions = {},
) {
    const {autoStart = true} = options
    let es: EventSource | null = null

    const start = () => {
        if (es) {
            return
        }

        es = new EventSource(url)

        es.addEventListener('done', (ev: MessageEvent) => {
            console.log("done message", ev)
            if (options.onDone) {
                options.onDone(ev.data as string)
            }
            stop()
        })

        es.addEventListener('app-error', (ev: MessageEvent) => {
            console.log("app error", ev)
            const msg = (ev.data as string) || '服务器返回未知错误'
            if (options.onError) {
                options.onError(new Error(msg))
            }
            stop()
        })

        es.onopen = () => {
            console.log('[SSE] connection opened')
            if (options.onOpen) {
                options.onOpen()
            }
        }

        es.onerror = (ev) => {
            console.error('[SSE] error', ev)
            if (options.onError) {
                const err = eventToError(ev)
                options.onError(err)
            }
            stop()
        }

        es.onmessage = (e: MessageEvent) => {
            if (options.onMsg) {
                options.onMsg(e.data)
            }
        }
    }

    const stop = () => {
        console.log('[SSE] connection closed')
        if (es) {
            es.close()
            es = null
        }
    }

    const instance = getCurrentInstance()

    if (instance) {
        if (autoStart) {
            onMounted(() => {
                start()
            })
        }

        onBeforeUnmount(() => {
            stop()
        })
    }

    return {
        start,
        stop,
    }
}

export function pushStageRoute(
    router: Router,
    businessType: PlanKey,
    stage: string,
) {
    if (!businessType || !stage) return

    // 特殊路由：基础信息
    if (stage === StageBasic) {
        return router.push({
            name: 'test-basic-info',
            params: {typ: businessType},
        })
    }

    // 特殊路由：测评报告
    if (stage === StageReport) {
        return router.push({
            name: 'test-report',
            params: {typ: businessType},
        })
    }

    // 其它阶段：统一走 test-stage
    return router.push({
        name: 'test-stage',
        params: {
            businessType,
            testStage: stage,
        },
    })
}

export function useSseLogs(
    maxLines = 8,
    minChunkLen = 20,
) {
    const logLines = ref<string[]>([])
    const rawMessage = ref('')

    // 是否已经插入过“第一行固定提示文案”
    const hasSeedLine = ref(false)

    const truncatedLatestMessage = computed(() => logLines.value)

    function pushLine(text: string) {
        const content = text.trim()
        if (!content) return

        // 所有行统一增加前缀 AI>
        const line = `AI> ${content}`
        logLines.value.push(line)

        if (logLines.value.length > maxLines) {
            logLines.value.splice(0, logLines.value.length - maxLines)
        }
    }

    function handleSseMsg(chunk: string) {
        // 第一次收到后端消息时，先插入一条固定的“人类可读”的提示行
        if (!hasSeedLine.value) {
            pushLine('已收到你的回答，正在加载试题模板…')
            hasSeedLine.value = true
        }

        rawMessage.value += chunk

        // 还没攒够长度，就先不刷到日志窗口里
        if (rawMessage.value.length < minChunkLen) {
            return
        }

        // 把当前累计内容刷成一条新日志
        const flushed = rawMessage.value
        rawMessage.value = ''

        pushLine(flushed)
    }

    function resetLogs() {
        logLines.value = []
        rawMessage.value = ''
        hasSeedLine.value = false
    }

    return {
        logLines,
        truncatedLatestMessage,
        rawMessage,
        handleSseMsg,
        resetLogs,
    }
}

export const DEFAULT_HOBBIES: string[] = [
    // 体育类
    '篮球',
    '足球',
    '羽毛球',
    '跑步',
    '游泳',
    '乒乓球',
    '健身',

    // 艺术类
    '音乐',
    '绘画',
    '舞蹈',
    '摄影',
    '书法',
    '写作',

    // 科技类
    '编程',
    '机器人',
    '科学实验',
    '电子制作',
    '下棋',

    // 生活方式类
    '旅行',
    '美食',
    '志愿活动',
    '阅读',
    '看电影',
    '园艺',
]

export const basicPlan: PlanInfo = {
    key: TestTypeBasic,
    name: '基础版',
    price: 29.9,
    desc: '组合推荐 + 学科优势评估',
}

const proPlan: PlanInfo = {
    key: TestTypePro,
    name: '专业版',
    price: 49.9,
    desc: '基础版+更加全面的参数解读',
    tag: '推荐',
}

const advPlan: PlanInfo = {
    key: TestTypeAdv,
    name: '增强版',
    price: 79.9,
    desc: '专业版 +专业选择推荐+职业规划建议',
}

const schoolPlan: PlanInfo = {
    key: TestTypeSchool,
    name: '校本定制版',
    price: 59.9,
    desc: '结合校园真是数据，精准报告，多维对比',
}

export const currentProductsMap = reactive<Record<PlanKey, PlanInfo>>({
    [TestTypeBasic]: basicPlan,
    [TestTypePro]: proPlan,
    [TestTypeAdv]: advPlan,
    [TestTypeSchool]: schoolPlan,
})

export async function loadProducts() {
    try {
        const res = await apiRequest<PlanInfo[]>(API_PATHS.LOAD_PRODUCTS, {
            method: 'GET',
        })

        if (!Array.isArray(res) || res.length === 0) {
            return
        }

        for (const p of res) {
            currentProductsMap[p.key] = p
        }

    } catch (err) {
        console.error('loadProducts failed, fallback to local planMap:', err)
    }
}