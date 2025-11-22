export const TestTypeBasic = "basic"
export const TestTypePro = "pro"
export const TestTypeSchool = "school"

export const StageBasic = "basic-info"
export const StageReport = "report"
export const StageRiasec = "riasec"
export const StageAsc = "asc"
export const StageOcean = "ocean"
export const StageMotivation = "motivation"

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

export interface CommonResponse {
    ok: boolean
    msg: string | null
    next_route: string | null
    next_route_index: number
}

import {onMounted, onBeforeUnmount, getCurrentInstance} from 'vue'
import type { Router } from 'vue-router'

export interface UseSSEOptions {
    onMsg?: (data: any) => void
    onOpen?: () => void
    onError?: (event: Error) => void
    onClose?: () => void
    onDone?: (question: string) => void

    // 新增一个可选配置：是否自动在 mounted 时启动
    autoStart?: boolean
}

function eventToError(ev: Event, message = '[SSE] connection error'): Error {
    console.log(ev)
    const err = new Error(message)
    ;(err as any).cause = ev      // 挂在 cause 上，方便调试
    ;(err as any).rawEvent = ev   // 你也可以自定义属性
    return err
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
    businessType: string,
    stage: string,
) {
    if (!businessType || !stage) return

    // 特殊路由：基础信息
    if (stage === StageBasic) {
        return router.push({
            name: 'test-basic-info',
            params: { typ: businessType },
        })
    }

    // 特殊路由：测评报告
    if (stage === StageReport) {
        return router.push({
            name: 'test-report',
            params: { typ: businessType },
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

export function useTestCommon() {
    return {
        // 业务类型
        TestTypeBasic,
        TestTypePro,
        TestTypeSchool,

        // 各个阶段的路由 key
        StageBasic,
        StageReport,
        StageRiasec,
        StageAsc,
        StageOcean,
        StageMotivation,

        // 模式 & 量表
        Mode33,
        Mode312,
        scaleOptions,
    }
}
