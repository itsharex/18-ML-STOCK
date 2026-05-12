//go:build !darwin

package tray

import "context"

func initTray(ctx context.Context) {}
func quitTray() {}
func updateTrayTitle(title string, changePercent float64) {}
