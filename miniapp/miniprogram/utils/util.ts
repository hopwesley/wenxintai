export const formatTime = (date: Date) => {
  const year = date.getFullYear()
  const month = date.getMonth() + 1
  const day = date.getDate()
  const hour = date.getHours()
  const minute = date.getMinutes()
  const second = date.getSeconds()

  return (
    [year, month, day].map(formatNumber).join('/') +
    ' ' +
    [hour, minute, second].map(formatNumber).join(':')
  )
}

const formatNumber = (n: number) => {
  const s = n.toString()
  return s[1] ? s : '0' + s
}

export function getErrorMessage(err: any, fallback: string) {
  if (!err) return fallback
  if (typeof err === 'string') return err
  if (err.errMsg) return String(err.errMsg)     // 微信 API 风格
  if (err.message) return String(err.message)   // Error 实例
  return fallback
}
