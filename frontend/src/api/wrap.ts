// 用 Proxy 把 wailsjs 自动生成的所有 Go 方法包装一层：
// - 错误归一化为 AppError（见 errors.ts）
// - 保留原始函数签名与类型推断（TypeScript 透过 Proxy 看到的仍是 typeof Go）
//
// 这样领域 facade 文件只需要做 `export const { X, Y } = Go`，
// 而不必为每个方法写 try/catch boilerplate。

import * as Go from '../../wailsjs/go/main/App'
import { normalizeError } from './errors'

type GoApi = typeof Go

export const Go_: GoApi = new Proxy({} as GoApi, {
  get(_, prop: string) {
    const original = (Go as any)[prop]
    if (typeof original !== 'function') return original
    return async (...args: any[]) => {
      try {
        return await original(...args)
      } catch (e) {
        throw normalizeError(prop, e)
      }
    }
  },
})
