package analyzer

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// ComparableRecommendation 可比公司推荐项
type ComparableRecommendation struct {
	Symbol      string   `json:"symbol"`
	Name        string   `json:"name"`
	Score       float64  `json:"score"`       // 0-100 相似度得分
	Reasons     []string `json:"reasons"`     // 推荐理由
	DataQuality string   `json:"dataQuality"` // high/medium/low（本地是否有财报数据）
}

// RecommendComparables 基于多维度相似度自动推荐可比公司
// targetSymbol: 目标股票代码（带点格式，如 "000001.SZ"）
// targetProfile: 目标股票资料（行业、市值等）
// targetData: 目标股票财务数据（用于提取 ROE、毛利率）
// dataDir: 本地数据根目录（用于扫描候选股票）
// maxResults: 最大返回数量
func RecommendComparables(targetSymbol string, targetProfile *StockProfile, targetData *FinancialData, dataDir string, maxResults int) []ComparableRecommendation {
	if maxResults <= 0 {
		maxResults = 5
	}

	// 提取目标股票的关键特征
	targetIndustry := ""
	targetMarketCap := 0.0
	if targetProfile != nil {
		targetIndustry = targetProfile.Industry
		targetMarketCap = targetProfile.MarketCap
	}

	targetROE := 0.0
	targetGM := 0.0
	if targetData != nil && len(targetData.Years) > 0 {
		year := targetData.Years[0]
		equity := targetData.GetValueOrZero(targetData.BalanceSheet, "所有者权益合计", year)
		if equity == 0 {
			totalAssets := targetData.GetValueOrZero(targetData.BalanceSheet, "总资产", year)
			totalLiabilities := targetData.GetValueOrZero(targetData.BalanceSheet, "总负债", year)
			equity = totalAssets - totalLiabilities
		}
		netProfit := targetData.GetValueOrZero(targetData.IncomeStatement, "净利润", year)
		revenue := targetData.GetValueOrZero(targetData.IncomeStatement, "营业收入", year)
		cost := targetData.GetValueOrZero(targetData.IncomeStatement, "营业成本", year)
		if equity > 0 {
			targetROE = netProfit / equity
		}
		if revenue > 0 {
			targetGM = (revenue - cost) / revenue
		}
	}

	// 扫描本地数据目录获取候选股票
	candidates := scanLocalCandidates(dataDir, targetSymbol)

	// 计算每个候选股票的相似度
	var scored []ComparableRecommendation
	for _, c := range candidates {
		score, reasons, dataQuality := computeSimilarity(
			targetIndustry, targetMarketCap, targetROE, targetGM,
			c,
		)
		if score > 0 {
			scored = append(scored, ComparableRecommendation{
				Symbol:      c.Symbol,
				Name:        c.Name,
				Score:       score,
				Reasons:     reasons,
				DataQuality: dataQuality,
			})
		}
	}

	// 按得分降序排序
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].Score > scored[i].Score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	if len(scored) > maxResults {
		scored = scored[:maxResults]
	}
	return scored
}

// candidateInfo 候选股票内部信息
type candidateInfo struct {
	Symbol    string
	Name      string
	Industry  string
	MarketCap float64
	ROE       float64
	GM        float64
	HasData   bool // 是否有本地财报数据
}

// StockProfile 推荐算法使用的股票资料子集（避免循环导入）
type StockProfile struct {
	Industry  string
	MarketCap float64
}

// scanLocalCandidates 扫描本地数据目录获取候选股票
func scanLocalCandidates(dataDir, excludeSymbol string) []candidateInfo {
	dataRoot := filepath.Join(dataDir, "data")
	entries, err := os.ReadDir(dataRoot)
	if err != nil {
		return nil
	}

	var candidates []candidateInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		symbol := entry.Name()
		if symbol == excludeSymbol {
			continue
		}

		info := candidateInfo{Symbol: symbol}

		// 尝试读取 profile.json 获取行业和市值
		profilePath := filepath.Join(dataRoot, symbol, "profile.json")
		if data, err := os.ReadFile(profilePath); err == nil {
			// 简单 JSON 解析（避免引入完整 json 库依赖）
			info.Industry = extractJSONString(data, "industry")
			info.Name = extractJSONString(data, "name")
			if mc := extractJSONFloat(data, "market_cap"); mc > 0 {
				info.MarketCap = mc
			}
		}

		// 尝试读取财务数据获取 ROE 和毛利率
		bsPath := filepath.Join(dataRoot, symbol, "balance_sheet.json")
		isPath := filepath.Join(dataRoot, symbol, "income_statement.json")
		if _, err1 := os.Stat(bsPath); err1 == nil {
			if _, err2 := os.Stat(isPath); err2 == nil {
				info.HasData = true
				// 简化：从 income_statement.json 和 balance_sheet.json 中提取最新年份数据
				// 实际实现中这里可以调用 LoadFinancialData，但为了性能避免完整加载
				info.ROE, info.GM = extractLatestMetrics(dataRoot, symbol)
			}
		}

		// 至少要有行业或财务数据才作为候选
		if info.Industry != "" || info.HasData {
			candidates = append(candidates, info)
		}
	}
	return candidates
}

