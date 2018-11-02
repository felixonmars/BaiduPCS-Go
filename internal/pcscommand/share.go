package pcscommand

import (
	"errors"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/baidupcs/dlinkclient"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	ppath "github.com/iikira/BaiduPCS-Go/pcspath"
	"github.com/iikira/BaiduPCS-Go/pcstable"
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
func RunShareList(page int) {
	if page < 1 {
		page = 1
	}
	records, err := GetBaiduPCS().ShareList(page)
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

		tb.Append([]string{strconv.Itoa(k), strconv.FormatInt(record.ShareID, 10), record.Shortlink, record.Passwd, strings.TrimSuffix(record.TypicalPath[:strings.LastIndex(record.TypicalPath, "/")+1], "/"), record.TypicalPath})
	}
	tb.Render()
}

func getShareDLink(pcspath string) (dlink string, err error) {
	var (
		pcs = GetBaiduPCS()
	)

	for page := 1; ; page++ {
		records, pcsError := pcs.ShareList(page)
		if pcsError != nil {
			return "", pcsError
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
					dlink = getLink(record.Shortlink, record.Passwd, ppath.TrimPrefix(pcspath, rootSharePath))
					if dlink != "" {
						return
					}
				}
				continue
			}

			// 尝试获取
			if strings.HasPrefix(pcspath, rootSharePath) {
				dlink = getLink(record.Shortlink, record.Passwd, ppath.TrimPrefix(pcspath, rootSharePath))
				if dlink != "" {
					return
				}
				continue
			}
		}
	}

	return "", errors.New("未在已分享列表中找到分享信息")
}

func getLink(shareLink, passwd, filePath string) (dlink string) {
	dc := dlinkclient.NewDlinkClient()
	dc.SetClient(pcsconfig.Config.HTTPClient())
	short, err := dc.CacheShareReg(shareLink, passwd)
	if err != nil {
		pcsCommandVerbose.Warnf("%s\n", err)
		return
	}

	for page := 1; ; page++ {
		list, err := dc.CacheShareList(short, path.Dir(filePath), page)
		if err != nil {
			pcsCommandVerbose.Warnf("%s\n", err)
			return
		}
		if len(list) == 0 {
			break
		}

		for _, f := range list {
			if strings.Compare(f.Filename, path.Base(filePath)) == 0 {
				dlink, err = dc.CacheLinkRedirect(f.Link)
				if err != nil {
					pcsCommandVerbose.Warnf("%s\n", err)
				}
				return
			}
		}
	}

	return
}
