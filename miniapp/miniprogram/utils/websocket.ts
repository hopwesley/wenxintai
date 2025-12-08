export type WebSocketMessageType = 'data' | 'done' | 'error'

export interface WebSocketCallbacks {
  onData?: (payload: any) => void
  onDone?: (payload: any) => void
  onError?: (message: string) => void
  onLog?: (message: string) => void
}

export interface ManagedSocket {
  sendPing: (intervalMs?: number) => void
  close: () => void
}

export const connectWebSocket = (
  url: string,
  callbacks: WebSocketCallbacks,
): ManagedSocket => {
  const socketTask = wx.connectSocket({ url })
  let pingTimer: WechatMiniprogram.Timer | undefined

  const log = (message: string) => {
    if (callbacks.onLog) callbacks.onLog(message)
  }

  socketTask.onOpen(() => {
    log('WebSocket connected')
  })

  socketTask.onError((error) => {
    log(`WebSocket error: ${JSON.stringify(error)}`)
    if (callbacks.onError) callbacks.onError(error.errMsg)
  })

  socketTask.onClose(() => {
    if (pingTimer) {
      clearInterval(pingTimer)
    }
    log('WebSocket closed')
  })

  socketTask.onMessage((res) => {
    try {
      const parsed = JSON.parse(res.data as string)
      const { type, payload } = parsed as { type: WebSocketMessageType; payload: any }
      if (type === 'data' && callbacks.onData) callbacks.onData(payload)
      if (type === 'done' && callbacks.onDone) callbacks.onDone(payload)
      if (type === 'error' && callbacks.onError) callbacks.onError(payload || 'Server error')
    } catch (err: any) {
      const message = err?.message || 'Invalid WebSocket message'
      log(message)
      if (callbacks.onError) callbacks.onError(message)
    }
  })

  const sendPing = (intervalMs = 15000) => {
    if (pingTimer) {
      clearInterval(pingTimer)
    }
    pingTimer = setInterval(() => {
      socketTask.send({ data: 'ping' })
    }, intervalMs)
  }

  const close = () => {
    if (pingTimer) clearInterval(pingTimer)
    socketTask.close({})
  }

  return { sendPing, close }
}
