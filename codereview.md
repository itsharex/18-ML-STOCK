# StockFinLens 代码审查报告

> **审查来源**: GPT-5.5 全盘 Code Review + Kimi 人工复核
> **审查日期**: 2026-05-14
> **基线版本**: v1.3.37
> **基线测试结果**: `go test -short ./...` 通过, `npm test` 通过, `npm run build` 通过
>
> **总体结论**: 当前功能已经很多，下一阶段不建议继续堆模块，优先把"可信度、稳定性、可维护性、用户闭环"补齐。

---

## 一、最优先修复（Bug / 稳定性）

| # | 问题 | 文件位置 | 复核结论 | 优先级 | 状态 | 备注 |
|---|------|----------|----------|--------|------|------|
| B1 | **风险爬虫数据 Key 不匹配，导致非财务风险实际未生效** | `app.go:2248` 写入 `pledge_ratio`/`inquiry_count_1y`/`reduction_count_1y`，但 `analyzer/risk_alert.go:273` 读取的是驼峰 `pledgeRatio`/`inquiryCount`/`reductionCount`。同时 `analyzer/engine.go:34` `RunAnalysisWithAll` 接收 `extras` 参数后未回填到 `data.Extras`。 | ✅ **属实，严重 Bug** | P0 | 🔴 待修复 | 需三处联动修复：统一 key 命名、回填 extras、加回归测试 |
| B2 | **自选股活跃度空指针风险** | `app.go:786`、`app.go:902` 在 `quote == nil` 分支的 `fmt.Printf` 中仍访问 `quote.CirculatingMarketCap`（条件判断因短路安全，但日志打印会 panic）。 | ⚠️ **部分属实** | P0 | 🔴 待修复 | 将 `quote.CirculatingMarketCap` 改为安全访问，或调整日志格式 |
| B3 | **分析入口边界 panic 风险** | `app.go:2525` 直接传 `finData.Extras`，但 `finData` 可能为 nil。 | ❌ **不属实** | - | 🟢 无需修复 | `finData` 在行 1481 已做 nil 检查并提前返回，不会执行到 2525 行 |
| B4 | **并发结构嵌套 goroutine 导致竞态** | `app.go:1874` `wgNet.Add(4)`，但资金流 goroutine（`app.go:1989`）嵌套在舆情 goroutine 内部。舆情 goroutine 可能先 Done，导致 `wgNet.Wait()` 提前返回而资金流尚未完成。 | ✅ **属实，严重 Bug** | P0 | 🔴 待修复 | 将四个 goroutine 平铺启动，禁止嵌套 |

---

## 二、功能缺口

| # | 问题 | 文件位置 | 复核结论 | 优先级 | 状态 | 备注 |
|---|------|----------|----------|--------|------|------|
| F1 | **季报 / TTM 缺失，投研时效性不足** | `analyzer/data.go:141` 明确过滤季报，只保留年报（`-12-31` 结尾）。 | ✅ 属实 | P1 | 🔴 待规划 | 建议新增"年报长期质量 + 季报滚动预警 + TTM 估值"三层口径 |
| F2 | **审计意见仍是手动占位** | `analyzer/steps.go:16` 仍提示用户"请查询年报确认"。 | ✅ 属实 | P1 | 🔴 待规划 | 非标意见、强调事项、关键审计事项、审计机构变更应自动解析或结构化导入 |
| F3 | **没有持仓 / 组合层** | 当前为单股分析工具。 | ✅ 属实 | P2 | 🔴 待规划 | 大功能，专业用户需要：持仓成本、仓位、组合行业暴露、组合风险热力图、估值分位、再平衡建议、止损/止盈提醒 |
| F4 | **缺少"报告差异追踪"** | 每次分析覆盖 `latest.md`，用户无法感知"本次和上次相比风险在哪里变了"。 | ✅ 属实 | P1 | 🔴 待规划 | 建议生成 diff：分数变化、风险新增/解除、关键指标变化、数据源变化 |
| F5 | **可比公司仍偏手动** | 当前可比公司需要用户手动添加。 | ✅ 属实 | P1 | 🔴 待规划 | 建议自动推荐：同申万/中信/GICS 行业、相近市值、相近业务关键词、相近毛利率/ROE 结构，并标注可比理由 |

