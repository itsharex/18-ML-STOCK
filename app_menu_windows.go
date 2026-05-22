//go:build windows

package main

// setupApplicationMenu Windows 版本不显示应用菜单
func (a *App) setupApplicationMenu() {
	// Windows 下不设置菜单栏
}
