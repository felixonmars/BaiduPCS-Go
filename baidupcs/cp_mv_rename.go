package baidupcs

import (
	"github.com/json-iterator/go"
)

// Rename 重命名文件/目录
func (pcs *BaiduPCS) Rename(from, to string) (err error) {
	return pcs.cpmvOp(OperationRename, &CpMvJSON{
		From: from,
		To:   to,
	})
}

// Copy 批量拷贝文件/目录
func (pcs *BaiduPCS) Copy(cpmvJSON ...*CpMvJSON) (err error) {
	return pcs.cpmvOp(OperationCopy, cpmvJSON...)
}

// Move 批量移动文件/目录
func (pcs *BaiduPCS) Move(cpmvJSON ...*CpMvJSON) (err error) {
	return pcs.cpmvOp(OperationMove, cpmvJSON...)
}

func (pcs *BaiduPCS) cpmvOp(op string, cpmvJSON ...*CpMvJSON) (err error) {
	dataReadCloser, err := pcs.prepareCpMvOp(op, cpmvJSON...)
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := NewErrorInfo(op)

	d := jsoniter.NewDecoder(dataReadCloser)
	err = d.Decode(errInfo)
	if err != nil {
		errInfo.jsonError(err)
		return errInfo
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}

	return nil
}
