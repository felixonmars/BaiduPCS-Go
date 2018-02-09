package pcscommand

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/pcsconfig"
	"os"
)

var (
	info = new(baidupcs.PCSApi)
)

func GetPCSInfo() *baidupcs.PCSApi {
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
