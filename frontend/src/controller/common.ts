export const TestTypeBasic = "basic"
export const TestTypePro = "pro"
export const TestTypeSchool = "school"

export const StageBasic = "basic-info"
export const StageReport = "report"
export const StageRiasec = "riasec"
export const StageAsc = "asc"
export const StageOcean = "ocean"
export const StageMotivation = "motivation"

export const Mode33 = '3+3'
export const Mode312 = '3+1+2'
export type ModeOption = '3+3' | '3+1+2'


export interface CommonResponse {
    ok: boolean
    msg: string | null
}

import {onMounted, onBeforeUnmount} from 'vue'

export interface UseSSEOptions {
    onMsg?: (data: any) => void
    onOpen?: () => void
    onError?: (event: Error) => void
    onClose?: () => void
    onDone?: (question: string) => void
}

function eventToError(ev: Event, message = '[SSE] connection error'): Error {
    console.log(ev)
    const err = new Error(message)
    ;(err as any).cause = ev            // 挂在 cause 上，方便调试
    ;(err as any).rawEvent = ev        // 你也可以自定义属性
    return err
}

export function useSubscriptBySSE(
    eventID: string,
    businessType: string,
    testType: string,
    options: UseSSEOptions = {},
) {
    let es: EventSource | null = null

    onMounted(() => {
        const params = new URLSearchParams({
            business_type:businessType,
            test_type:testType,
        })

        const url = `/api/sub/${eventID}?${params.toString()}`

        es = new EventSource(url)

        es.addEventListener('done', (ev) => {
            if (options.onDone) {
                options.onDone(ev.data as string)
            }
            stop()
        });

        es.addEventListener('app-error', (ev) => {
            const msg = ev.data || '服务器返回未知错误'
            if (options.onError) {
                options.onError(new Error(msg))
            }
            stop();
        });


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

        es.onmessage = (e) => {
            if (options.onMsg) {
                options.onMsg(e.data)
            }
        }

    })

    onBeforeUnmount(() => {
        stop()
    })
    const stop = () => {
        if (es) {
            es.close()
            es = null
        }
    }

    return
}