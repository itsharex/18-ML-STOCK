#!/bin/bash
# StockFinLens L1/L2 回归测试统一入口
# 用法: ./scripts/run-regression.sh [quick|full]
#   quick: go test -short ./... + npm test（无网络，适合 CI）
#   full:  go test ./... + npm test（含端到端测试，适合发布前验证）
#
# 结果自动保存到 test-results/ 目录（带时间戳）

set -e

MODE="${1:-quick}"
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RESULT_DIR="$PROJECT_ROOT/test-results"
RESULT_FILE="$RESULT_DIR/regression_${MODE}_${TIMESTAMP}.log"
SUMMARY_FILE="$RESULT_DIR/latest_summary.md"

mkdir -p "$RESULT_DIR"

echo "========================================"
echo "StockFinLens 回归测试"
echo "模式: $MODE"
echo "时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================"

# 清空结果文件
echo "# StockFinLens 回归测试报告" > "$RESULT_FILE"
echo "" >> "$RESULT_FILE"
echo "- **模式**: $MODE" >> "$RESULT_FILE"
echo "- **时间**: $(date '+%Y-%m-%d %H:%M:%S')" >> "$RESULT_FILE"
echo "- **Commit**: $(git rev-parse --short HEAD 2>/dev/null || echo 'N/A')" >> "$RESULT_FILE"
echo "" >> "$RESULT_FILE"

GO_STATUS="⏭️ 跳过"
FE_STATUS="⏭️ 跳过"
OVERALL="⚠️ 未完成"

# 运行 Go 测试
if [ "$MODE" = "quick" ]; then
    echo "[1/2] 运行 Go 快速回归测试（-short，跳过网络）..."
    echo "## Go 后端测试（quick）" >> "$RESULT_FILE"
    echo '```' >> "$RESULT_FILE"
    if go test -short ./... 2>&1 | tee -a "$RESULT_FILE"; then
        GO_STATUS="✅ 通过"
    else
        GO_STATUS="❌ 失败"
        OVERALL="❌ 失败"
    fi
    echo '```' >> "$RESULT_FILE"
    echo "" >> "$RESULT_FILE"

    echo "[2/2] 运行前端组件测试..."
    echo "## 前端测试" >> "$RESULT_FILE"
    echo '```' >> "$RESULT_FILE"
    if (cd frontend && npm test 2>&1 | tee -a "$RESULT_FILE"); then
        FE_STATUS="✅ 通过"
    else
        FE_STATUS="❌ 失败"
        OVERALL="❌ 失败"
    fi
    echo '```' >> "$RESULT_FILE"

elif [ "$MODE" = "full" ]; then
    echo "[1/2] 运行 Go 完整回归测试（含网络请求和端到端）..."
    echo "## Go 后端测试（full）" >> "$RESULT_FILE"
    echo '```' >> "$RESULT_FILE"
    if go test ./... 2>&1 | tee -a "$RESULT_FILE"; then
        GO_STATUS="✅ 通过"
    else
        GO_STATUS="❌ 失败"
        OVERALL="❌ 失败"
    fi
    echo '```' >> "$RESULT_FILE"
    echo "" >> "$RESULT_FILE"

    echo "[2/2] 运行前端组件测试..."
    echo "## 前端测试" >> "$RESULT_FILE"
    echo '```' >> "$RESULT_FILE"
    if (cd frontend && npm test 2>&1 | tee -a "$RESULT_FILE"); then
        FE_STATUS="✅ 通过"
    else
        FE_STATUS="❌ 失败"
        OVERALL="❌ 失败"
    fi
    echo '```' >> "$RESULT_FILE"

else
    echo "未知模式: $MODE"
    echo "用法: $0 [quick|full]"
    exit 1
fi

# 判定总体结果
if [ "$GO_STATUS" = "✅ 通过" ] && [ "$FE_STATUS" = "✅ 通过" ]; then
    OVERALL="✅ 全部通过"
fi

echo "" >> "$RESULT_FILE"
echo "## 汇总" >> "$RESULT_FILE"
echo "" >> "$RESULT_FILE"
echo "| 项目 | 结果 |" >> "$RESULT_FILE"
echo "|------|------|" >> "$RESULT_FILE"
echo "| Go 后端 | $GO_STATUS |" >> "$RESULT_FILE"
echo "| 前端 | $FE_STATUS |" >> "$RESULT_FILE"
echo "| **总体** | **$OVERALL** |" >> "$RESULT_FILE"

# 更新最新摘要
ln -sf "$RESULT_FILE" "$SUMMARY_FILE"

echo ""
echo "========================================"
echo "$OVERALL"
echo "========================================"
echo "报告已保存: $RESULT_FILE"
echo "最新摘要: $SUMMARY_FILE"

if [ "$OVERALL" != "✅ 全部通过" ]; then
    echo ""
    echo "⚠️  测试未全部通过，请修复后再发布！"
    exit 1
fi
