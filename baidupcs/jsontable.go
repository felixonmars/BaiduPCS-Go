package baidupcs

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/json-iterator/go"
	"io"
	"strconv"
	"strings"
)

type (
	// PathJSON 网盘路径
	PathJSON struct {
		Path string `json:"path"`
	}

	// PathsListJSON 网盘路径列表
	PathsListJSON struct {
		List []*PathJSON `json:"list"`
	}

	// CpMvJSON 源文件目录的地址和目标文件目录的地址
	CpMvJSON struct {
		From string `json:"from"` // 源文件或目录
		To   string `json:"to"`   // 目标文件或目录
	}

	// CpMvListJSON []*CpMvJSON 对象数组
	CpMvListJSON struct {
		List []*CpMvJSON `json:"list"`
	}

	// BlockListJSON 文件分块信息JSON
	BlockListJSON struct {
		BlockList []string `json:"block_list"`
	}
)

// JSON json 数据构造
func (plj *PathsListJSON) JSON(paths ...string) (data []byte, err error) {
	plj.List = make([]*PathJSON, len(paths))

	for k := range paths {
		plj.List[k] = &PathJSON{
			Path: paths[k],
		}
	}

	data, err = jsoniter.Marshal(plj)
	return
}

// JSON json 数据构造
func (cj *CpMvJSON) JSON() (data []byte, err error) {
	data, err = jsoniter.Marshal(cj)
	return
}

// JSON json 数据构造
func (clj *CpMvListJSON) JSON() (data []byte, err error) {
	data, err = jsoniter.Marshal(clj)
	return
}

func (clj *CpMvListJSON) String() string {
	builder := &strings.Builder{}

	tb := pcstable.NewTable(builder)
	tb.SetHeader([]string{"#", "原路径", "目标路径"})

	for k := range clj.List {
		if clj.List[k] == nil {
			continue
		}
		tb.Append([]string{strconv.Itoa(k), clj.List[k].From, clj.List[k].To})
	}

	tb.Render()
	return builder.String()
}

func handleJSONParse(op string, data io.Reader, info interface{}) (pcsError pcserror.Error) {
	var (
		d       = jsoniter.NewDecoder(data)
		err     error
		errInfo pcserror.Error
	)
	switch op {
	case OperationGetUK:
		userInfo := info.(*userInfoJSON)
		err = d.Decode(userInfo)
		errInfo = userInfo.PanErrorInfo

	case OperationQuotaInfo:
		jsonData := info.(*quotaInfo)
		err = d.Decode(jsonData)
		errInfo = jsonData.PCSErrInfo

	case OperationFilesDirectoriesMeta, OperationFilesDirectoriesList, OperationSearch:
		jsonData := info.(*fdDataJSONExport)
		err = d.Decode(jsonData)
		errInfo = jsonData.PCSErrInfo

	case OperationUpload:
		jsonData := info.(*uploadJSON)
		err = d.Decode(jsonData)
		errInfo = jsonData.PCSErrInfo

	case OperationUploadTmpFile:
		jsonData := info.(*uploadTmpFileJSON)
		err = d.Decode(jsonData)
		errInfo = jsonData.PCSErrInfo

	case OperationUploadPrecreate:
		jsonData := info.(*uploadPrecreateJSON)
		err = d.Decode(jsonData)
		errInfo = jsonData.PanErrorInfo

	case OperationLocateDownload:
		jsonData := info.(*locateDownloadJSON)
		err = d.Decode(jsonData)
		errInfo = jsonData.PCSErrInfo

	case OperationCloudDlAddTask:
		jsonData := info.(*cloudDlAddTaskJSON)
		err = d.Decode(jsonData)
		errInfo = jsonData.PCSErrInfo

	case OperationCloudDlQueryTask:
		jsonData := info.(*cloudDlQueryTaskJSON)
		err = d.Decode(jsonData)
		errInfo = jsonData.PCSErrInfo

	case OperationCloudDlListTask:
		jsonData := info.(*cloudDlListTaskJSON)
		err = d.Decode(jsonData)
		errInfo = jsonData.PCSErrInfo

	case OperationShareSet:
		jsonData := info.(*sharePSetJSON)
		err = d.Decode(jsonData)
		errInfo = jsonData.PanErrorInfo

	case OperationShareList:
		jsonData := info.(*shareListJSON)
		err = d.Decode(jsonData)
		errInfo = jsonData.PanErrorInfo

	default:
		panic("unknown op")
	}

	if errInfo == nil {
		errInfo = pcserror.NewPCSErrorInfo(op)
	}

	if err != nil {
		errInfo.SetJSONError(err)
		return errInfo
	}

	if errInfo.GetRemoteErrCode() != 0 {
		errInfo.SetRemoteError()
		return errInfo
	}

	return nil
}
