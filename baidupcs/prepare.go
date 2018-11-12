package baidupcs

import (
	"bytes"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"github.com/iikira/BaiduPCS-Go/requester/multipartreader"
	"github.com/json-iterator/go"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unsafe"
)

func handleRespClose(resp *http.Response) error {
	if resp != nil {
		return resp.Body.Close()
	}
	return nil
}

func handleRespStatusError(opreation string, resp *http.Response) pcserror.Error {
	errInfo := pcserror.NewPCSErrorInfo(opreation)
	// http 响应错误处理
	switch resp.StatusCode {
	case 413: // Request Entity Too Large
		// 上传的文件太大了
		resp.Body.Close()
		errInfo.SetNetError(fmt.Errorf("http 响应错误, %s", resp.Status))
		return errInfo
	}

	return nil
}

// PrepareUK 获取用户 UK, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareUK() (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL := GetHTTPScheme(pcs.isHTTPS) + "://pan.baidu.com/api/user/getinfo?need_selfinfo=1"

	errInfo := pcserror.NewPanErrorInfo(OperationGetUK)
	resp, err := pcs.client.Req("GET", pcsURL, nil, netdiskUAHeader)
	if err != nil {
		handleRespClose(resp)
		errInfo.SetNetError(err)
		return nil, errInfo
	}
	return resp.Body, nil
}

// PrepareQuotaInfo 获取当前用户空间配额信息, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareQuotaInfo() (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL := pcs.generatePCSURL("quota", "info")
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationQuotaInfo, pcsURL)

	resp, err := pcs.client.Req("GET", pcsURL.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationQuotaInfo,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareFilesDirectoriesBatchMeta 获取多个文件/目录的元信息, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareFilesDirectoriesBatchMeta(paths ...string) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	sendData, err := (&PathsListJSON{}).JSON(paths...)
	if err != nil {
		panic(OperationFilesDirectoriesMeta + ", json 数据构造失败, " + err.Error())
	}

	pcsURL := pcs.generatePCSURL("file", "meta")
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationFilesDirectoriesMeta, pcsURL)

	// 表单上传
	mr := multipartreader.NewMultipartReader()
	mr.AddFormFeild("param", bytes.NewReader(sendData))
	mr.CloseMultipart()

	resp, err := pcs.client.Req("POST", pcsURL.String(), mr, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationFilesDirectoriesMeta,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareFilesDirectoriesList 获取目录下的文件和目录列表, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareFilesDirectoriesList(path string, options *OrderOptions) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	if options == nil {
		options = DefaultOrderOptions
	}
	if path == "" {
		path = "/"
	}

	pcsURL := pcs.generatePCSURL("file", "list", map[string]string{
		"path":  path,
		"by":    *(*string)(unsafe.Pointer(&options.By)),
		"order": *(*string)(unsafe.Pointer(&options.Order)),
		"limit": "0-2147483647",
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationFilesDirectoriesList, pcsURL)

	resp, err := pcs.client.Req("GET", pcsURL.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationFilesDirectoriesList,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareSearch 按文件名搜索文件, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareSearch(targetPath, keyword string, recursive bool) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	var re string
	if recursive {
		re = "1"
	} else {
		re = "0"
	}
	pcsURL := pcs.generatePCSURL("file", "search", map[string]string{
		"path": targetPath,
		"wd":   keyword,
		"re":   re,
	})

	resp, err := pcs.client.Req("GET", pcsURL.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationSearch,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareRemove 批量删除文件/目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRemove(paths ...string) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	sendData, err := (&PathsListJSON{}).JSON(paths...)
	if err != nil {
		panic(OperationMove + ", json 数据构造失败, " + err.Error())
	}

	pcsURL := pcs.generatePCSURL("file", "delete")
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationRemove, pcsURL)

	// 表单上传
	mr := multipartreader.NewMultipartReader()
	mr.AddFormFeild("param", bytes.NewReader(sendData))
	mr.CloseMultipart()

	resp, err := pcs.client.Req("POST", pcsURL.String(), mr, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationRemove,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareMkdir 创建目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareMkdir(pcspath string) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL := pcs.generatePCSURL("file", "mkdir", map[string]string{
		"path": pcspath,
	})
	baiduPCSVerbose.Infof("%s URL: %s", OperationMkdir, pcsURL)

	resp, err := pcs.client.Req("POST", pcsURL.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationMkdir,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

func (pcs *BaiduPCS) prepareCpMvOp(op string, cpmvJSON ...*CpMvJSON) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	var method string
	switch op {
	case OperationCopy:
		method = "copy"
	case OperationMove, OperationRename:
		method = "move"
	default:
		panic("Unknown opreation: " + op)
	}

	errInfo := pcserror.NewPCSErrorInfo(op)
	sendData, err := (&CpMvListJSON{
		List: cpmvJSON,
	}).JSON()
	if err != nil {
		//json 数据生成失败
		panic(err)
	}

	pcsURL := pcs.generatePCSURL("file", method)
	baiduPCSVerbose.Infof("%s URL: %s\n", op, pcsURL)

	// 表单上传
	mr := multipartreader.NewMultipartReader()
	mr.AddFormFeild("param", bytes.NewReader(sendData))
	mr.CloseMultipart()

	resp, err := pcs.client.Req("POST", pcsURL.String(), mr, nil)
	if err != nil {
		handleRespClose(resp)
		errInfo.SetNetError(err)
		return nil, errInfo
	}

	return resp.Body, nil
}

// PrepareRename 重命名文件/目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRename(from, to string) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	return pcs.prepareCpMvOp(OperationRename, &CpMvJSON{
		From: from,
		To:   to,
	})
}

// PrepareCopy 批量拷贝文件/目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCopy(cpmvJSON ...*CpMvJSON) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	return pcs.prepareCpMvOp(OperationCopy, cpmvJSON...)
}

