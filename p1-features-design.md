# P1 大功能/重构 实现方案

> 本文档为 F1/F2/F4/F5/O2 五个 P1 项的详细设计方案。
> 在编码前需用户确认，确认后逐项执行。

---

## 一、F1: 季报/TTM 支持

### 1.1 问题现状
- `analyzer/data.go:141` 硬过滤非年报数据，`FinancialData.Years` 只有年报
- 报告中只有"近5年趋势"，缺乏季度滚动视角
- 投研时效性差：年报发布滞后（次年4月），用户无法及时感知季度变化

### 1.2 实现逻辑（三层口径）

**核心原则**：不破坏现有 18 步年报分析流程，新增季度数据作为独立分析层。

```
┌─────────────────────────────────────────────────────────────┐
│  口径A: 年报长期质量（现有 18 步财报透镜，不变）               │
│  口径B: 季报滚动预警（新增：季度环比/同比趋势 + 恶化检测）      │
│  口径C: TTM 估值（新增：滚动 4 季度合并计算关键指标）           │
└─────────────────────────────────────────────────────────────┘
```

**步骤1: 数据层改造**
- `analyzer/data.go`：`LoadFinancialData` 不再过滤季报，保留所有日期格式（`YYYY-MM-DD`）
- `FinancialData` 新增字段：
  ```go
  type FinancialData struct {
      // ... 现有字段 ...
      Quarters           []string                        // 所有季度日期（降序）
      QuarterBalanceSheet map[string]map[string]float64  // 季度数据与年报共用存储结构
      // IncomeStatement / CashFlow 同理，键为 "2024-03-31" 格式
  }
  ```
- `extractYearsFloat` 改名为 `extractPeriodsFloat`，返回所有期间；新增 `extractYearsOnly` 用于现有年报逻辑

**步骤2: 季报滚动预警模块（新增 `analyzer/quarterly.go`）**
- `BuildQuarterlyAlert(data *FinancialData) *QuarterlyAlert`
- 检测逻辑：
  - 营收季度环比连续 2 季度下滑 → "营收环比持续萎缩"
  - 净利润季度环比连续 2 季度下滑 → "盈利环比持续恶化"
  - 经营现金流季度为负 → "季度经营现金流告负"
  - 毛利率季度环比下降 > 3pct → "毛利率季度承压"
  - 应收账款/营收比率季度上升 → "回款质量季度恶化"
- 输出：`QuarterlyAlert` 结构，包含 `Level`（warning/danger）和 `Items []QuarterlyAlertItem`

**步骤3: TTM 计算模块（新增 `analyzer/ttm.go`）**
- `BuildTTMMetrics(data *FinancialData) *TTMMetrics`
- 取最近 4 个季度的利润表/现金流数据累加（滚动 TTM）
- 计算指标：
  - TTM 营业收入、TTM 净利润
  - TTM 经营现金流
  - TTM ROE（TTM 净利润 / 最新净资产）
  - TTM 净利率、TTM 现金流/净利润比率
- 如果有年报数据（4个季度正好覆盖一年），优先用年报；否则用季报累加

**步骤4: 报告展示**
- 在 `report.go` 中新增 `writeModuleTTM` 模块（插在模块2和模块3之间）
- 展示内容：
  ```markdown
  ## 模块X: 季度滚动与 TTM 透视

  ### 季度预警
  | 指标 | 最新季度 | 上季度 | 变化 | 状态 |
  |------|----------|--------|------|------|
  | 营业收入 | 12.5亿 | 15.2亿 | -17.8% | 🔴 环比萎缩 |
  | 毛利率 | 28.5% | 31.2% | -2.7pct | 🟡 承压 |

  ### TTM 滚动指标（最近4季度）
  | 指标 | TTM值 | 去年同期 | 同比 | 说明 |
  |------|-------|----------|------|------|
  | TTM 营收 | 58.3亿 | 62.1亿 | -6.1% | 滚动口径 |
  | TTM 净利润 | 8.2亿 | 10.5亿 | -21.9% | 🔴 显著下滑 |
  | TTM 经营现金流 | 9.1亿 | 7.8亿 | +16.7% | 🟢 现金改善 |
  | TTM ROE | 12.3% | 15.8% | -3.5pct | 低于 15% 阈值 |
  ```

