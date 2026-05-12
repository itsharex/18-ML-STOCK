//go:build darwin

package tray

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
void setupTrayIcon(const char* iconPath, const char* tooltip);
void updateTrayTitle(const char* title, double changePercent);
*/
import "C"
import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed trayicon.png
var trayIconBytes []byte

var trayCtx context.Context

//export trayShowWindow
func trayShowWindow() {
	if trayCtx != nil {
		runtime.WindowShow(trayCtx)
	}
}

//export trayQuitApp
func trayQuitApp() {
	if trayCtx != nil {
		runtime.Quit(trayCtx)
	}
}

func initTray(ctx context.Context) {
	trayCtx = ctx

	// 将嵌入的图标写入临时文件（NSImage 需要文件路径）
	tmpDir := os.TempDir()
	iconPath := filepath.Join(tmpDir, "stockfinlens_trayicon.png")
	if err := os.WriteFile(iconPath, trayIconBytes, 0644); err != nil {
		fmt.Printf("[Tray] failed to write icon: %v\n", err)
		iconPath = ""
	}

	var cPath *C.char
	if iconPath != "" {
		cPath = C.CString(iconPath)
		defer C.free(unsafe.Pointer(cPath))
	}
	cTooltip := C.CString("StockFinLens 财报透镜")
	defer C.free(unsafe.Pointer(cTooltip))

	C.setupTrayIcon(cPath, cTooltip)
}

// updateTrayTitle 平台特定实现：更新 tray 显示的股票信息（A股涨红跌绿）
func updateTrayTitle(title string, changePercent float64) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.updateTrayTitle(cTitle, C.double(changePercent))
}

func quitTray() {
	trayCtx = nil
}
