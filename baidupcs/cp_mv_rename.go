package baidupcs

// Rename 重命名文件/目录
func (pcs *BaiduPCS) Rename(from, to string) (pcsError Error) {
	return pcs.cpmvOp(OperationRename, &CpMvJSON{
		From: from,
		To:   to,
	})
}

// Copy 批量拷贝文件/目录
func (pcs *BaiduPCS) Copy(cpmvJSON ...*CpMvJSON) (pcsError Error) {
	return pcs.cpmvOp(OperationCopy, cpmvJSON...)
}

// Move 批量移动文件/目录
func (pcs *BaiduPCS) Move(cpmvJSON ...*CpMvJSON) (pcsError Error) {
	return pcs.cpmvOp(OperationMove, cpmvJSON...)
}

func (pcs *BaiduPCS) cpmvOp(op string, cpmvJSON ...*CpMvJSON) (pcsError Error) {
	dataReadCloser, err := pcs.prepareCpMvOp(op, cpmvJSON...)
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := decodeJSONError(op, dataReadCloser)
	return errInfo
}