**步骤5: 前端展示**
- `FinancialTrendDrawer.tsx` 中增加"季度视图"Tab，与现有的年度趋势图并列
- 季度视图用柱状图展示最近 8 个季度的营收/净利润/毛利率

### 1.3 涉及文件
- `analyzer/data.go` — 加载逻辑改造
- `analyzer/types.go` — 新增 QuarterlyAlert、TTMMetrics 类型
- `analyzer/quarterly.go` — 新增（季报预警逻辑）
- `analyzer/ttm.go` — 新增（TTM 计算逻辑）
- `analyzer/report.go` — 新增报告模块
- `frontend/src/FinancialTrendDrawer.tsx` — 季度视图 Tab
- `frontend/wailsjs/go/models.ts` — 新增类型自动生成

### 1.4 风险评估
| 风险点 | 等级 | 缓解措施 |
|--------|------|----------|
| 季报和年报键格式不一致导致混乱 | 中 | 统一用 `YYYY-MM-DD` 键，提取年份/季度时用工具函数 |
| 累加 4 个季度时数据缺失 | 中 | 缺失时 TTM 标记为不完整，降级展示 |
| 现有 18 步分析受季度数据干扰 | 低 | 完全隔离，年报逻辑不变，仅新增模块读取季度数据 |

### 1.5 工作量估计
- 后端数据改造 + 新模块：~2 天
- 报告渲染 + 前端图表：~1 天
- 测试：~0.5 天

---

## 二、F2: 审计意见自动解析

### 2.1 问题现状
- `analyzer/steps.go:16` 全部返回"请查询年报确认"
- `downloader/auditor.go` 已有 `FetchAuditorHistory`，但仅用于外部风险（审计机构变更），未填充审计意见

### 2.2 实现逻辑

**核心原则**：复用现有 cninfo 公告查询能力，扩展审计意见解析。

**步骤1: 扩展审计数据结构**
- `analyzer/types.go` 新增：
  ```go
  type AuditOpinion struct {
      Year        string // "2023"
      Opinion     string // "标准无保留意见" / "保留意见" / "无法表示意见" / "否定意见"
      Auditor     string // "立信会计师事务所"
      IsStandard  bool   // 是否标准无保留
      Emphasis    []string // 强调事项段要点
      Key Matters []string // 关键审计事项
  }
  ```
- `FinancialData` 新增 `AuditOpinions map[string]*AuditOpinion`

**步骤2: 审计意见获取（扩展 `downloader/auditor.go`）**
- 现有 `FetchAuditorHistory` 调用 `scripts/fetch_auditor_history.py` 查询 cninfo
- 扩展 Python 脚本：在查询审计机构变更的同时，解析最新一期审计报告中的意见类型
- cninfo 公告中，审计报告标题通常包含"审计报告"，内容中可提取意见类型关键词：
  - 标准无保留："审计了...财务报表"
  - 保留意见："保留意见"
  - 无法表示意见："无法表示意见"
  - 否定意见："否定意见"
  - 强调事项："强调事项"
  - 关键审计事项："关键审计事项"

**步骤3: step1Audit 改造**
```go
func step1Audit(data *FinancialData) StepResult {
    // 如果有自动解析的审计意见，直接展示
    // 如果没有，保持"请查询年报确认"作为 fallback
}
```
- 标准无保留：Pass = true
- 非标准意见：Pass = false，并在结论中说明具体问题

**步骤4: 风险警示集成**
- `BuildRiskAlertSummary` 中增加审计意见检查：
  - 保留意见/无法表示意见/否定意见 → high 级别 flag，一票否决
  - 强调事项 → medium 级别 flag

**步骤5: 报告展示**
- `report.go` 中新增/改造审计意见展示（step1 的数据目前没有独立模块展示，可在模块1执行摘要中增加审计意见段落）

