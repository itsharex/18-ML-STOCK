import { describe, it, expect, beforeEach } from 'vitest'
import { loadSettings, saveSettings, DEFAULT_SETTINGS } from './Settings'

describe('Settings 工具函数', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('loadSettings 无缓存时返回默认值', () => {
    const s = loadSettings()
    expect(s.theme).toBe('dark')
    expect(s.klineDefaultRange).toBe('6m')
    expect(s.autoCheckUpdate).toBe(true)
    expect(s.reportYears).toBe(5)
  })

  it('saveSettings 和 loadSettings 能正确持久化', () => {
    const custom = { ...DEFAULT_SETTINGS, theme: 'light' as const, reportYears: 10 }
    saveSettings(custom)
    const loaded = loadSettings()
    expect(loaded.theme).toBe('light')
    expect(loaded.reportYears).toBe(10)
    // 其他字段保持默认值
    expect(loaded.klineDefaultRange).toBe('6m')
  })

  it('loadSettings 能合并部分缓存', () => {
    localStorage.setItem('stockfinlens-settings-v1', JSON.stringify({ theme: 'system' }))
    const s = loadSettings()
    expect(s.theme).toBe('system')
    expect(s.klineDefaultRange).toBe('6m') // 未被覆盖的字段仍为默认值
  })

  it('loadSettings 遇到非法 JSON 时回退到默认值', () => {
    localStorage.setItem('stockfinlens-settings-v1', 'not-json')
    const s = loadSettings()
    expect(s.theme).toBe('dark')
  })
})
