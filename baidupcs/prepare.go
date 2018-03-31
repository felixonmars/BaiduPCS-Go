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
)

func handleRespClose(resp *http.Response) error {
	if resp != nil {
		return resp.Body.Close()
	}
	return nil
}

func handleRespStatusError(opreation string, resp *http.Response) (err error) {
	if resp == nil {
		return fmt.Errorf("resp is nil")
	}

	// http 响应错误处理
	switch resp.StatusCode {
	case 413: // Request Entity Too Large
		// 上传的文件太大了
		resp.Body.Close()
		err = fmt.Errorf("%s, http 响应错误, %s", opreation, resp.Status)
		return
	}

	return nil
}

// PrepareQuotaInfo 获取当前用户空间配额信息, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareQuotaInfo() (dataReadCloser io.ReadCloser, err error) {
	pcs.setPCSURL("quota", "info")

	resp, err := pcs.client.Req("GET", pcs.url.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 网络错误, %s", OperationQuotaInfo, err)
	}

	return resp.Body, nil
}

// PrepareFilesDirectoriesBatchMeta 获取多个文件/目录的元信息, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareFilesDirectoriesBatchMeta(paths ...string) (dataReadCloser io.ReadCloser, err error) {
	sendData, err := (&PathsListJSON{}).JSON(paths...)
	if err != nil {
		panic(OperationFilesDirectoriesBatchMeta + ", json 数据构造失败, " + err.Error())
	}

	pcs.setPCSURL("file", "meta")

	// 表单上传
	mr := multipartreader.NewMultipartReader()
	mr.AddFormFeild("param", bytes.NewReader(sendData))

	resp, err := pcs.client.Req("POST", pcs.url.String(), mr, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 网络错误, %s", OperationFilesDirectoriesBatchMeta, err)
	}

	return resp.Body, nil
}

// PrepareFilesDirectoriesList 获取目录下的文件和目录列表, 可选是否递归, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareFilesDirectoriesList(path string, recurse bool) (dataReadCloser io.ReadCloser, err error) {
	if path == "" {
		path = "/"
	}

	pcs.setPCSURL("file", "list", map[string]string{
		"path":  path,
		"by":    "name",
		"order": "asc", // 升序
		"limit": "0-2147483647",
	})

	resp, err := pcs.client.Req("GET", pcs.url.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 网络错误, %s", OperationFilesDirectoriesList, err)
	}

	return resp.Body, nil
}

// PrepareRemove 批量删除文件/目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRemove(paths ...string) (dataReadCloser io.ReadCloser, err error) {
	sendData, err := (&PathsListJSON{}).JSON(paths...)
	if err != nil {
		panic(OperationMove + ", json 数据构造失败, " + err.Error())
	}

	pcs.setPCSURL("file", "delete")

	// 表单上传
	mr := multipartreader.NewMultipartReader()
	mr.AddFormFeild("param", bytes.NewReader(sendData))

	resp, err := pcs.client.Req("POST", pcs.url.String(), mr, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 网络错误, %s", OperationRemove, err)
	}

	return resp.Body, nil
}

// PrepareMkdir 创建目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareMkdir(pcspath string) (dataReadCloser io.ReadCloser, err error) {
	pcs.setPCSURL("file", "mkdir", map[string]string{
		"path": pcspath,
	})

	resp, err := pcs.client.Req("POST", pcs.url.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 网络错误, %s", OperationMkdir, err)
	}

	return resp.Body, nil
}

func (pcs *BaiduPCS) prepareCpMvOp(op string, cpmvJSON ...*CpMvJSON) (dataReadCloser io.ReadCloser, err error) {
	var method string
	switch op {
	case OperationCopy:
		method = "copy"
	case OperationMove, OperationRename:
		method = "move"
	default:
		panic("Unknown opreation: " + op)
	}

	sendData, err := (&CpMvListJSON{
		List: cpmvJSON,
	}).JSON()
	if err != nil {
		return nil, err
	}

	pcs.setPCSURL("file", method)

	// 表单上传
	mr := multipartreader.NewMultipartReader()
	mr.AddFormFeild("param", bytes.NewReader(sendData))

	resp, err := pcs.client.Req("POST", pcs.url.String(), mr, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 网络错误, %s", op, err)
	}

	return resp.Body, nil
}

