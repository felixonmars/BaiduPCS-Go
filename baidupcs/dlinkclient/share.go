package dlinkclient

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
	"strconv"
)

type (
	// FileDirectory 文件/目录信息
	FileDirectory struct {
		Path     string `json:"path"`
		Filename string `json:"filename"`
		Ctime    int64  `json:"ctime"`
		Mtime    int64  `json:"mtime"`
		MD5      string `json:"md5"`
		Size     int64  `json:"size"`
		Isdir    int    `json:"isdir"`
		Link     string `json:"link"`
	}

	// FDList 文件列表
	FDList struct {
		*pcserror.DlinkErrInfo
		List []*FileDirectory `json:"list"`
	}

	// RegStat 注册状态
	RegStat struct {
		*pcserror.DlinkErrInfo
		Short string `json:"short"`
	}

	RedirectRes struct {
		*pcserror.DlinkErrInfo
		Link string `json:"link"`
	}
)

// ShareReg reg
func (dc *DlinkClient) ShareReg(shareURL, passwd string) (short string, dlinkError pcserror.Error) {
	dc.lazyInit()

	var (
		u = dc.genShareURL("reg", map[string]string{
			"share_url": shareURL,
			"passwd":    passwd,
		})
		regStat = RegStat{
			DlinkErrInfo: pcserror.NewDlinkErrInfo(OperationReg),
		}
	)

	resp, err := dc.client.Req("GET", u.String(), nil, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		regStat.SetNetError(err)
		return "", regStat.DlinkErrInfo
	}

	dlinkError = handleJSONParse(OperationReg, resp.Body, &regStat)
	if dlinkError != nil {
		return
	}

	return regStat.Short, nil
}

func (dc *DlinkClient) ShareList(short, dir string, page int) (fds []*FileDirectory, dlinkError pcserror.Error) {
	dc.lazyInit()

	var (
		u = dc.genShareURL("list", map[string]string{
			"short": short,
			"dir":   dir,
			"page":  strconv.Itoa(page),
		})
		fdList = FDList{
			DlinkErrInfo: pcserror.NewDlinkErrInfo(OperationList),
		}
	)

	resp, err := dc.client.Req("GET", u.String(), nil, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		fdList.SetNetError(err)
		return nil, fdList.DlinkErrInfo
	}

	dlinkError = handleJSONParse(OperationList, resp.Body, &fdList)
	if dlinkError != nil {
		return
	}

	return fdList.List, nil
}
