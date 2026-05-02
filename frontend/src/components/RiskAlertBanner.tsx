import { useState } from 'react'
import { analyzer } from '../../wailsjs/go/models'

interface RiskAlertBannerProps {
  alert?: analyzer.RiskAlertSummary
}

export function RiskAlertBanner({ alert }: RiskAlertBannerProps) {
  const [expanded, setExpanded] = useState(false)
  if (!alert || alert.level === 'low') return null

  const isHigh = alert.level === 'high'

  // 提取数字：如 "该股票存在 3 项中风险信号" → "3项中风险"
  const shortMsg = alert.primaryMsg
    .replace(/^.*存在\s+(\d+)\s+项(高风险|中风险|中高风险)信号.*$/, '$1项$2')
    .replace(/^.*🟢\s*/, '')

  return (
    <div className={`risk-alert-banner ${isHigh ? 'risk-alert-high' : 'risk-alert-medium'}`}>
      <div
        className="risk-alert-header"
        onClick={() => setExpanded(!expanded)}
        style={{ cursor: alert.flags && alert.flags.length > 0 ? 'pointer' : 'default' }}
      >
        <span>{shortMsg}</span>
        {alert.flags && alert.flags.length > 0 && (
          <span className={`risk-alert-toggle ${expanded ? 'expanded' : ''}`}>›</span>
        )}
      </div>
      {expanded && alert.flags && alert.flags.length > 0 && (
        <div className="risk-alert-body">
          {alert.flags.map((f: any, i: number) => (
            <div key={i} className={`risk-alert-flag risk-alert-flag-${f.level}`}>
              <span>{f.level === 'high' ? '🔴' : '🟡'}</span>
              <span>{f.name}{f.format ? `：${f.format}` : ''}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