```markdown
## 1.1 审计意见

| 年份 | 审计意见 | 事务所 | 是否标准 | 备注 |
|------|----------|--------|----------|------|
| 2023 | 标准无保留意见 | 立信会计师事务所 | ✅ | — |
| 2022 | 带强调事项段的无保留意见 | 立信会计师事务所 | ⚠️ | 强调：持续经营能力存在重大不确定性 |
```

### 2.3 涉及文件
- `downloader/auditor.go` — 扩展数据结构和获取逻辑
- `scripts/fetch_auditor_history.py` — 扩展审计意见解析
- `analyzer/types.go` — 新增 AuditOpinion 类型
- `analyzer/steps.go` — 改造 step1Audit
- `analyzer/risk_alert.go` — 增加审计意见风险检查
- `analyzer/report.go` — 增加审计意见展示段落

### 2.4 风险评估
| 风险点 | 等级 | 缓解措施 |
|--------|------|----------|
| cninfo 公告解析准确率 | 中 | 先用关键词匹配，准确率不高时标记为"需人工复核" |
| Python 脚本改动影响现有审计机构变更查询 | 低 | 扩展而非替换，原有 AuditorHistory 结构不变 |
| 审计意见类型中文表述多样 | 中 | 建立映射表：保留/无法表示/否定/强调/标准 |

### 2.5 工作量估计
- Python 脚本扩展 + Go 下载器改造：~1 天
- 分析器改造 + 风险集成：~0.5 天
- 报告展示：~0.5 天

---

## 三、F4: 报告差异追踪

### 3.1 问题现状
- `storage.go` 中 `SaveReport` 会清理旧报告，只保留 `latest.md`
- `SaveSnapshot` 也是覆盖写入，无历史
- 用户无法感知"本次和上次相比风险在哪里变了"

### 3.2 实现逻辑

**核心原则**：保留分析快照历史，自动生成 diff 摘要。

**步骤1: 快照历史化改造（`storage.go`）**
- `SaveSnapshot` 改为保留历史，而非覆盖：
  ```go
  func (s *Storage) SaveSnapshot(symbol string, report *analyzer.AnalysisReport) error {
      // 保存到 snapshots/{symbol}/{timestamp}.json
      // 同时保存一份 latest.json 作为快捷访问
      // 保留最近 10 次快照，旧自动清理
  }
  ```
- 新增 `ListSnapshotHistory(symbol string) ([]SnapshotInfo, error)`
- 新增 `LoadSnapshotByTime(symbol string, timestamp string) (*analyzer.AnalysisReport, error)`

**步骤2: 差异计算引擎（新增 `analyzer/diff.go`）**
- `ComputeAnalysisDiff(current, previous *AnalysisReport) *AnalysisDiff`
- 计算维度：
  ```go
  type AnalysisDiff struct {
      ScoreChange      float64          // 总分变化（当前 - 上次）
      GradeChanged     bool             // 等级是否变化（A→B）
      NewFlags         []RiskAlertFlag  // 新增风险
      ResolvedFlags    []RiskAlertFlag  // 解除的风险
      KeyMetricChanges []MetricChange   // 关键指标变化（ROE、毛利率、营收增速等）
      DataSourceChanges []string        // 数据源变化（如新增可比公司）
  }
  ```
- 关键指标变化阈值：变化 > 5% 或跨越阈值时记录

**步骤3: diff 报告生成（`analyzer/report.go` 新增 `writeModuleDiff`）**
```markdown
## 模块X: 与上次分析对比

> 上次分析时间：2024-03-15 | 本次分析时间：2024-05-14

### 评分变化
- 综合评分：**72 → 68**（↓4分，等级维持 B）

### 风险变化
| 类型 | 数量 | 详情 |
|------|------|------|
| 🆕 新增风险 | 1 | 大股东高比例质押（80%） |
| ✅ 解除风险 | 1 | 营收断崖式下跌（已恢复） |
| ⏸️ 持续风险 | 2 | A-Score 偏高、毛利率下滑 |

### 关键指标变化
| 指标 | 上次 | 本次 | 变化 |
|------|------|------|------|
| ROE | 15.2% | 13.8% | ↓1.4pct |
| 毛利率 | 31.2% | 28.5% | ↓2.7pct 🔴 |
| 营收增速 | -5% | -18% | ↓13pct 🔴 |
| 经营现金流/净利润 | 1.2 | 0.8 | ↓0.4 🟡 |
```

