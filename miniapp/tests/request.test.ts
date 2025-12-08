import { request } from '../miniprogram/utils/request'
import * as store from '../miniprogram/utils/store'

describe('request wrapper', () => {
  const setAuthSpy = jest.spyOn(store, 'setAuthInfo')

  beforeEach(() => {
    setAuthSpy.mockClear()
    ;(global as any).wx = {
      request: jest.fn(),
      getStorageSync: jest.fn(() => ''),
      setStorageSync: jest.fn(),
    }
  })

  it('resolves data and stores cookie', async () => {
    ;(wx.request as jest.Mock).mockImplementation((opts: any) => {
      opts.success({ statusCode: 200, data: { ok: true }, header: { 'Set-Cookie': 'token=1' } })
    })
    const res = await request<{ ok: boolean }>({ url: '/api/demo' })
    expect(res.ok).toBe(true)
    expect(setAuthSpy).toHaveBeenCalledWith(expect.objectContaining({ cookie: 'token=1' }))
  })

  it('rejects on http error', async () => {
    ;(wx.request as jest.Mock).mockImplementation((opts: any) => {
      opts.success({ statusCode: 500, data: { err: true }, header: {} })
    })
    await expect(request({ url: '/api/fail' })).rejects.toMatchObject({ status: 500 })
  })

  it('rejects on network failure', async () => {
    ;(wx.request as jest.Mock).mockImplementation((opts: any) => {
      opts.fail({ errMsg: 'offline' })
    })
    await expect(request({ url: '/api/offline' })).rejects.toMatchObject({ message: 'offline' })
  })
})