// PrepareMove 批量移动文件/目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareMove(cpmvJSON ...*CpMvJSON) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	return pcs.prepareCpMvOp(OperationMove, cpmvJSON...)
}

// prepareRapidUpload 秒传文件, 不进行文件夹检查
func (pcs *BaiduPCS) prepareRapidUpload(targetPath, contentMD5, sliceMD5, crc32 string, length int64) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL := pcs.generatePCSURL("file", "rapidupload", map[string]string{
		"path":           targetPath,                    // 上传文件的全路径名
		"content-length": strconv.FormatInt(length, 10), // 待秒传的文件长度
		"content-md5":    contentMD5,                    // 待秒传的文件的MD5
		"slice-md5":      sliceMD5,                      // 待秒传的文件前256kb的MD5
		"content-crc32":  crc32,                         // 待秒传文件CRC32
		"ondup":          "overwrite",                   // overwrite: 表示覆盖同名文件; newcopy: 表示生成文件副本并进行重命名，命名规则为“文件名_日期.后缀”
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationRapidUpload, pcsURL)

	resp, err := pcs.client.Req("POST", pcsURL.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationRapidUpload,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareRapidUpload 秒传文件, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRapidUpload(targetPath, contentMD5, sliceMD5, crc32 string, length int64) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsError = pcs.checkIsdir(OperationRapidUpload, targetPath)
	if pcsError != nil {
		return nil, pcsError
	}

	return pcs.prepareRapidUpload(targetPath, contentMD5, sliceMD5, crc32, length)
}

// prepareLocateDownload 获取下载链接, 可指定 User-Agent
func (pcs *BaiduPCS) prepareLocateDownload(pcspath, userAgent string) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL := &url.URL{
		Scheme: GetHTTPScheme(pcs.isHTTPS),
		Host:   PCSBaiduCom,
		Path:   "/rest/2.0/pcs/file",
		RawQuery: (url.Values{
			"app_id": []string{PanAppID},
			"method": []string{"locatedownload"},
			"path":   []string{pcspath},
			"ver":    []string{"2"},
		}).Encode(),
	}
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationLocateDownload, pcsURL)

	var header map[string]string
	if userAgent != "" {
		header = netdiskUAHeader
	}

	resp, err := pcs.client.Req("GET", pcsURL.String(), nil, header)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationLocateDownload,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareLocateDownload 获取下载链接, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareLocateDownload(pcspath string) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	return pcs.prepareLocateDownload(pcspath, "")
}

