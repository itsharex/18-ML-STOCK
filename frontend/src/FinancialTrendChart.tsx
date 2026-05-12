import { useEffect, useMemo, useRef, useState } from 'react'
import * as echarts from 'echarts'
import type { main } from '../wailsjs/go/models'
import { GetFinancialTrends } from '../wailsjs/go/main/App'

interface Props {
  code: string
  name?: string
}

type MetricKey = 'roe' | 'grossMargin' | 'revenueGrowth' | 'cashContent' | 'debtRatio'

interface MetricConfig {
  key: MetricKey
  label: string
  color: string
}

const METRICS: MetricConfig[] = [
  { key: 'roe', label: 'ROE', color: '#3b82f6' },
  { key: 'grossMargin', label: '毛利率', color: '#10b981' },
  { key: 'revenueGrowth', label: '营收增长', color: '#f59e0b' },
  { key: 'cashContent', label: '现金含量', color: '#8b5cf6' },
  { key: 'debtRatio', label: '负债率', color: '#ef4444' },
]

export function FinancialTrendChart({ code, name }: Props) {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstanceRef = useRef<echarts.ECharts | null>(null)
  const [data, setData] = useState<main.FinancialTrendsData | null>(null)
  const [loading, setLoading] = useState(false)
  const [activeKeys, setActiveKeys] = useState<MetricKey[]>(['roe', 'grossMargin', 'revenueGrowth', 'cashContent', 'debtRatio'])

  useEffect(() => {
    if (!code) return
    setLoading(true)
    GetFinancialTrends(code)
      .then((res) => setData(res || null))
      .catch(() => setData(null))
      .finally(() => setLoading(false))
  }, [code])

  useEffect(() => {
    const el = chartRef.current
    if (!el) return
    const instance = echarts.init(el, undefined, { renderer: 'canvas' })
    chartInstanceRef.current = instance
    const timers = [60, 200].map((ms) => setTimeout(() => instance.resize(), ms))
    const handleResize = () => instance.resize()
    window.addEventListener('resize', handleResize)
    return () => {
      timers.forEach(clearTimeout)
      window.removeEventListener('resize', handleResize)
      instance.dispose()
      chartInstanceRef.current = null
    }
  }, [])

  useEffect(() => {
    const instance = chartInstanceRef.current
    if (!instance || !data?.items?.length) return

    const items = [...data.items].reverse()
    const years = items.map((i) => i.year)

    const series = METRICS.filter((m) => activeKeys.includes(m.key)).map((m) => ({
      name: m.label,
      type: 'line' as const,
      smooth: true,
      symbol: 'circle',
      symbolSize: 6,
      lineStyle: { width: 3, color: m.color },
      itemStyle: { color: m.color },
      data: items.map((i) => {
        const v = (i as any)[m.key]
        return v != null ? Number(v) : null
      }),
    }))

    const option: echarts.EChartsOption = {
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(15,23,42,0.95)',
        borderColor: 'rgba(148,163,184,0.2)',
        textStyle: { color: '#e2e8f0' },
        formatter: (params: any) => {
          let html = `<div style="font-weight:600;margin-bottom:4px;">${params[0]?.axisValue}年</div>`
          params.forEach((p: any) => {
            const val = p.value != null ? `${p.value.toFixed(2)}%` : '-'
            html += `<div style="display:flex;align-items:center;gap:6px;margin:2px 0;">
              <span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:${p.color};"></span>
              <span style="flex:1;">${p.seriesName}</span>
              <span style="font-weight:600;">${val}</span>
            </div>`
          })
          return html
        },
      },
      legend: { show: false },
      grid: { left: 48, right: 24, top: 24, bottom: 32 },
      xAxis: {
        type: 'category',
        data: years,
        axisLine: { lineStyle: { color: 'rgba(148,163,184,0.3)' } },
        axisLabel: { color: '#94a3b8' },
        axisTick: { show: false },
      },
      yAxis: {
        type: 'value',
        axisLine: { show: false },
        axisLabel: { color: '#94a3b8', formatter: '{value}%' },
        splitLine: { lineStyle: { color: 'rgba(148,163,184,0.1)' } },
      },
      series,
    }

    instance.setOption(option, true)
    requestAnimationFrame(() => instance.resize())
  }, [data, activeKeys])

  const toggleMetric = (key: MetricKey) => {
    setActiveKeys((prev) => {
      if (prev.includes(key)) return prev.filter((k) => k !== key)
      return [...prev, key]
    })
  }

  const hasData = useMemo(() => (data?.items?.length || 0) > 0, [data])

  const isLight = document.body.classList.contains('light')
  const titleColor = isLight ? '#1e293b' : '#e2e8f0'
  const metaColor = isLight ? '#64748b' : '#64748b'
  const borderColor = isLight ? 'rgba(148,163,184,0.25)' : 'rgba(148,163,184,0.15)'
  const bgColor = isLight ? 'rgba(241,245,249,0.6)' : 'rgba(15,23,42,0.4)'
  const maskBg = isLight ? 'rgba(241,245,249,0.8)' : 'rgba(20,27,36,0.6)'
  const inactiveText = isLight ? '#64748b' : '#94a3b8'

  return (
    <div style={{ margin: '16px 0', padding: '12px 16px', border: `1px solid ${borderColor}`, borderRadius: 8, background: bgColor }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}>
        <h4 style={{ margin: 0, fontSize: 14, fontWeight: 600, color: titleColor }}>
          {name || code} 财务指标趋势
        </h4>
        <span style={{ fontSize: 12, color: metaColor }}>数据来源：本地财报（最近5年）</span>
      </div>

      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 12 }}>
        {METRICS.map((m) => {
          const active = activeKeys.includes(m.key)
          return (
            <button
              key={m.key}
              onClick={() => toggleMetric(m.key)}
              style={{
                minWidth: 80,
                padding: '4px 10px',
                whiteSpace: 'nowrap',
                borderRadius: 6,
                border: '1px solid',
                borderColor: active ? m.color : borderColor,
                background: active ? `${m.color}20` : 'transparent',
                color: active ? m.color : inactiveText,
                fontSize: 12,
                fontWeight: 500,
                cursor: 'pointer',
                transition: 'all .15s ease',
              }}
            >
              {m.label}
            </button>
          )
        })}
      </div>

      <div style={{ position: 'relative', width: '100%', height: 280 }}>
        <div ref={chartRef} style={{ width: '100%', height: '100%' }} />
        {(loading || !hasData) && (
          <div
            style={{
              position: 'absolute',
              inset: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: inactiveText,
              fontSize: 14,
              background: maskBg,
              backdropFilter: 'blur(2px)',
              borderRadius: 8,
            }}
          >
            {loading ? '加载中...' : '暂无财务数据，请先下载财报'}
          </div>
        )}
      </div>
    </div>
  )
}
