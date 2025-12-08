import { connectWebSocket } from '../miniprogram/utils/websocket'

const buildSocketMock = () => {
  const listeners = {
    open: [] as Function[],
    message: [] as Function[],
    error: [] as Function[],
    close: [] as Function[],
  }
  const socket = {
    send: jest.fn(),
    close: jest.fn(),
    onOpen: (cb: Function) => listeners.open.push(cb),
    onMessage: (cb: Function) => listeners.message.push(cb),
    onError: (cb: Function) => listeners.error.push(cb),
    onClose: (cb: Function) => listeners.close.push(cb),
  }
  return { listeners, socket }
}

describe('websocket manager', () => {
  beforeEach(() => {
    jest.useFakeTimers()
    const mock = buildSocketMock()
    ;(global as any).wx = {
      connectSocket: jest.fn(() => mock.socket as any),
    }
    ;(global as any).__socketMock = mock
  })

  it('dispatches parsed messages', () => {
    const onData = jest.fn()
    const onDone = jest.fn()
    const ws = connectWebSocket('/ws', { onData, onDone })
    const { listeners } = (global as any).__socketMock
    listeners.message.forEach((fn: Function) => fn({ data: JSON.stringify({ type: 'data', payload: { v: 1 } }) }))
    expect(onData).toHaveBeenCalledWith({ v: 1 })
    listeners.message.forEach((fn: Function) => fn({ data: JSON.stringify({ type: 'done', payload: { done: true } }) }))
    expect(onDone).toHaveBeenCalled()
    ws.close()
  })

  it('sends heartbeat', () => {
    const ws = connectWebSocket('/ws', {})
    const { socket } = (global as any).__socketMock
    ws.sendPing(10)
    jest.advanceTimersByTime(30)
    expect(socket.send).toHaveBeenCalled()
    ws.close()
  })
})
