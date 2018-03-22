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
func (pcs *BaiduPCS) RapidUpload(targetPath, contentMD5, sliceMD5, crc32 string, length int64) (err error) {
	dataReadCloser, err := pcs.PrepareRapidUpload(targetPath, contentMD5, sliceMD5, crc32, length)
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := NewErrorInfo(operationRapidUpload)

	d := jsoniter.NewDecoder(dataReadCloser)
	err = d.Decode(errInfo)
	if err != nil {
		return fmt.Errorf("%s, %s, %s", operationRapidUpload, StrJSONParseError, err)
	}

	switch errInfo.ErrCode {
	case 31079:
		// file md5 not found, you should use upload api to upload the whole file.
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}

	return nil
}

// Upload 上传单个文件
func (pcs *BaiduPCS) Upload(targetPath string, uploadFunc UploadFunc) (err error) {
	dataReadCloser, err := pcs.PrepareUpload(targetPath, uploadFunc)
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	// 数据处理
	jsonData := &struct {
		*PathJSON
		*ErrInfo
	}{
		ErrInfo: NewErrorInfo(operationUpload),
	}

	d := jsoniter.NewDecoder(dataReadCloser)

	err = d.Decode(jsonData)
	if err != nil {
		return fmt.Errorf("%s, %s, %s", operationUpload, StrJSONParseError, err)
	}

	if jsonData.ErrCode != 0 {
		return jsonData.ErrInfo
	}

	if jsonData.Path == "" {
		return fmt.Errorf("%s, unknown response data, file saved path not found", operationUpload)
	}

	return nil
}

// UploadTmpFile 分片上传—文件分片及上传
func (pcs *BaiduPCS) UploadTmpFile(uploadFunc UploadFunc) (md5 string, err error) {
	dataReadCloser, err := pcs.PrepareUploadTmpFile(uploadFunc)
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	// 数据处理
	jsonData := &struct {
		MD5 string `json:"md5"`
		*ErrInfo
	}{
		ErrInfo: NewErrorInfo(operationUploadTmpFile),
	}

	d := jsoniter.NewDecoder(dataReadCloser)

	err = d.Decode(jsonData)
	if err != nil {
		return "", fmt.Errorf("%s, %s, %s", operationUpload, StrJSONParseError, err)
	}

	if jsonData.ErrCode != 0 {
		return "", jsonData.ErrInfo
	}

	// 未找到md5
	if jsonData.MD5 == "" {
		return "", fmt.Errorf("%s, unknown response data, md5 not found", operationUpload)
	}

	return jsonData.MD5, nil
}

// UploadCreateSuperFile 分片上传—合并分片文件
func (pcs *BaiduPCS) UploadCreateSuperFile(targetPath string, blockList ...string) (err error) {
	dataReadCloser, err := pcs.PrepareUploadCreateSuperFile(targetPath, blockList...)
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := NewErrorInfo(operationUploadCreateSuperFile)

	d := jsoniter.NewDecoder(dataReadCloser)
	err = d.Decode(errInfo)
	if err != nil {
		return fmt.Errorf("%s, %s, %s", operationUploadCreateSuperFile, StrJSONParseError, err)
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}

	return nil
}
