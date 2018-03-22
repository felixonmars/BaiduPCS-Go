package baidupcs

import (
	"fmt"
	"github.com/json-iterator/go"
)

// Remove 批量删除文件/目录
func (pcs *BaiduPCS) Remove(paths ...string) (err error) {
	dataReadCloser, err := pcs.PrepareRemove(paths...)
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := NewErrorInfo(operationRemove)

	d := jsoniter.NewDecoder(dataReadCloser)
	err = d.Decode(errInfo)
	if err != nil {
		return fmt.Errorf("%s, json 数据解析失败, %s", operationRemove, err)
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}

	return nil
}

// Mkdir 创建目录
func (pcs *BaiduPCS) Mkdir(pcspath string) (err error) {
	dataReadCloser, err := pcs.PrepareMkdir(pcspath)
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := NewErrorInfo(operationMkdir)

	d := jsoniter.NewDecoder(dataReadCloser)
	err = d.Decode(errInfo)
	if err != nil {
		return fmt.Errorf("%s, json 数据解析失败, %s", operationMkdir, err)
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}
	return
}
