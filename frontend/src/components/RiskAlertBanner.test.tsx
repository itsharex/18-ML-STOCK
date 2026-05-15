import { describe, it, expect } from 'vitest'
import { render, fireEvent } from '@testing-library/react'
import { RiskAlertBanner } from './RiskAlertBanner'

describe('RiskAlertBanner', () => {
  it('alert 为 null 时不渲染', () => {
    const { container } = render(<RiskAlertBanner alert={undefined} />)
    expect(container.firstChild).toBeNull()
  })

  it('low 等级时不渲染', () => {
    const alert = {
      level: 'low',
      primaryMsg: '该股票存在 1 项低风险信号',
      flags: [],
    } as any
    const { container } = render(<RiskAlertBanner alert={alert} />)
    expect(container.firstChild).toBeNull()
  })

  it('high 等级时渲染并显示精简消息', () => {
    const alert = {
      level: 'high',
      primaryMsg: '该股票存在 3 项高风险信号',
      flags: [
        { level: 'high', name: '大股东高比例质押', format: '股权质押 80%' },
      ],
    } as any
    const { getByText } = render(<RiskAlertBanner alert={alert} />)
    expect(getByText('3项高风险')).toBeDefined()
  })

  it('点击 header 展开/收起 flags', () => {
    const alert = {
      level: 'high',
      primaryMsg: '该股票存在 2 项高风险信号',
      flags: [
        { level: 'high', name: '大股东高比例质押', format: '股权质押 80%' },
        { level: 'medium', name: '毛利率下滑', format: '毛利率 15%' },
      ],
    } as any
    const { container, getByText } = render(<RiskAlertBanner alert={alert} />)

    const header = container.querySelector('.risk-alert-header')
    expect(header).not.toBeNull()

    // 初始收起状态
    expect(container.querySelector('.risk-alert-body')).toBeNull()

    // 点击展开
    fireEvent.click(header!)
    expect(getByText('大股东高比例质押：股权质押 80%')).toBeDefined()
    expect(getByText('毛利率下滑：毛利率 15%')).toBeDefined()

    // 再次点击收起
    fireEvent.click(header!)
    expect(container.querySelector('.risk-alert-body')).toBeNull()
  })

  it('无 flags 时不显示展开箭头', () => {
    const alert = {
      level: 'medium',
      primaryMsg: '该股票存在 1 项中风险信号',
      flags: [],
    } as any
    const { container } = render(<RiskAlertBanner alert={alert} />)
    expect(container.querySelector('.risk-alert-toggle')).toBeNull()
  })
})