// PrepareRename 重命名文件/目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRename(from, to string) (dataReadCloser io.ReadCloser, err error) {
	return pcs.prepareCpMvOp(OperationRename, &CpMvJSON{
		From: from,
		To:   to,
	})
}

// PrepareCopy 批量拷贝文件/目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCopy(cpmvJSON ...*CpMvJSON) (dataReadCloser io.ReadCloser, err error) {
	return pcs.prepareCpMvOp(OperationCopy, cpmvJSON...)
}

// PrepareMove 批量移动文件/目录, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareMove(cpmvJSON ...*CpMvJSON) (dataReadCloser io.ReadCloser, err error) {
	return pcs.prepareCpMvOp(OperationMove, cpmvJSON...)
}

// PrepareRapidUpload 秒传文件, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareRapidUpload(targetPath, contentMD5, sliceMD5, crc32 string, length int64) (dataReadCloser io.ReadCloser, err error) {
	// 检测文件是否存在于网盘路径
	// 很重要, 如果文件存在会直接覆盖!!! 即使是根目录!
	if !pcs.IsFile(targetPath) {
		return nil, fmt.Errorf("%s 遇到错误, 保存路径不可以覆盖目录", OperationRapidUpload)
	}

	pcs.setPCSURL("file", "rapidupload", map[string]string{
		"path":           targetPath,                    // 上传文件的全路径名
		"content-length": strconv.FormatInt(length, 10), // 待秒传的文件长度
		"content-md5":    contentMD5,                    // 待秒传的文件的MD5
		"slice-md5":      sliceMD5,                      // 待秒传的文件的MD5
		"content-crc32":  crc32,                         // 待秒传文件CRC32
		"ondup":          "overwrite",                   // overwrite: 表示覆盖同名文件; newcopy: 表示生成文件副本并进行重命名，命名规则为“文件名_日期.后缀”
	})

	resp, err := pcs.client.Req("POST", pcs.url.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return
	}

	return resp.Body, nil
}

// PrepareUpload 上传单个文件, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareUpload(targetPath string, uploadFunc UploadFunc) (dataReadCloser io.ReadCloser, err error) {
	if !pcs.IsFile(targetPath) {
		return nil, fmt.Errorf("%s, 遇到错误, 保存路径不可以覆盖目录", OperationUpload)
	}

	pcs.setPCSURL("file", "upload", map[string]string{
		"path":  targetPath,
		"ondup": "overwrite",
	})

	resp, err := uploadFunc(pcs.url.String(), pcs.client.Jar.(*cookiejar.Jar))
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 发生错误, %s", OperationUpload, err)
	}

	err = handleRespStatusError(OperationUpload, resp)
	if err != nil {
		return
	}

	return resp.Body, nil
}

// PrepareUploadTmpFile 分片上传—文件分片及上传, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareUploadTmpFile(uploadFunc UploadFunc) (dataReadCloser io.ReadCloser, err error) {
	pcs.setPCSURL("file", "upload", map[string]string{
		"type": "tmpfile",
	})

	resp, err := uploadFunc(pcs.url.String(), pcs.client.Jar.(*cookiejar.Jar))
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 发生错误, %s", OperationUploadTmpFile, err)
	}

	err = handleRespStatusError(OperationUpload, resp)
	if err != nil {
		return
	}

	return resp.Body, nil
}

