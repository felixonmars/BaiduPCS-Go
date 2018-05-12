package pcscommand

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/baidu-tools/pan"
	"path"
)

// RunShareSet 执行分享
func RunShareSet(paths []string, option *baidupcs.ShareOption) {
	pcspaths, err := getAllAbsPaths(paths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	shared, err := GetBaiduPCS().ShareSet(pcspaths, option)
	if err != nil {
		fmt.Printf("%s失败: %s\n", baidupcs.OperationShareSet, err)
		return
	}

	fmt.Printf("shareID: %d, 链接: %s\n", shared.ShareID, shared.Link)
}

// RunShareCancel 执行取消分享
func RunShareCancel(shareIDs []int64) {
	if len(shareIDs) == 0 {
		fmt.Printf("%s失败, 没有任何 shareid\n", baidupcs.OperationShareCancel)
		return
	}

	err := GetBaiduPCS().ShareCancel(shareIDs)
	if err != nil {
		fmt.Printf("%s失败: %s\n", baidupcs.OperationShareCancel, err)
		return
	}

	fmt.Printf("%s成功\n", baidupcs.OperationShareCancel)
}

func getShareDLink(pcspath string) (dlink string) {
	pcs := GetBaiduPCS()
	s, pcsError := pcs.ShareSet([]string{pcspath}, nil)
	if pcsError != nil {
		pcsCommandVerbose.Warn(pcsError.Error())
		return ""
	}

	// 取消分享
	defer pcs.ShareCancel([]int64{s.ShareID})

	sInfo := pan.NewSharedInfo(s.Link)
	sInfo.SetHTTPS(pcsconfig.Config.EnableHTTPS())
	sInfo.Client = requester.NewHTTPClient()
	sInfo.Client.SetHTTPSecure(pcsconfig.Config.EnableHTTPS())

	uk, pcsError := pcs.UK()
	if pcsError != nil {
		pcsCommandVerbose.Warn(pcsError.Error())
		err := sInfo.InitInfo()
		if err != nil {
			pcsCommandVerbose.Warn(err.Error())
			return ""
		}
	} else {
		sInfo.UK = uk
		sInfo.ShareID = s.ShareID
		sInfo.RootSharePath = path.Dir(pcspath)
	}

	fd, err := sInfo.Meta(path.Base(pcspath))
	if err != nil {
		pcsCommandVerbose.Warn(err.Error())
		return ""
	}

	return fd.Dlink
}
