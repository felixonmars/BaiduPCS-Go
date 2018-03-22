package baidupcs

import (
	"fmt"
	"github.com/json-iterator/go"
)

// Rename 重命名文件/目录
func (pcs *BaiduPCS) Rename(from, to string) (err error) {
	return pcs.cpmvOp(operationRename, &CpMvJSON{
		From: from,
		To:   to,
	})
}

// Copy 批量拷贝文件/目录
func (pcs *BaiduPCS) Copy(cpmvJSON ...*CpMvJSON) (err error) {
	return pcs.cpmvOp(operationCopy, cpmvJSON...)
}

// Move 批量移动文件/目录
func (pcs *BaiduPCS) Move(cpmvJSON ...*CpMvJSON) (err error) {
	return pcs.cpmvOp(operationMove, cpmvJSON...)
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
		return fmt.Errorf("%s, %s, %s", op, StrJSONParseError, err)
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}

	return nil
}
