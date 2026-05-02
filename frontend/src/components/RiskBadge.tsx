interface RiskBadgeProps {
  level?: string
  size?: 'small' | 'medium'
}

export function RiskBadge({ level = 'low', size = 'small' }: RiskBadgeProps) {
  if (level === 'low') return null

  const color = level === 'high' ? '#ef4444' : '#f59e0b'
  const label = level === 'high' ? '高风险' : '中风险'
  const fontSize = size === 'small' ? 10 : 12
  const padding = size === 'small' ? '1px 5px' : '2px 8px'

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 2,
        padding,
        borderRadius: 4,
        background: color + '18',
        color,
        fontSize,
        fontWeight: 600,
        whiteSpace: 'nowrap',
      }}
    >
      <span style={{ fontSize: fontSize + 2 }}>{level === 'high' ? '🔴' : '🟡'}</span>
      {size === 'medium' && label}
    </span>
  )
}
