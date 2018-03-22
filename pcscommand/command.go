// Package pcscommand 命令包
package pcscommand

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/pcsconfig"
	"os"
)

var (
	info = new(baidupcs.BaiduPCS)
)

// GetPCSInfo 重载并返回 PCS 配置信息
func GetPCSInfo() *baidupcs.BaiduPCS {
	ReloadInfo()
	return info
}

// ReloadInfo 重载配置
func ReloadInfo() {
	pcsconfig.Reload()
	info = baidupcs.NewPCS(pcsconfig.Config.MustGetActive().BDUSS)
}

// ReloadIfInConsole 程序在 Console 模式下才会重载配置
func ReloadIfInConsole() {
	if len(os.Args) == 1 {
		ReloadInfo()
	}
}