---

## 三、需要优化

| # | 问题 | 文件位置 | 复核结论 | 优先级 | 状态 | 备注 |
|---|------|----------|----------|--------|------|------|
| O1 | **ML 可信度需降噪** | `analyzer/ml_features.go:18` 注释写"8 个季度"但实际用年报序列；`analyzer/ml_features.go:318` Engine-D 使用硬编码默认市场特征（`peTTM=25.0`、`pb=2.5` 等）。 | ✅ 属实 | P1 | 🔴 待修复 | UI 和报告中应明确"模型输入完整度/置信等级"，不把默认填充的预测展示得像真实预测 |
| O2 | **前端和后端主文件过大** | `frontend/src/App.tsx` (~3420 行) 状态非常集中；`app.go` (~3460 行)、`report.go` 也偏大。 | ✅ 属实 | P1 | 🔴 待规划 | 建议按低风险边界拆分：数据源服务、分析编排服务、导出服务、设置服务、报告模块组件 |
| O3 | **报告 Markdown 渲染存在 XSS 风险** | `frontend/src/App.tsx:3599` 使用了 `rehypeRaw`。 | ✅ 属实 | P1 | 🔴 待修复 | 外部新闻标题或导入内容混入 HTML 时，Wails 环境有 XSS 风险。建议禁用 raw HTML 或加 sanitize 白名单 |
| O4 | **SFL token 本地明文保存** | `storage.go:986` 保存 token，`storage.go:1037` 用 `0644` 写文件。 | ✅ 属实 | P1 | 🔴 待修复 | 建议改为系统钥匙串/凭据管理器，至少文件权限改为 `0600` |
| O5 | **构建和文档版本漂移** | `go.mod:3` 要求 Go `1.25.0`；`README.md:140` 仍写 `>=1.22`；`功能列表.md:85` 写 `Go 1.26`；`AGENTS.md:13` 版本号仍写 `1.3.29`（实际已 `1.3.37`）。 | ✅ 属实 | P1 | 🔴 待修复 | 发布前统一版本号、Go 要求、功能数量描述 |

---

## 四、测试建议

| # | 问题 | 文件位置 | 复核结论 | 优先级 | 状态 | 备注 |
|---|------|----------|----------|--------|------|------|
| T1 | **前端测试覆盖不足** | 当前前端仅 `frontend/src/Settings.test.ts:1` 有测试。 | ✅ 属实 | P1 | 🔴 待规划 | 建议补三类测试：风险数据链路回归、数据源失败/空返回回归、前端关键流程组件测试 |
| T2 | **回归脚本扫描范围问题** | `go test ./...` 会扫到 `frontend/node_modules`。 | ✅ 属实 | P2 | 🔴 待修复 | 建议回归脚本改成显式 Go 包列表，避免第三方目录影响测试稳定性 |

---

## 五、修复任务跟踪清单

### P0 - 严重 Bug（立即修复）

- [x] **B1** 修复风险爬虫数据 Key 不匹配问题 ✅ 2026-05-14
  - [x] 统一 `app.go` 与 `analyzer/risk_alert.go` 的 extras key 命名（下划线 vs 驼峰）：`risk_alert.go` 三处驼峰 key 改为下划线（`pledgeRatio`→`pledge_ratio`、`inquiryCount`→`inquiry_count_1y`、`reductionCount`→`reduction_count_1y`）
  - [x] 在 `analyzer/engine.go:34` `RunAnalysisWithAll` 中将传入的 `extras` 回填到 `data.Extras`
  - [x] `app.go:2248` 的 `make(map[string]float64)` 改为先检查 `finData.Extras == nil`，避免覆盖已有数据
  - [x] 回归验证：`go test -short ./...` 全部通过
- [x] **B2** 修复自选股活跃度空指针风险 ✅ 2026-05-14
  - [x] `app.go:786`、`app.go:902` 的 `fmt.Printf` 中安全访问 `quote.CirculatingMarketCap`（nil 时打印 0）
