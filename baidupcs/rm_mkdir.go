package baidupcs

// Remove 批量删除文件/目录
func (pcs *BaiduPCS) Remove(paths ...string) (pcsError Error) {
	dataReadCloser, pcsError := pcs.PrepareRemove(paths...)
	if pcsError != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := decodeJSONError(OperationRemove, dataReadCloser)
	return errInfo
}

// Mkdir 创建目录
func (pcs *BaiduPCS) Mkdir(pcspath string) (pcsError Error) {
	dataReadCloser, pcsError := pcs.PrepareMkdir(pcspath)
	if pcsError != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := decodeJSONError(OperationMkdir, dataReadCloser)
	return errInfo
}
