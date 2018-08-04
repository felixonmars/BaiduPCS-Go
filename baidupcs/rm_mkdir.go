package baidupcs

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
)

// Remove 批量删除文件/目录
func (pcs *BaiduPCS) Remove(paths ...string) (pcsError pcserror.Error) {
	dataReadCloser, pcsError := pcs.PrepareRemove(paths...)
	if pcsError != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := pcserror.DecodePCSJSONError(OperationRemove, dataReadCloser)
	return errInfo
}

// Mkdir 创建目录
func (pcs *BaiduPCS) Mkdir(pcspath string) (pcsError pcserror.Error) {
	dataReadCloser, pcsError := pcs.PrepareMkdir(pcspath)
	if pcsError != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := pcserror.DecodePCSJSONError(OperationMkdir, dataReadCloser)
	return errInfo
}
