//go:build darwin

package updater

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// ApplyUpdate 在 macOS 上应用更新：打开 dmg 后退出当前进程
// 否则用户拖动 .app 到 Applications 时会被提示"目标程序正在运行"。
// 延迟 1.5s 退出，给 dmg 窗口足够时间弹出，同时让 Wails 调用方收到成功返回、前端能展示提示。
func ApplyUpdate(dmgPath string) error {
	cmd := exec.Command("open", dmgPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("打开 DMG 失败: %w", err)
	}
	go func() {
		time.Sleep(1500 * time.Millisecond)
		os.Exit(0)
	}()
	return nil
}