**步骤4: 前端 diff 面板**
- App.tsx 中分析完成后，如果存在历史快照，自动展示 diff 摘要
- 在"亮点与风险"面板下方增加"变化追踪"折叠区域
- 用红/绿色箭头直观展示变化方向

**步骤5: 数据清理策略**
- 每个股票最多保留 10 份快照历史
- 超过 10 份时删除最旧的一份
- 用户可在设置中选择"保留分析历史"开关（默认开启）

### 3.3 涉及文件
- `storage.go` — SaveSnapshot 历史化、新增 List/Load 历史接口
- `analyzer/diff.go` — 新增（差异计算引擎）
- `analyzer/types.go` — 新增 AnalysisDiff、SnapshotInfo 类型
- `analyzer/report.go` — 新增 diff 报告模块
- `app.go` — 分析完成后调用 diff 计算，传递 diff 数据到前端
- `frontend/src/App.tsx` — 新增 diff 面板 UI
- `frontend/wailsjs/go/models.ts` — 新增类型

### 3.4 风险评估
| 风险点 | 等级 | 缓解措施 |
|--------|------|----------|
| 快照文件增多占用磁盘 | 低 | 限制 10 份/股票，单份快照约 50KB，可忽略 |
| 第一次分析无历史，diff 为空 | 低 | 优雅处理：提示"首次分析，无历史对比" |
| diff 计算因数据结构变化出错 | 中 | 加 recover() 保护，diff 失败不影响主报告生成 |

### 3.5 工作量估计
- 存储层改造：~0.5 天
- diff 计算引擎：~1 天
- 报告渲染 + 前端面板：~1 天

---

## 四、F5: 可比公司自动推荐

### 4.1 问题现状
- 用户需要手动输入代码添加可比公司
- 没有推荐逻辑，新手用户不知道该选哪些

### 4.2 实现逻辑

**核心原则**：基于多维度相似度打分，Top-N 推荐并标注推荐理由。

**步骤1: 推荐算法（新增 `analyzer/recommend.go`）**
```go
func RecommendComparables(
    targetSymbol string,
    targetProfile *downloader.StockProfile,
    targetData *analyzer.FinancialData,
    allStocks []StockInfo, // 全市场股票库
) []ComparableRecommendation

type ComparableRecommendation struct {
    Symbol      string  // "000001.SZ"
    Name        string  // "平安银行"
    Score       float64 // 0-100 相似度得分
    Reasons     []string // 推荐理由：["同属银行业", "市值相近(±20%)", "ROE结构相似"]
    DataQuality string  // "high" / "medium" / "low"（本地是否有财报数据）
}
```

**相似度计算维度**（权重可配置）：

| 维度 | 权重 | 计算方式 |
|------|------|----------|
| 行业匹配 | 30% | 同申万/中信/GICS 一级行业得满分，二级得 70%，不同得 0 |
| 市值相近 | 20% | `1 - |ln(targetCap/peerCap)| / ln(5)`，超出 5 倍得 0 |
| 业务关键词 | 20% | 经营范围/概念标签 Jaccard 相似度 |
| ROE 结构相似 | 15% | 最近一年 ROE 差值 < 3pct 得满分，< 10pct 得一半 |
| 毛利率结构相似 | 15% | 最近一年毛利率差值 < 5pct 得满分，< 15pct 得一半 |

**步骤2: 数据源**
- 行业：`StockProfile.Industry`（来自东财 API）
- 市值：`StockProfile.MarketCap` 或 `QuoteData.MarketCap`
- 经营范围：`StockProfile.BusinessScope`（需扩展 Profile 结构，当前可能没有）
- 概念标签：`GetStockConcepts` 已有
- 财务指标：`FinancialData` 中最新年份 ROE、毛利率

