package pcscommand

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/baidu-tools/pan"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
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

// RunShareList 执行列出分享列表
func RunShareList() {
	records, err := GetBaiduPCS().ShareList(1)
	if err != nil {
		fmt.Printf("%s失败: %s\n", baidupcs.OperationShareList, err)
		return
	}

	tb := pcstable.NewTable(os.Stdout)
	tb.SetHeader([]string{"#", "ShareID", "分享链接", "提取密码", "根目录", "路径"})
	for k, record := range records {
		if record == nil {
			continue
		}

		tb.Append([]string{strconv.Itoa(k), strconv.FormatInt(record.ShareID, 10), record.Shortlink, record.Passwd, record.TypicalPath[:strings.LastIndex(record.TypicalPath, "/")+1], record.TypicalPath})
	}
	tb.Render()
}

func getShareDLink(pcspath string) (dlink string) {
	var (
		pcs = GetBaiduPCS()
	)

	for page := 1; ; page++ {
		records, pcsError := pcs.ShareList(page)
		if pcsError != nil {
			pcsCommandVerbose.Warn(pcsError.Error())
			break
		}

		// 完成
		if len(records) == 0 {
			break
		}

		for _, record := range records {
			if record == nil {
				continue
			}

			rootSharePath := path.Dir(record.TypicalPath)
			if len(record.FsIds) == 1 {
				if strings.HasPrefix(pcspath, record.TypicalPath) {
					dlink = getLink(record.ShareID, record.Shortlink, record.Passwd, rootSharePath, strings.TrimPrefix(pcspath, rootSharePath))
					return
				}
				continue
			}

			// 尝试获取
			if strings.HasPrefix(pcspath, rootSharePath) {
				dlink = getLink(record.ShareID, record.Shortlink, record.Passwd, rootSharePath, strings.TrimPrefix(pcspath, rootSharePath))
				if dlink != "" {
					return
				}
				continue
			}
		}
	}

	pcsCommandVerbose.Infof("%s: 未在已分享列表中找到分享信息\n", pcspath)
	s, pcsError := pcs.ShareSet([]string{pcspath}, nil)
	if pcsError != nil {
		pcsCommandVerbose.Warn(pcsError.Error())
		return ""
	}

	// 取消分享
	defer pcs.ShareCancel([]int64{s.ShareID})

	dlink = getLink(s.ShareID, s.Link, "", path.Dir(pcspath), path.Base(pcspath))
	return
}

func getLink(shareID int64, shareLink, passwd, rootSharePath, filePath string) (dlink string) {
	sInfo := pan.NewSharedInfo(shareLink)
	sInfo.Client = requester.NewHTTPClient()
	sInfo.Client.SetHTTPSecure(pcsconfig.Config.EnableHTTPS())

	if passwd != "" {
		err := sInfo.Auth(passwd)
		if err != nil {
			pcsCommandVerbose.Warn(err.Error())
			return ""
		}
	}

	uk, pcsError := GetBaiduPCS().UK()
	if pcsError != nil {
		pcsCommandVerbose.Warn(pcsError.Error())
		err := sInfo.InitInfo()
		if err != nil {
			pcsCommandVerbose.Warn(err.Error())
			return ""
		}
	} else {
		sInfo.UK = uk
		sInfo.ShareID = shareID
		sInfo.RootSharePath = rootSharePath
	}

	fd, err := sInfo.Meta(filePath)
	if err != nil {
		pcsCommandVerbose.Warn(err.Error())
		return ""
	}

	u, err := url.Parse(fd.Dlink)
	if err != nil {
		pcsCommandVerbose.Warn(err.Error())
		return ""
	}

	if pcsconfig.Config.EnableHTTPS() {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}

	return u.String()
}
