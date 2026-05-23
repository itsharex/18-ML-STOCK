package downloader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/liusaipu/stockfinlens/analyzer"
)

func TestDownloadAndAnalyze603501(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过实时网络测试（需打东财 HTTP；发布前手动 go test ./... 跑一次作为外部接口 canary）")
	}
	data, err := DownloadFinancialReports(context.Background(), "SH", "603501")
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	tmpDir := t.TempDir()
	dir := filepath.Join(tmpDir, "data", "603501.SH")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := SaveAsJSON(data, dir); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	report, err := analyzer.RunAnalysis(tmpDir, "603501.SH", analyzer.AnalysisOptions{})
	if err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	fmt.Printf("Report Grade: %s\n", report.OverallGrade)
	fmt.Printf("Latest Year Score: %.1f\n", report.Score[report.Years[0]])
	fmt.Printf("Years: %v\n", report.Years)

	// 基本断言
	if len(report.Years) == 0 {
		t.Fatal("no years in report")
	}
	if report.Score[report.Years[0]] <= 0 {
		t.Fatal("score should be > 0")
	}
	if report.OverallGrade == "" {
		t.Fatal("grade should not be empty")
	}
}