// PrepareLocatePanAPIDownload 从百度网盘首页获取下载链接, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareLocatePanAPIDownload(fidList ...int64) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	// 初始化
	var (
		sign, err = pcs.ph.CacheSignature()
	)
	if err != nil {
		return nil, &pcserror.PanErrorInfo{
			Operation: OperationLocatePanAPIDownload,
			ErrType:   pcserror.ErrTypeOthers,
			Err:       err,
		}
	}

	panURL := pcs.generatePanURL("download", nil)
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationLocatePanAPIDownload, panURL)

	resp, err := pcs.client.Req("POST", panURL.String(), map[string]string{
		"sign":      sign.Sign(),
		"timestamp": sign.Timestamp(),
		"fidlist":   mergeInt64List(fidList...),
	}, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   NetdiskUA,
	})
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PanErrorInfo{
			Operation: OperationLocatePanAPIDownload,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareUpload 上传单个文件, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareUpload(targetPath string, uploadFunc UploadFunc) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsError = pcs.checkIsdir(OperationUpload, targetPath)
	if pcsError != nil {
		return nil, pcsError
	}

	pcsURL := pcs.generatePCSURL("file", "upload", map[string]string{
		"path":  targetPath,
		"ondup": "overwrite",
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationUpload, pcsURL)

	resp, err := uploadFunc(pcsURL.String(), pcs.client.Jar)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationUpload,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	pcsError = handleRespStatusError(OperationUpload, resp)
	if pcsError != nil {
		return
	}

	return resp.Body, nil
}

// PrepareUploadTmpFile 分片上传—文件分片及上传, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareUploadTmpFile(uploadFunc UploadFunc) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL := pcs.generatePCSURL("file", "upload", map[string]string{
		"type": "tmpfile",
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationUploadTmpFile, pcsURL)

	resp, err := uploadFunc(pcsURL.String(), pcs.client.Jar)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationUploadTmpFile,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	pcsError = handleRespStatusError(OperationUploadTmpFile, resp)
	if pcsError != nil {
		return
	}

	return resp.Body, nil
}

// PrepareUploadCreateSuperFile 分片上传—合并分片文件, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareUploadCreateSuperFile(targetPath string, blockList ...string) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsError = pcs.checkIsdir(OperationUploadCreateSuperFile, targetPath)
	if pcsError != nil {
		return nil, pcsError
	}

	bl := BlockListJSON{
		BlockList: blockList,
	}

	sendData, err := jsoniter.Marshal(&bl)
	if err != nil {
		panic(err)
	}

	pcsURL := pcs.generatePCSURL("file", "createsuperfile", map[string]string{
		"path":  targetPath,
		"ondup": "overwrite",
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationUploadCreateSuperFile, pcsURL)

	// 表单上传
	mr := multipartreader.NewMultipartReader()
	mr.AddFormFeild("param", bytes.NewReader(sendData))
	mr.CloseMultipart()

	resp, err := pcs.client.Req("POST", pcsURL.String(), mr, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationUploadCreateSuperFile,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareUploadPrecreate 分片上传—Precreate, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareUploadPrecreate(targetPath, contentMD5, sliceMD5, crc32 string, size int64, bolckList ...string) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL := &url.URL{
		Scheme: GetHTTPScheme(pcs.isHTTPS),
		Host:   PanBaiduCom,
		Path:   "api/precreate",
	}
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationUploadPrecreate, pcsURL)

	resp, err := pcs.client.Req("POST", pcsURL.String(), map[string]string{
		"path":         targetPath,
		"size":         strconv.FormatInt(size, 10),
		"isdir":        "0",
		"block_list":   mergeStringList(bolckList...),
		"autoinit":     "1",
		"content-md5":  contentMD5,
		"slice-md5":    sliceMD5,
		"contentCrc32": crc32,
		"rtype":        "2",
	}, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   NetdiskUA,
	})
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationUploadPrecreate,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	pcsError = handleRespStatusError(OperationUploadPrecreate, resp)
	if pcsError != nil {
		return
	}

	return resp.Body, nil
}

