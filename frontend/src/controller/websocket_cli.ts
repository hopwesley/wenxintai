type Handler<T> = (payload: T) => void

interface StreamMessage<T = any> {
  id?: string
  type: string
  data?: T
}

interface StatusPayload {
  phase?: string
}

interface TokenPayload {
  text?: string
}

interface ProgressPayload {
  tokens?: number
}

interface FinalPayload {
  report_id?: string
}

interface ErrorPayload {
  code?: string
  message?: string
}

export interface StreamClient {
  onStatus(cb: Handler<StatusPayload>): () => void
  onToken(cb: Handler<TokenPayload>): () => void
  onProgress(cb: Handler<ProgressPayload>): () => void
  onFinal(cb: Handler<FinalPayload>): () => void
  onError(cb: Handler<ErrorPayload>): () => void
  close(): void
}

class StreamClientImpl implements StreamClient {
  private readonly assessmentId: string
  private readonly useSSE: boolean
  private eventSource: EventSource | null = null
  private socket: WebSocket | null = null
  private lastEventId: string | undefined
  private reconnectTimer: number | undefined
  private shouldReconnect = true
  private readonly reconnectDelay = 2000

  private statusHandlers: Handler<StatusPayload>[] = []
  private tokenHandlers: Handler<TokenPayload>[] = []
  private progressHandlers: Handler<ProgressPayload>[] = []
  private finalHandlers: Handler<FinalPayload>[] = []
  private errorHandlers: Handler<ErrorPayload>[] = []

  constructor(assessmentId: string) {
    this.assessmentId = assessmentId
    this.useSSE = typeof window !== 'undefined' && typeof window.EventSource !== 'undefined'
    this.open()
  }

  onStatus(cb: Handler<StatusPayload>): () => void {
    this.statusHandlers.push(cb)
    return () => this.removeHandler(this.statusHandlers, cb)
  }

  onToken(cb: Handler<TokenPayload>): () => void {
    this.tokenHandlers.push(cb)
    return () => this.removeHandler(this.tokenHandlers, cb)
  }

  onProgress(cb: Handler<ProgressPayload>): () => void {
    this.progressHandlers.push(cb)
    return () => this.removeHandler(this.progressHandlers, cb)
  }

  onFinal(cb: Handler<FinalPayload>): () => void {
    this.finalHandlers.push(cb)
    return () => this.removeHandler(this.finalHandlers, cb)
  }

  onError(cb: Handler<ErrorPayload>): () => void {
    this.errorHandlers.push(cb)
    return () => this.removeHandler(this.errorHandlers, cb)
  }

  close(): void {
    this.shouldReconnect = false
    if (this.reconnectTimer !== undefined) {
      window.clearTimeout(this.reconnectTimer)
      this.reconnectTimer = undefined
    }
    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
    }
    if (this.socket) {
      this.socket.close()
      this.socket = null
    }
  }

  private open(): void {
    if (!this.shouldReconnect) return
    if (this.useSSE) {
      this.openSSE()
    } else {
      this.openWebSocket()
    }
  }

  private openWebSocket(): void {
    if (typeof window === 'undefined') return
    if (this.socket) {
      this.socket.close()
    }
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const host = window.location.host
    const base = `${protocol}://${host}/ws/assessments/${encodeURIComponent(this.assessmentId)}`
    const url = this.lastEventId ? `${base}?last_event_id=${encodeURIComponent(this.lastEventId)}` : base
    const ws = new WebSocket(url)
    this.socket = ws

    ws.onmessage = (evt) => {
      this.handleMessage(evt.data)
    }

    ws.onclose = () => {
      this.scheduleReconnect()
    }

    ws.onerror = () => {
      this.scheduleReconnect()
    }
  }

  private handleMessage(raw: any): void {
    if (typeof raw !== 'string') return
    try {
      const message = JSON.parse(raw) as StreamMessage
      if (message.id) {
        this.lastEventId = message.id
      }
      this.dispatch(message)
    } catch (error) {
      console.warn('[StreamClient] failed to parse message', error)
    }
  }

  private dispatch(message: StreamMessage): void {
    const data = (message.data ?? {}) as any
    switch (message.type) {
      case 'status':
        this.statusHandlers.forEach((handler) => handler(data as StatusPayload))
        break
      case 'token':
        this.tokenHandlers.forEach((handler) => handler(data as TokenPayload))
        break
      case 'progress':
        this.progressHandlers.forEach((handler) => handler(data as ProgressPayload))
        break
      case 'final':
        this.finalHandlers.forEach((handler) => handler(data as FinalPayload))
        break
      case 'error':
        this.errorHandlers.forEach((handler) => handler(data as ErrorPayload))
        break
      default:
        break
    }
  }

  private scheduleReconnect(): void {
    if (!this.shouldReconnect) return
    if (this.reconnectTimer !== undefined) return
    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
    }
    if (this.socket) {
      this.socket.close()
      this.socket = null
    }
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = undefined
      this.open()
    }, this.reconnectDelay)
  }

  private removeHandler<T>(handlers: Handler<T>[], cb: Handler<T>): void {
    const index = handlers.indexOf(cb)
    if (index >= 0) {
      handlers.splice(index, 1)
    }
  }
}

export function connect(assessmentId: string): StreamClient {
  return new StreamClientImpl(assessmentId)
}
