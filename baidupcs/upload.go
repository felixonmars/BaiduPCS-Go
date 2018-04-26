package baidupcs

import (
	"fmt"
	"github.com/json-iterator/go"
	"net/http"
	"net/http/cookiejar"
)

// UploadFunc 上传文件处理函数
type UploadFunc func(uploadURL string, jar *cookiejar.Jar) (resp *http.Response, err error)

// RapidUpload 秒传文件
func (pcs *BaiduPCS) RapidUpload(targetPath, contentMD5, sliceMD5, crc32 string, length int64) (pcsError Error) {
	dataReadCloser, pcsError := pcs.PrepareRapidUpload(targetPath, contentMD5, sliceMD5, crc32, length)
	if pcsError != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := decodeJSONError(OperationUpload, dataReadCloser)
	if errInfo == nil {
		return nil
	}

	switch errInfo.ErrorCode() {
	case 31079:
		// file md5 not found, you should use upload api to upload the whole file.
	}

	return errInfo
}

// Upload 上传单个文件
func (pcs *BaiduPCS) Upload(targetPath string, uploadFunc UploadFunc) (pcsError Error) {
	dataReadCloser, pcsError := pcs.PrepareUpload(targetPath, uploadFunc)
	if pcsError != nil {
		return
	}

	defer dataReadCloser.Close()

	// 数据处理
	jsonData := &struct {
		*PathJSON
		*ErrInfo
	}{
		ErrInfo: NewErrorInfo(OperationUpload),
	}

	d := jsoniter.NewDecoder(dataReadCloser)

	err := d.Decode(jsonData)
	if err != nil {
		jsonData.ErrInfo.jsonError(err)
		return jsonData.ErrInfo
	}

	if jsonData.ErrCode != 0 {
		return jsonData.ErrInfo
	}

	if jsonData.Path == "" {
		jsonData.ErrInfo.errType = ErrTypeInternalError
		jsonData.ErrInfo.err = fmt.Errorf("unknown response data, file saved path not found")
		return jsonData.ErrInfo
	}

	return nil
}

// UploadTmpFile 分片上传—文件分片及上传
func (pcs *BaiduPCS) UploadTmpFile(uploadFunc UploadFunc) (md5 string, pcsError Error) {
	dataReadCloser, pcsError := pcs.PrepareUploadTmpFile(uploadFunc)
	if pcsError != nil {
		return "", pcsError
	}

	defer dataReadCloser.Close()

	// 数据处理
	jsonData := &struct {
		MD5 string `json:"md5"`
		*ErrInfo
	}{
		ErrInfo: NewErrorInfo(OperationUploadTmpFile),
	}

	d := jsoniter.NewDecoder(dataReadCloser)

	err := d.Decode(jsonData)
	if err != nil {
		jsonData.ErrInfo.jsonError(err)
		return "", jsonData.ErrInfo
	}

	if jsonData.ErrCode != 0 {
		return "", jsonData.ErrInfo
	}

	// 未找到md5
	if jsonData.MD5 == "" {
		jsonData.ErrInfo.errType = ErrTypeInternalError
		jsonData.ErrInfo.err = fmt.Errorf("unknown response data, md5 not found, error: %s", err)
		return "", jsonData.ErrInfo
	}

	return jsonData.MD5, nil
}

// UploadCreateSuperFile 分片上传—合并分片文件
func (pcs *BaiduPCS) UploadCreateSuperFile(targetPath string, blockList ...string) (pcsError Error) {
	dataReadCloser, pcsError := pcs.PrepareUploadCreateSuperFile(targetPath, blockList...)
	if pcsError != nil {
		return pcsError
	}

	defer dataReadCloser.Close()

	errInfo := decodeJSONError(OperationUploadCreateSuperFile, dataReadCloser)
	return errInfo
}
