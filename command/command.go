package baidupcscmd

import (
	"fmt"
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
	baidu, err := pcsconfig.Config.GetBaiduUserByUID(pcsconfig.Config.BaiduActiveUID)
	if err != nil {
		fmt.Println(err)
		return
	}
	info = baidupcs.NewPCS(baidu.BDUSS)
	info.Workdir = pcsconfig.Config.Workdir
	thread = pcsconfig.Config.MaxParallel
}

// ReloadIfInConsole 程序在 Console 模式下才回重载配置
func ReloadIfInConsole() {
	if len(os.Args) == 1 {
		ReloadInfo()
	}
}
