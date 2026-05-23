// 统一错误归一化：所有 facade 抛出的错误都符合 AppError 形状，
// UI 层不需要 try/catch 每个 Go 调用、不需要自己判断错误文本。

export interface AppError {
  op: string         // 操作名（与 Go 方法名一致），用于诊断
  message: string    // 用户可读消息
  code?: string      // 业务错误码（如有），便于代码分支
  retryable: boolean // 网络/超时错误为 true，业务错误为 false
  cause?: unknown    // 原始 error，调试用
}

/**
 * 把任意 throw 出来的值规范化成 AppError。
 * 网络错误（含 timeout / context canceled / EOF）标记为 retryable=true。
 */
export function normalizeError(op: string, e: unknown): AppError {
  if (isAppError(e)) {
    return e
  }
  if (e instanceof Error) {
    const msg = e.message
    const retryable = /timeout|network|connection refused|EOF|context canceled|context deadline/i.test(msg)
    return { op, message: msg, retryable, cause: e }
  }
  return { op, message: String(e), retryable: false, cause: e }
}

export function isAppError(e: unknown): e is AppError {
  return typeof e === 'object' && e !== null
    && 'op' in e && 'message' in e && 'retryable' in e
}
