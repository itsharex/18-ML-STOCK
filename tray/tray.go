package tray

import "context"

// Init 初始化系统托盘（平台特定实现）
func Init(ctx context.Context) {
	initTray(ctx)
}

// Quit 清理系统托盘
func Quit() {
	quitTray()
}

// UpdateTitle 更新 tray 标题（跨平台公共接口）
// title: 显示的文本，changePercent: 涨跌幅（用于决定颜色，A股涨红跌绿）
func UpdateTitle(title string, changePercent float64) {
	updateTrayTitle(title, changePercent)
}