// PrepareUploadCreateSuperFile 分片上传—合并分片文件, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareUploadCreateSuperFile(targetPath string, blockList ...string) (dataReadCloser io.ReadCloser, err error) {
	if !pcs.IsFile(targetPath) {
		return nil, fmt.Errorf("%s 遇到错误, 保存路径不可以覆盖目录", OperationUploadCreateSuperFile)
	}

	bl := &struct {
		BlockList []string `json:"block_list"`
	}{
		BlockList: blockList,
	}

	sendData, err := jsoniter.Marshal(bl)
	if err != nil {
		panic(OperationUploadCreateSuperFile + " 发生错误, " + err.Error())
	}

	pcs.setPCSURL("file", "createsuperfile", map[string]string{
		"path":  targetPath,
		"ondup": "overwrite",
	})

	// 表单上传
	mr := multipartreader.NewMultipartReader()
	mr.AddFormFeild("param", bytes.NewReader(sendData))

	resp, err := pcs.client.Req("POST", pcs.url.String(), mr, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 网络错误, %s", OperationUploadCreateSuperFile, err)
	}

	return resp.Body, nil
}

// PrepareCloudDlAddTask 添加离线下载任务, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlAddTask(sourceURL, savePath string) (dataReadCloser io.ReadCloser, err error) {
	pcs.setPCSURL2("services/cloud_dl", "add_task", map[string]string{
		"save_path":  savePath,
		"source_url": sourceURL,
		"timeout":    "86400",
	})

	resp, err := pcs.client.Req("POST", pcs.url.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 网络错误, %s", OperationCloudDlAddTask, err)
	}

	return resp.Body, nil
}

// PrepareCloudDlQueryTask 精确查询离线下载任务, 只返回服务器响应数据和错误信息,
// taskids 例子: 12123,234234,2344, 用逗号隔开多个 task_id
func (pcs *BaiduPCS) PrepareCloudDlQueryTask(taskIDs string) (dataReadCloser io.ReadCloser, err error) {
	pcs.setPCSURL2("services/cloud_dl", "query_task", map[string]string{
		"task_ids": taskIDs,
		"op_type":  "1",
	})

	resp, err := pcs.client.Req("POST", pcs.url.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 网络错误, %s", OperationCloudDlQueryTask, err)
	}

	return resp.Body, nil
}

// PrepareCloudDlListTask 查询离线下载任务列表, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlListTask() (dataReadCloser io.ReadCloser, err error) {
	pcs.setPCSURL2("services/cloud_dl", "list_task", map[string]string{
		"need_task_info": "1",
		"status":         "255",
		"start":          "0",
		"limit":          "1000",
	})

	resp, err := pcs.client.Req("POST", pcs.url.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 网络错误, %s", OperationCloudDlListTask, err)
	}

	return resp.Body, nil
}

func (pcs *BaiduPCS) prepareCloudDlCDTask(opreation, method string, taskID int64) (dataReadCloser io.ReadCloser, err error) {
	pcs.setPCSURL2("services/cloud_dl", method, map[string]string{
		"task_id": strconv.FormatInt(taskID, 10),
	})

	resp, err := pcs.client.Req("POST", pcs.url.String(), nil, nil)
	if err != nil {
		handleRespClose(resp)
		return nil, fmt.Errorf("%s, 网络错误, %s", opreation, err)
	}

	return resp.Body, nil
}

// PrepareCloudDlCancelTask 取消离线下载任务, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlCancelTask(taskID int64) (dataReadCloser io.ReadCloser, err error) {
	return pcs.prepareCloudDlCDTask(OperationCloudDlCancelTask, "cancel_task", taskID)
}

// PrepareCloudDlDeleteTask 取消离线下载任务, 只返回服务器响应数据和错误信息
func (pcs *BaiduPCS) PrepareCloudDlDeleteTask(taskID int64) (dataReadCloser io.ReadCloser, err error) {
	return pcs.prepareCloudDlCDTask(OperationCloudDlDeleteTask, "delete_task", taskID)
}