**步骤3: 数据获取策略**
- 推荐时不需要下载所有股票的财报
- 优先使用本地已有财报数据的股票（`~/.config/stock-analyzer/data/` 下已有的 symbol）
- 如果本地不足 5 个，从全市场股票库中按行业+市值筛选，标记为"需下载财报"

**步骤4: UI 展示**
- 在可比公司面板增加"自动推荐"按钮
- 推荐结果展示为卡片列表：
  ```
  ┌─────────────────────────────────────────────┐
  │  000001.SZ  平安银行     相似度 87  [添加]   │
  │  推荐理由：同属银行业 · 市值相近 · ROE结构相似 │
  │  数据状态：✅ 本地有财报                        │
  └─────────────────────────────────────────────┘
  │  600036.SH  招商银行     相似度 82  [添加]   │
  │  推荐理由：同属银行业 · 毛利率结构相似           │
  │  数据状态：⚠️ 需下载财报                        │
  └─────────────────────────────────────────────┘
  ```

**步骤5: 交互逻辑**
- 点击"自动推荐"→调用 `RecommendComparables`→展示 Top 5 推荐
- 点击"添加"→加入可比公司列表，同时自动触发财报下载（如果本地无数据）
- 用户可勾选"自动添加 Top 3 推荐到可比公司"

### 4.3 涉及文件
- `analyzer/recommend.go` — 新增（推荐算法核心）
- `analyzer/types.go` — 新增 ComparableRecommendation 类型
- `analyzer/comparable.go` — 可能需扩展 ComparableAnalysis 以支持推荐理由展示
- `app.go` — 新增 `RecommendComparables` Wails 绑定方法
- `frontend/src/App.tsx` — 可比公司面板增加推荐 UI

### 4.4 风险评估
| 风险点 | 等级 | 缓解措施 |
|--------|------|----------|
| 行业分类标准不统一（申万/中信/GICS） | 中 | 以东财返回的 Industry 为主，同一字段即可匹配 |
| 经营范围字段可能缺失 | 中 | 缺失时降低该维度权重，用概念标签替代 |
| 推荐结果质量差 | 低 | 先标注"实验性"，用户可手动调整 |
| 全市场遍历性能 | 低 | 本地股票库约 5000+，纯内存计算 < 10ms |

### 4.5 工作量估计
- 推荐算法 + 后端接口：~1.5 天
- 前端推荐 UI：~1 天
- 数据获取策略优化：~0.5 天

---

## 五、O2: 前后端大文件拆分

### 5.1 问题现状
- `App.tsx` ~3842 行，80+ 个 state
- `app.go` ~3953 行
- `report.go` ~2834 行

### 5.2 实现逻辑（低风险渐进式拆分）

**核心原则**：只拆文件，不拆接口；保持 Wails 绑定方法签名不变；先拆纯函数/纯组件，最后拆大函数。

**阶段1: report.go 拆分（风险最低）**

当前 `report.go` 有 18 个 writeModuleX 函数 + 大量辅助函数。

拆分方案：
```
analyzer/report.go              → 保留 GenerateMarkdown + writeRiskAlertBanner + writeTOC + 总入口
analyzer/report_module1.go      → writeModule1 (执行摘要)
analyzer/report_module4.go      → writeModule4 / WriteModule4Only (可比公司)
analyzer/report_module8.go      → writeModule8 (ML预测)
analyzer/report_module13.go     → writeModule13 (投资建议)
analyzer/report_helpers.go      → 所有评分/评语辅助函数（约 800 行）
```

**优点**：每个模块独立文件，不影响业务逻辑；report.go 从 2834 行降至 ~400 行。

**阶段2: app.go 拆分（中等风险）**

`app.go` 的核心痛点是 `analyzeStockInternal` 约 700 行。但直接拆分会影响 Wails 绑定（因为绑定方法必须在 `App` 结构体上）。

拆分方案：将 `analyzeStockInternal` 拆为一个独立包或文件中的纯函数：
```go
// analyzer/orchestrator.go（新增）
func OrchestrateAnalysis(ctx *AnalysisContext) (*AnalysisReport, error)

// AnalysisContext 包含分析所需的所有输入
 type AnalysisContext struct {
     Symbol       string
     FinData      *FinancialData
     QuoteData    *QuoteData
     // ... 所有参数
 }
```

