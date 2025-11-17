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
export type ModeOption = typeof Mode33 | typeof Mode312


export interface CommonResponse {
    ok: boolean
    msg: string | null
}

import {onMounted, onBeforeUnmount} from 'vue'

export interface UseSSEOptions {
    onMsg?: (data: any) => void
    onOpen?: () => void
    onError?: (event: Error) => void
    onClose?:()=>void
}

function eventToError(ev: Event, message = '[SSE] connection error'): Error {
    const err = new Error(message)
    ;(err as any).cause = ev            // 挂在 cause 上，方便调试
    ;(err as any).rawEvent = ev        // 你也可以自定义属性
    return err
}

export function useSubscriptBySSE(
    eventID: string,
    scaleKey: string,
    testType: string,
    options: UseSSEOptions = {},
){
    let es: EventSource | null = null

    onMounted(() => {
        const params = new URLSearchParams({
            scaleKey,
            testType,
        })

        const url = `/api/sub/${eventID}?${params.toString()}`

        es = new EventSource(url)

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
        }

        es.onmessage = (e) => {
            try {
                const data = JSON.parse(e.data)
                if (options.onMsg) {
                    options.onMsg(data)
                }
            } catch (e2) {
                console.warn('[SSE] invalid json', e.data)
                if (options.onError) {
                    const err = e2 instanceof Error ? e2 : new Error('[SSE] invalid json: ' + String(e2))
                    options.onError(err)
                }
            }
        }
    })

    onBeforeUnmount(() => {
        if (es) {
            es.close()
            es = null
        }
    })
}