// PrepareUploadSuperfile2 另一个上传接口
func (pcs *BaiduPCS) PrepareUploadSuperfile2(uploadid, targetPath string, partseq int, partOffset int64, uploadFunc UploadFunc) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL := pcs.generatePCSURL("superfile2", "upload", map[string]string{
		"type":       "tmpfile",
		"path":       targetPath,
		"partseq":    strconv.Itoa(partseq),
		"partoffset": strconv.FormatInt(partOffset, 10),
		"uploadid":   uploadid,
		"vip":        "1",
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationUploadSuperfile2, pcsURL)

	resp, err := uploadFunc(pcsURL.String(), pcs.client.Jar)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationUploadSuperfile2,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	pcsError = handleRespStatusError(OperationUpload, resp)
	if pcsError != nil {
		return
	}

	return resp.Body, nil
}

// PrepareCloudDlAddTask 添加离线下载任务, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlAddTask(sourceURL, savePath string) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL2 := pcs.generatePCSURL2("services/cloud_dl", "add_task", map[string]string{
		"save_path":  savePath,
		"source_url": sourceURL,
		"timeout":    "2147483647",
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationCloudDlAddTask, pcsURL2)

	resp, err := pcs.client.Req("POST", pcsURL2.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationCloudDlAddTask,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareCloudDlQueryTask 精确查询离线下载任务, 只返回服务器响应数据和错误信息,
// taskids 例子: 12123,234234,2344, 用逗号隔开多个 task_id
func (pcs *BaiduPCS) PrepareCloudDlQueryTask(taskIDs string) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL2 := pcs.generatePCSURL2("services/cloud_dl", "query_task", map[string]string{
		"op_type": "1",
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationCloudDlQueryTask, pcsURL2)

	// 表单上传
	mr := multipartreader.NewMultipartReader()
	mr.AddFormFeild("task_ids", strings.NewReader(taskIDs))
	mr.CloseMultipart()

	resp, err := pcs.client.Req("POST", pcsURL2.String(), mr, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationCloudDlQueryTask,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareCloudDlListTask 查询离线下载任务列表, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlListTask() (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL2 := pcs.generatePCSURL2("services/cloud_dl", "list_task", map[string]string{
		"need_task_info": "1",
		"status":         "255",
		"start":          "0",
		"limit":          "1000",
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationCloudDlListTask, pcsURL2)

	resp, err := pcs.client.Req("POST", pcsURL2.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationCloudDlListTask,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

func (pcs *BaiduPCS) prepareCloudDlCDTask(opreation, method string, taskID int64) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL2 := pcs.generatePCSURL2("services/cloud_dl", method, map[string]string{
		"task_id": strconv.FormatInt(taskID, 10),
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", opreation, pcsURL2)

	resp, err := pcs.client.Req("POST", pcsURL2.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: opreation,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareCloudDlCancelTask 取消离线下载任务, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlCancelTask(taskID int64) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	return pcs.prepareCloudDlCDTask(OperationCloudDlCancelTask, "cancel_task", taskID)
}

// PrepareCloudDlDeleteTask 取消离线下载任务, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlDeleteTask(taskID int64) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	return pcs.prepareCloudDlCDTask(OperationCloudDlDeleteTask, "delete_task", taskID)
}

// PrepareCloudDlClearTask 清空离线下载任务记录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlClearTask() (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL2 := pcs.generatePCSURL2("services/cloud_dl", "clear_task")
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationCloudDlClearTask, pcsURL2)

	resp, err := pcs.client.Req("POST", pcsURL2.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &pcserror.PCSErrInfo{
			Operation: OperationCloudDlClearTask,
			ErrType:   pcserror.ErrTypeNetError,
			Err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareSharePSet 私密分享文件, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareSharePSet(paths []string, period int) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()
	pcsURL := &url.URL{
		Scheme: GetHTTPScheme(pcs.isHTTPS),
		Host:   PanBaiduCom,
		Path:   "share/pset",
	}

	errInfo := pcserror.NewPanErrorInfo(OperationShareSet)
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationShareSet, pcsURL)

	resp, err := pcs.client.Req("POST", pcsURL.String(), map[string]string{
		"path_list":    mergeStringList(paths...),
		"schannel":     "0",
		"channel_list": "[]",
		"period":       strconv.Itoa(period),
	}, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   NetdiskUA,
	})
	if err != nil {
		handleRespClose(resp)
		errInfo.SetNetError(err)
		return nil, errInfo
	}
	return resp.Body, nil
}

// PrepareShareCancel 取消分享, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareShareCancel(shareIDs []int64) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()

	pcsURL := &url.URL{
		Scheme: GetHTTPScheme(pcs.isHTTPS),
		Host:   PanBaiduCom,
		Path:   "share/cancel",
	}

	errInfo := pcserror.NewPanErrorInfo(OperationShareCancel)
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationShareCancel, pcsURL)

	ss := converter.SliceInt64ToString(shareIDs)

	resp, err := pcs.client.Req("POST", pcsURL.String(), map[string]string{
		"shareid_list": "[" + strings.Join(ss, ",") + "]",
	}, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   NetdiskUA,
	})
	if err != nil {
		handleRespClose(resp)
		errInfo.SetNetError(err)
		return nil, errInfo
	}
	return resp.Body, nil
}

// PrepareShareList 列出分享列表, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareShareList(page int) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()

	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	query.Set("desc", "1")
	query.Set("order", "time")

	pcsURL := &url.URL{
		Scheme:   GetHTTPScheme(pcs.isHTTPS),
		Host:     PanBaiduCom,
		Path:     "share/record",
		RawQuery: query.Encode(),
	}

	errInfo := pcserror.NewPanErrorInfo(OperationShareList)
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationShareList, pcsURL)

	resp, err := pcs.client.Req("GET", pcsURL.String(), nil, netdiskUAHeader)
	if err != nil {
		handleRespClose(resp)
		errInfo.SetNetError(err)
		return nil, errInfo
	}
	return resp.Body, nil
}

