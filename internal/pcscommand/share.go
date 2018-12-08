package pcscommand

import (
	"errors"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	// ErrShareInfoNotFound 未在已分享列表中找到分享信息
	ErrShareInfoNotFound = errors.New("未在已分享列表中找到分享信息")
)

// RunShareSet 执行分享
func RunShareSet(paths []string, option *baidupcs.ShareOption) {
	pcspaths, err := matchPathByShellPattern(paths...)
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

		tb.Append([]string{strconv.Itoa(k), strconv.FormatInt(record.ShareID, 10), record.Shortlink, record.Passwd, path.Clean(path.Dir(record.TypicalPath)), record.TypicalPath})
	}
	tb.Render()
}

// getShareDLink pcspath 为文件的路径, 不是目录
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

			if record.Status != 0 { // 分享状态异常
				continue
			}

			if record.TypicalPath == baidupcs.PathSeparator { //TypicalPath为根目录
				continue
			}

			rootSharePath, _ := path.Split(record.TypicalPath)
			if rootSharePath == "" { // 分享状态异常
				continue
			}

			// 粗略搜索
			if len(record.FsIds) == 1 {
				switch record.TypicalCategory {
				case -1: // 文件夹
					if strings.HasPrefix(pcspath, record.TypicalPath+baidupcs.PathSeparator) {
						dlink, err = getLink(record.Shortlink, record.Passwd, pcspath, true)
						return
					}
				default: // 文件
					if pcspath == record.TypicalPath {
						dlink, err = getLink(record.Shortlink, record.Passwd, pcspath, false)
						return
					}
				}

				continue
			}

			// 尝试获取
			if strings.HasPrefix(pcspath, rootSharePath) {
				dlink, err = getLink(record.Shortlink, record.Passwd, pcspath, false)
				if err != nil {
					continue
				}
				return
			}
		}
	}

	if err != nil {
		return
	}
	return "", ErrShareInfoNotFound
}

func getLink(shareLink, passwd, filePath string, skipRoot bool) (dlink string, err error) {
	dc := pcsconfig.Config.DlinkClient()
	short, err := dc.CacheShareReg(shareLink, passwd)
	if err != nil {
		return
	}

	var dir string
	if skipRoot {
		dir = path.Dir(filePath)
	} else {
		rfl, err := dc.CacheShareList(short, baidupcs.PathSeparator, 1)
		if err != nil {
			return "", err
		}

		for _, rf := range rfl {
			if rf.Isdir == 1 {
				if strings.HasPrefix(filePath, rf.Path+baidupcs.PathSeparator) {
					dir = path.Dir(filePath)
					break
				}
				continue
			}

			if rf.Path == filePath {
				dlink, err = dc.LinkRedirect(rf.Link)
				if err != nil {
					return "", ErrDlinkNotFound
				}
				return dlink, err
			}
		}
	}

	if dir == "" {
		return "", ErrDlinkNotFound
	}

	for page := 1; ; page++ {
		list, err := dc.CacheShareList(short, dir, page)
		if err != nil {
			return "", err
		}
		if len(list) == 0 {
			break
		}

		for _, f := range list {
			if f.Path == filePath {
				dlink, err = dc.CacheLinkRedirect(f.Link)
				if err != nil {
					return "", ErrDlinkNotFound
				}
				return dlink, err
			}
		}
		if len(list) < 100 {
			break
		}
	}

	return "", ErrDlinkNotFound
}