- [x] **B4** 修复并发结构嵌套 goroutine ✅ 2026-05-14
  - [x] 将 `app.go:1989` 的资金流 goroutine 从舆情 goroutine 内部移出，四个 goroutine 平铺启动

### P1 - 重要优化

- [x] **F1** 季报/TTM 支持 ✅ 2026-05-14
  - `analyzer/data.go`：不再过滤季报，保留所有期间；新增 `FinancialData.Quarters` 字段
  - 新增 `analyzer/quarterly.go`：季度滚动预警（营收/净利润/毛利率/经营现金流环比检测）
  - 新增 `analyzer/ttm.go`：TTM 滚动指标计算（TTM 营收/净利润/ROE/净利率/现金流比率）
  - `analyzer/report.go`：新增季度预警和 TTM 指标报告模块
- [x] **F2** 审计意见自动解析 ✅ 2026-05-14
  - 扩展 `scripts/fetch_auditor_history.py`：解析审计报告公告标题推断意见类型（标准/保留/无法表示/否定/强调事项）
  - `downloader/auditor.go`：新增 `AuditOpinion` 结构体
  - `analyzer/steps.go`：`step1Audit` 从占位符变为真实数据展示
  - `analyzer/risk_alert.go`：非标审计意见触发 high/medium 风险 flag，一票否决
  - `analyzer/report.go`：新增审计意见展示段落
- [x] **F4** 报告差异追踪 ✅ 2026-05-14
  - `storage.go`：`SaveSnapshot` 历史化（保留最近 10 份/股票）
  - 新增 `analyzer/diff.go`：差异计算引擎（评分变化、风险新增/解除/持续、关键指标变化）
  - `analyzer/report.go`：新增 diff 报告模块
  - 前端 App.tsx：新增"📈 与上次分析对比"折叠面板
- [x] **F5** 可比公司自动推荐 ✅ 2026-05-14
  - 新增 `analyzer/recommend.go`：5 维度相似度算法（行业30%/市值20%/关键词20%/ROE15%/毛利率15%）
  - `app.go`：新增 `RecommendComparables` Wails 绑定
  - 前端 App.tsx：可比公司面板增加"🔍 自动推荐"按钮，推荐卡片展示相似度+推荐理由+数据状态
- [x] **O1** ML 可信度降噪 ✅ 2026-05-14
- [x] **O2** 前后端大文件拆分 ✅ 2026-05-14
  - 阶段1：`analyzer/report.go`（~2900行）→ `report.go` + `report_modules.go` + `report_helpers.go`
  - 阶段2：`app.go`（~4000行）→ `app.go`（~3200行）+ `app_analysis.go`（~800行）
- [x] **O3** 禁用/替换 rehypeRaw，消除 XSS 风险 ✅ 2026-05-14
- [x] **O4** SFL token 安全存储（文件权限 0600） ✅ 2026-05-14
- [x] **O5** 统一文档版本号 ✅ 2026-05-14
- [x] **T1** 补充前端测试 ✅ 2026-05-14
  - 新增 `RiskBadge.test.tsx`（5 个用例）
  - 新增 `RiskAlertBanner.test.tsx`（5 个用例）
  - 前端测试从 4 个增至 14 个

### P2 - 中长期规划

- [ ] **F3** 持仓/组合层（大功能，需独立设计）
- [ ] **T2** 回归脚本改为显式 Go 包列表

---

## 六、复核补充说明

1. **B3 已排除**: `app.go:2525` 的 `finData.Extras` 访问不会 panic，因为行 1481 已做 nil 检查并提前返回。原 review 建议此处有误，无需修复。
2. **B1 影响范围大**: 该 Bug 导致风险爬虫（股权质押、问询函、减持）的数据自引入以来就未实际参与 A-Score 计算，直接影响风险判断可信度，必须优先修复。
3. **B4 竞态风险**: 当前代码在测试环境中"能跑"是因为网络请求耗时足够让嵌套 goroutine 在 `wgNet.Wait()` 返回前完成，但在网络极快或本地缓存命中时会暴露竞态。
4. **O5 建议自动化**: 建议在构建脚本中增加版本号一致性校验的覆盖范围，将 README、功能列表.md、AGENTS.md 也纳入校验。
