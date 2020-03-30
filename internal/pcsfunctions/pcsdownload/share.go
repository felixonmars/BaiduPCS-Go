package pcsdownload

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"path"
	"strings"
)

// GetShareDLink pcspath 为文件的路径, 不是目录
func GetShareDLink(pcs *baidupcs.BaiduPCS, pcspath string) (dlink string, err error) {
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
						dlink, err = GetLinkByRecord(pcs, record, pcspath, true)
						return
					}
				default: // 文件
					if pcspath == record.TypicalPath {
						dlink, err = GetLinkByRecord(pcs, record, pcspath, false)
						return
					}
				}

				continue
			}

			// 尝试获取
			if strings.HasPrefix(pcspath, rootSharePath) {
				dlink, err = GetLinkByRecord(pcs, record, pcspath, false)
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

// GetShareRecordPasswd 获取Passwd
func GetLinkByRecord(pcs *baidupcs.BaiduPCS, record *baidupcs.ShareRecordInfo, filePath string, skipRoot bool) (dlink string, err error) {
	if record.Public == 0 {
		// 私密分享
		info, pcsError := pcs.ShareSURLInfo(record.ShareID)
		if pcsError != nil {
			// 获取错误
			return "", pcsError
		}
		record.Passwd = info.Pwd
	}
	return GetLink(record.Shortlink, record.Passwd, filePath, skipRoot)
}

func GetLink(shareLink, passwd, filePath string, skipRoot bool) (dlink string, err error) {
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
				dlink, err = dc.CacheLinkRedirect(rf.Link)
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
