package baidupcs

import (
	"bytes"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/requester/multipartreader"
	"github.com/json-iterator/go"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
)

func handleRespClose(resp *http.Response) error {
	if resp != nil {
		return resp.Body.Close()
	}
	return nil
}

func handleRespStatusError(opreation string, resp *http.Response) Error {
	errInfo := &ErrInfo{
		operation: opreation,
		errType:   ErrTypeInternalError,
	}

	if resp == nil {
		errInfo.err = fmt.Errorf("resp is nil")
		return errInfo
	}

	errInfo.errType = ErrTypeNetError

	// http 响应错误处理
	switch resp.StatusCode {
	case 413: // Request Entity Too Large
		// 上传的文件太大了
		resp.Body.Close()
		errInfo.err = fmt.Errorf("http 响应错误, %s", resp.Status)
		return errInfo
	}

	return nil
}

// PrepareQuotaInfo 获取当前用户空间配额信息, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareQuotaInfo() (dataReadCloser io.ReadCloser, pcsError Error) {
	pcs.lazyInit()
	pcsURL := pcs.generatePCSURL("quota", "info")
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationQuotaInfo, pcsURL)

	resp, err := pcs.client.Req("GET", pcsURL.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &ErrInfo{
			operation: OperationQuotaInfo,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareFilesDirectoriesBatchMeta 获取多个文件/目录的元信息, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareFilesDirectoriesBatchMeta(paths ...string) (dataReadCloser io.ReadCloser, pcsError Error) {
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
		return nil, &ErrInfo{
			operation: OperationFilesDirectoriesMeta,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareFilesDirectoriesList 获取目录下的文件和目录列表, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareFilesDirectoriesList(path string) (dataReadCloser io.ReadCloser, pcsError Error) {
	pcs.lazyInit()
	if path == "" {
		path = "/"
	}

	pcsURL := pcs.generatePCSURL("file", "list", map[string]string{
		"path":  path,
		"by":    "name",
		"order": "asc", // 升序
		"limit": "0-2147483647",
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationFilesDirectoriesList, pcsURL)

	resp, err := pcs.client.Req("GET", pcsURL.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &ErrInfo{
			operation: OperationFilesDirectoriesList,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareRemove 批量删除文件/目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRemove(paths ...string) (dataReadCloser io.ReadCloser, pcsError Error) {
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
		return nil, &ErrInfo{
			operation: OperationRemove,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareMkdir 创建目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareMkdir(pcspath string) (dataReadCloser io.ReadCloser, pcsError Error) {
	pcs.lazyInit()
	pcsURL := pcs.generatePCSURL("file", "mkdir", map[string]string{
		"path": pcspath,
	})
	baiduPCSVerbose.Infof("%s URL: %s", OperationMkdir, pcsURL)

	resp, err := pcs.client.Req("POST", pcsURL.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &ErrInfo{
			operation: OperationMkdir,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	return resp.Body, nil
}

func (pcs *BaiduPCS) prepareCpMvOp(op string, cpmvJSON ...*CpMvJSON) (dataReadCloser io.ReadCloser, pcsError Error) {
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

	errInfo := NewErrorInfo(op)

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
		errInfo.errType = ErrTypeNetError
		errInfo.err = err
		return nil, errInfo
	}

	return resp.Body, nil
}

// PrepareRename 重命名文件/目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRename(from, to string) (dataReadCloser io.ReadCloser, pcsError Error) {
	return pcs.prepareCpMvOp(OperationRename, &CpMvJSON{
		From: from,
		To:   to,
	})
}

// PrepareCopy 批量拷贝文件/目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCopy(cpmvJSON ...*CpMvJSON) (dataReadCloser io.ReadCloser, pcsError Error) {
	return pcs.prepareCpMvOp(OperationCopy, cpmvJSON...)
}

// PrepareMove 批量移动文件/目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareMove(cpmvJSON ...*CpMvJSON) (dataReadCloser io.ReadCloser, pcsError Error) {
	return pcs.prepareCpMvOp(OperationMove, cpmvJSON...)
}

// PrepareRapidUpload 秒传文件, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRapidUpload(targetPath, contentMD5, sliceMD5, crc32 string, length int64) (dataReadCloser io.ReadCloser, pcsError Error) {
	pcs.lazyInit()
	pcsError = pcs.checkIsdir(OperationRapidUpload, targetPath)
	if pcsError != nil {
		return nil, pcsError
	}

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
		return nil, &ErrInfo{
			operation: OperationRapidUpload,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareUpload 上传单个文件, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareUpload(targetPath string, uploadFunc UploadFunc) (dataReadCloser io.ReadCloser, pcsError Error) {
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

	resp, err := uploadFunc(pcsURL.String(), pcs.client.Jar.(*cookiejar.Jar))
	if err != nil {
		handleRespClose(resp)
		return nil, &ErrInfo{
			operation: OperationUpload,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	pcsError = handleRespStatusError(OperationUpload, resp)
	if pcsError != nil {
		return
	}

	return resp.Body, nil
}

// PrepareUploadTmpFile 分片上传—文件分片及上传, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareUploadTmpFile(uploadFunc UploadFunc) (dataReadCloser io.ReadCloser, pcsError Error) {
	pcs.lazyInit()
	pcsURL := pcs.generatePCSURL("file", "upload", map[string]string{
		"type": "tmpfile",
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationUploadTmpFile, pcsURL)

	resp, err := uploadFunc(pcsURL.String(), pcs.client.Jar.(*cookiejar.Jar))
	if err != nil {
		handleRespClose(resp)
		return nil, &ErrInfo{
			operation: OperationUploadTmpFile,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	pcsError = handleRespStatusError(OperationUpload, resp)
	if pcsError != nil {
		return
	}

	return resp.Body, nil
}

// PrepareUploadCreateSuperFile 分片上传—合并分片文件, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareUploadCreateSuperFile(targetPath string, blockList ...string) (dataReadCloser io.ReadCloser, pcsError Error) {
	pcs.lazyInit()
	pcsError = pcs.checkIsdir(OperationUploadCreateSuperFile, targetPath)
	if pcsError != nil {
		return nil, pcsError
	}

	bl := &struct {
		BlockList []string `json:"block_list"`
	}{
		BlockList: blockList,
	}

	sendData, err := jsoniter.Marshal(bl)
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
		return nil, &ErrInfo{
			operation: OperationUploadCreateSuperFile,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareCloudDlAddTask 添加离线下载任务, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlAddTask(sourceURL, savePath string) (dataReadCloser io.ReadCloser, pcsError Error) {
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
		return nil, &ErrInfo{
			operation: OperationCloudDlAddTask,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareCloudDlQueryTask 精确查询离线下载任务, 只返回服务器响应数据和错误信息,
// taskids 例子: 12123,234234,2344, 用逗号隔开多个 task_id
func (pcs *BaiduPCS) PrepareCloudDlQueryTask(taskIDs string) (dataReadCloser io.ReadCloser, pcsError Error) {
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
		return nil, &ErrInfo{
			operation: OperationCloudDlQueryTask,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareCloudDlListTask 查询离线下载任务列表, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlListTask() (dataReadCloser io.ReadCloser, pcsError Error) {
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
		return nil, &ErrInfo{
			operation: OperationCloudDlListTask,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	return resp.Body, nil
}

func (pcs *BaiduPCS) prepareCloudDlCDTask(opreation, method string, taskID int64) (dataReadCloser io.ReadCloser, pcsError Error) {
	pcs.lazyInit()
	pcsURL2 := pcs.generatePCSURL2("services/cloud_dl", method, map[string]string{
		"task_id": strconv.FormatInt(taskID, 10),
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", opreation, pcsURL2)

	resp, err := pcs.client.Req("POST", pcsURL2.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, &ErrInfo{
			operation: opreation,
			errType:   ErrTypeNetError,
			err:       err,
		}
	}

	return resp.Body, nil
}

// PrepareCloudDlCancelTask 取消离线下载任务, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlCancelTask(taskID int64) (dataReadCloser io.ReadCloser, pcsError Error) {
	return pcs.prepareCloudDlCDTask(OperationCloudDlCancelTask, "cancel_task", taskID)
}

// PrepareCloudDlDeleteTask 取消离线下载任务, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlDeleteTask(taskID int64) (dataReadCloser io.ReadCloser, pcsError Error) {
	return pcs.prepareCloudDlCDTask(OperationCloudDlDeleteTask, "delete_task", taskID)
}