这样 `app.go` 中 `AnalyzeStock` 只需构建 `AnalysisContext` 并调用 `OrchestrateAnalysis`，自身从 700 行降至 ~50 行。

其他可拆分区域：
- `app.go` 中的导出功能（DownloadReport/ExportReportPDF/...）→ `app_export.go`
- `app.go` 中的配置功能（SFL/Update/Settings）→ `app_config.go`

**阶段3: App.tsx 拆分（风险较高）**

`App.tsx` 的问题不是文件大，而是 80+ 个 state 集中在单组件。

拆分方案（React Hooks 组合，不引入 Redux）：
```
frontend/src/hooks/useWatchlist.ts      → 自选股相关 state + 逻辑
frontend/src/hooks/useAnalysis.ts       → 分析流程相关 state + 逻辑
frontend/src/hooks/useComparables.ts    → 可比公司相关 state + 逻辑
frontend/src/hooks/useUIState.ts        → UI 状态（drawer/modal/loading 等）
frontend/src/components/ReportPanel.tsx → 报告渲染区域组件
frontend/src/components/ComparePanel.tsx → 可比公司面板组件
```

**建议**：O2 分阶段执行，先阶段1（report.go），再阶段2（app.go analyzeStockInternal），最后阶段3（App.tsx hooks）。每次拆分后完整回归测试。

### 5.3 涉及文件
- `analyzer/report.go` → 拆分为多个 report_moduleX.go + report_helpers.go
- `analyzer/orchestrator.go` — 新增（分析编排）
- `app.go` → 保留 Wails 绑定，业务逻辑移到 orchestrator.go / app_export.go / app_config.go
- `frontend/src/App.tsx` → 逐步拆出 hooks 和子组件

### 5.4 风险评估
| 风险点 | 等级 | 缓解措施 |
|--------|------|----------|
| 拆分导致 Wails 绑定失效 | 高 | 保持 `App` 结构体方法签名不变，只移动内部实现 |
| 拆分引入 import 循环 | 中 | 新包只依赖 analyzer/types，不依赖具体模块 |
| 前端 hooks 拆分后状态丢失 | 中 | 使用 useMemo/useCallback 保持引用稳定 |

### 5.5 工作量估计
- 阶段1 report.go 拆分：~1 天
- 阶段2 app.go 拆分：~1.5 天
- 阶段3 App.tsx 拆分：~2 天

---

## 六、执行优先级建议

考虑到"可信度、稳定性、可维护性"的优先级，建议按以下顺序执行：

| 优先级 | 功能 | 理由 |
|--------|------|------|
| 1 | **F4 报告差异追踪** | 用户闭环最直接，已有快照基础，改动范围可控 |
| 2 | **F2 审计意见自动解析** | 风险系统可信度提升，复用现有 cninfo 查询能力 |
| 3 | **F5 可比公司自动推荐** | 用户体验提升明显，算法独立，不影响核心分析 |
| 4 | **F1 季报/TTM 支持** | 投研时效性重要，但改动范围大（数据层改造） |
| 5 | **O2 文件拆分** | 可维护性提升，但属于"锦上添花"，可最后处理 |

---

## 七、需要用户确认的问题

1. **F1 季报**：是否接受"三层口径"方案（年报不变 + 季报/TTM 新增模块）？季度视图图表是否需要在 FinancialTrendDrawer 中展示？
2. **F2 审计意见**：是否接受先用关键词匹配 + "需人工复核" fallback 的方案？是否需要同时解析"关键审计事项"和"强调事项"？
3. **F4 差异追踪**：是否接受每个股票保留 10 份快照历史的限制？diff 面板放在"亮点与风险"下方是否合适？
4. **F5 可比推荐**：相似度维度权重（行业30%/市值20%/关键词20%/ROE15%/毛利率15%）是否合理？是否需要调整？
5. **O2 拆分**：是否接受分阶段执行？是否希望先只做 report.go 拆分？
6. **整体顺序**：是否同意上述优先级排序？或者你有其他偏好？
