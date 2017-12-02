package baidupcscmd

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/config"
	"os"
)

var (
	info   = new(baidupcs.PCSApi)
	thread int
)

func init() {
	ReloadInfo()
}

// ReloadInfo 重载配置
func ReloadInfo() {
	pcsconfig.Reload()
	info = baidupcs.NewPCS(pcsconfig.Config.GetActiveBDUSS())
	info.Workdir = pcsconfig.Config.Workdir
	thread = pcsconfig.Config.MaxParallel
}

func ReloadIfInConsole() {
	if len(os.Args) == 1 {
		ReloadInfo()
	}
}
