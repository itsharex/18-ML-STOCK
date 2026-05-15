import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { RiskBadge } from './RiskBadge'

describe('RiskBadge', () => {
  it('low 风险等级返回 null', () => {
    const { container } = render(<RiskBadge level="low" />)
    expect(container.firstChild).toBeNull()
  })

  it('high 风险等级显示红色标签', () => {
    const { container } = render(<RiskBadge level="high" size="medium" />)
    const badge = container.firstChild as HTMLElement
    expect(badge).not.toBeNull()
    expect(badge.textContent).toContain('高风险')
    expect(badge.textContent).toContain('🔴')
  })

  it('medium 风险等级显示黄色标签', () => {
    const { container } = render(<RiskBadge level="medium" size="medium" />)
    const badge = container.firstChild as HTMLElement
    expect(badge).not.toBeNull()
    expect(badge.textContent).toContain('中风险')
    expect(badge.textContent).toContain('🟡')
  })

  it('small 尺寸不显示文字标签', () => {
    const { container } = render(<RiskBadge level="high" size="small" />)
    const badge = container.firstChild as HTMLElement
    expect(badge).not.toBeNull()
    expect(badge.textContent).not.toContain('高风险')
  })

  it('默认 props 为 low + small', () => {
    const { container } = render(<RiskBadge />)
    expect(container.firstChild).toBeNull()
  })
})
