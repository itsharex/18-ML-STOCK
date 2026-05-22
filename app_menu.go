//go:build !windows

package main

import (
	"github.com/liusaipu/stockfinlens/tray"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// setupApplicationMenu 设置 macOS 顶部应用菜单
func (a *App) setupApplicationMenu() {
	appMenu := menu.NewMenu()

	// macOS 应用标准菜单（About, Hide, Quit 等）
	appMenu.Append(menu.AppMenu())

	// 显示菜单
	displayMenu := appMenu.AddSubmenu("显示")

	a.appMenuScrollItem = menu.Text("显示/关闭 滚动字幕", nil, func(cd *menu.CallbackData) {
		enabled := !tray.IsScrollEnabled()
		tray.SetScrollEnabled(enabled)
	})
	displayMenu.Append(a.appMenuScrollItem)

	a.appMenuIconItem = menu.Text("显示/隐藏 菜单图标", nil, func(cd *menu.CallbackData) {
		visible := !tray.IsIconVisible()
		tray.SetIconVisible(visible)
	})
	displayMenu.Append(a.appMenuIconItem)

	// 如果 tray 图标当前隐藏，禁用滚动字幕选项（无图标则无滚动意义）
	if !tray.IsIconVisible() {
		a.appMenuScrollItem.Disable()
	}

	// 关于菜单
	aboutMenu := appMenu.AddSubmenu("关于")
	aboutMenu.Append(menu.Text("关于 StockFinLens", nil, func(cd *menu.CallbackData) {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.InfoDialog,
			Title:   "关于 StockFinLens",
			Message: "版本 " + a.currentVersion,
		})
	}))

	runtime.MenuSetApplicationMenu(a.ctx, appMenu)
}