// PrepareRecycleList 列出回收站文件列表, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRecycleList(page int) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()

	panURL := pcs.generatePanURL("recycle/list", map[string]string{
		"num":  "100",
		"page": strconv.Itoa(page),
	})

	errInfo := pcserror.NewPanErrorInfo(OperationRecycleList)
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationRecycleList, panURL)

	resp, err := pcs.client.Req("GET", panURL.String(), nil, netdiskUAHeader)
	if err != nil {
		handleRespClose(resp)
		errInfo.SetNetError(err)
		return nil, errInfo
	}
	return resp.Body, nil
}

// PrepareRecycleRestore 还原回收站文件或目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRecycleRestore(fidList ...int64) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()

	pcsURL := pcs.generatePCSURL("file", "restore")

	errInfo := pcserror.NewPCSErrorInfo(OperationRecycleRestore)
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationRecycleRestore, pcsURL)

	fsIDList := make([]*FsIDJSON, 0, len(fidList))
	for k := range fidList {
		fsIDList = append(fsIDList, &FsIDJSON{
			FsID: fidList[k],
		})
	}
	fsIDListJSON := FsIDListJSON{
		List: fsIDList,
	}

	sendData, err := jsoniter.Marshal(&fsIDListJSON)
	if err != nil {
		panic(err)
	}

	// 表单上传
	mr := multipartreader.NewMultipartReader()
	mr.AddFormFeild("param", bytes.NewReader(sendData))
	mr.CloseMultipart()

	resp, err := pcs.client.Req("POST", pcsURL.String(), mr, nil)
	if err != nil {
		handleRespClose(resp)
		errInfo.SetNetError(err)
		return nil, errInfo
	}
	return resp.Body, nil
}

// PrepareRecycleDelete 删除回收站文件或目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRecycleDelete(fidList ...int64) (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()

	panURL := pcs.generatePanURL("recycle/delete", nil)

	errInfo := pcserror.NewPanErrorInfo(OperationRecycleDelete)
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationRecycleDelete, panURL)

	resp, err := pcs.client.Req("POST", panURL.String(), map[string]string{
		"fidlist": mergeInt64List(fidList...),
	}, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   NetdiskUA,
	})
	if err != nil {
		handleRespClose(resp)
		errInfo.SetNetError(err)
		return nil, errInfo
	}
	return resp.Body, nil
}

// PrepareRecycleClear 清空回收站, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRecycleClear() (dataReadCloser io.ReadCloser, pcsError pcserror.Error) {
	pcs.lazyInit()

	pcsURL := pcs.generatePCSURL("file", "delete", map[string]string{
		"type": "recycle",
	})

	errInfo := pcserror.NewPCSErrorInfo(OperationRecycleClear)
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationRecycleClear, pcsURL)

	resp, err := pcs.client.Req("GET", pcsURL.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		errInfo.SetNetError(err)
		return nil, errInfo
	}
	return resp.Body, nil
}