// computeSimilarity 计算候选股票与目标股票的相似度
func computeSimilarity(targetIndustry string, targetMarketCap, targetROE, targetGM float64, c candidateInfo) (float64, []string, string) {
	score := 0.0
	var reasons []string

	// 1. 行业匹配 (30%)
	if targetIndustry != "" && c.Industry != "" {
		if targetIndustry == c.Industry {
			score += 30
			reasons = append(reasons, "同属"+targetIndustry)
		} else {
			// 简单的行业关键词匹配（取行业名称的前2-4个字）
			targetKeys := extractIndustryKeywords(targetIndustry)
			candidateKeys := extractIndustryKeywords(c.Industry)
			matchCount := 0
			for _, tk := range targetKeys {
				for _, ck := range candidateKeys {
					if tk == ck && len(tk) >= 2 {
						matchCount++
					}
				}
			}
			if matchCount > 0 {
				partialScore := 30.0 * float64(matchCount) / float64(len(targetKeys))
				if partialScore > 30 {
					partialScore = 30
				}
				score += partialScore
				reasons = append(reasons, "行业相近")
			}
		}
	}

	// 2. 市值相近 (20%)
	if targetMarketCap > 0 && c.MarketCap > 0 {
		ratio := targetMarketCap / c.MarketCap
		if ratio < 1 {
			ratio = c.MarketCap / targetMarketCap
		}
		if ratio <= 2 {
			score += 20
			reasons = append(reasons, "市值相近")
		} else if ratio <= 5 {
			score += 20 * (1 - (ratio-2)/3)
			if score < 0 {
				score = 0
			}
			reasons = append(reasons, "市值相近")
		}
	}

	// 3. ROE 结构相似 (15%)
	if targetROE != 0 && c.ROE != 0 {
		diff := math.Abs(targetROE - c.ROE)
		if diff < 0.03 {
			score += 15
			reasons = append(reasons, "ROE结构相似")
		} else if diff < 0.10 {
			score += 15 * (1 - (diff-0.03)/0.07)
			reasons = append(reasons, "ROE结构相似")
		}
	}

	// 4. 毛利率结构相似 (15%)
	if targetGM != 0 && c.GM != 0 {
		diff := math.Abs(targetGM - c.GM)
		if diff < 0.05 {
			score += 15
			reasons = append(reasons, "毛利率结构相似")
		} else if diff < 0.15 {
			score += 15 * (1 - (diff-0.05)/0.10)
			reasons = append(reasons, "毛利率结构相似")
		}
	}

	// 5. 有财务数据加分 (20% -> 简化为"数据质量")
	dataQuality := "low"
	if c.HasData {
		score += 20
		dataQuality = "high"
		reasons = append(reasons, "本地有财报数据")
	} else if c.Industry != "" {
		score += 10
		dataQuality = "medium"
	}

	return score, reasons, dataQuality
}

// extractIndustryKeywords 从行业名称中提取关键词
func extractIndustryKeywords(industry string) []string {
	// 简单的分词：按常见分隔符分割
	parts := strings.FieldsFunc(industry, func(r rune) bool {
		return r == '、' || r == '/' || r == '·' || r == ' ' || r == '，' || r == ','
	})
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) >= 2 {
			result = append(result, p)
		}
	}
	return result
}

// extractJSONString 从 JSON 字节中简单提取字符串值（无引号转义处理，仅适用于简单 JSON）
func extractJSONString(data []byte, key string) string {
	// 简化：这里用字符串查找代替正则，避免引入 regexp 包
	keyPattern := `"` + key + `"`
	idx := strings.Index(string(data), keyPattern)
	if idx < 0 {
		return ""
	}
	rest := string(data)[idx+len(keyPattern):]
	// 跳过冒号和空白
	i := 0
	for i < len(rest) && (rest[i] == ':' || rest[i] == ' ' || rest[i] == '\t' || rest[i] == '\n' || rest[i] == '\r') {
		i++
	}
	if i < len(rest) && rest[i] == '"' {
		i++
		start := i
		for i < len(rest) && rest[i] != '"' {
			i++
		}
		return rest[start:i]
	}
	return ""
}

// extractJSONFloat 从 JSON 字节中简单提取 float64 值
func extractJSONFloat(data []byte, key string) float64 {
	keyPattern := `"` + key + `"`
	idx := strings.Index(string(data), keyPattern)
	if idx < 0 {
		return 0
	}
	rest := string(data)[idx+len(keyPattern):]
	i := 0
	for i < len(rest) && (rest[i] == ':' || rest[i] == ' ' || rest[i] == '\t' || rest[i] == '\n' || rest[i] == '\r') {
		i++
	}
	start := i
	for i < len(rest) && (rest[i] == '-' || rest[i] == '.' || (rest[i] >= '0' && rest[i] <= '9')) {
		i++
	}
	var val float64
	if _, err := fmt.Sscanf(rest[start:i], "%f", &val); err == nil {
		return val
	}
	return 0
}

// extractLatestMetrics 从本地数据中提取最新年份的 ROE 和毛利率
func extractLatestMetrics(dataRoot, symbol string) (roe, gm float64) {
	// 简化实现：调用 LoadFinancialData 获取完整数据
	// 为了性能，这里只读取不解析完整 JSON
	fd, err := LoadFinancialData(dataRoot, symbol)
	if err != nil || len(fd.Years) == 0 {
		return 0, 0
	}
	year := fd.Years[0]
	equity := fd.GetValueOrZero(fd.BalanceSheet, "所有者权益合计", year)
	if equity == 0 {
		totalAssets := fd.GetValueOrZero(fd.BalanceSheet, "总资产", year)
		totalLiabilities := fd.GetValueOrZero(fd.BalanceSheet, "总负债", year)
		equity = totalAssets - totalLiabilities
	}
	netProfit := fd.GetValueOrZero(fd.IncomeStatement, "净利润", year)
	revenue := fd.GetValueOrZero(fd.IncomeStatement, "营业收入", year)
	cost := fd.GetValueOrZero(fd.IncomeStatement, "营业成本", year)
	if equity > 0 {
		roe = netProfit / equity
	}
	if revenue > 0 {
		gm = (revenue - cost) / revenue
	}
	return
}
