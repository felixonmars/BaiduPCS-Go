package baidupcs

import (
	"errors"
	"github.com/json-iterator/go"
	"io"
	"path"
)

// Isdir 检查路径在网盘中是否为目录
func (pcs *BaiduPCS) Isdir(pcspath string) (isdir bool, pcsError Error) {
	if path.Clean(pcspath) == "/" {
		return true, nil
	}

	f, pcsError := pcs.FilesDirectoriesMeta(pcspath)
	if pcsError != nil {
		return false, pcsError
	}

	return f.Isdir, nil
}

func (pcs *BaiduPCS) checkIsdir(op string, targetPath string) Error {
	// 检测文件是否存在于网盘路径
	// 很重要, 如果文件存在会直接覆盖!!! 即使是根目录!
	isdir, pcsError := pcs.Isdir(targetPath)
	if pcsError != nil {
		// 忽略远程服务端返回的错误
		if pcsError.ErrorType() != ErrTypeRemoteError {
			return pcsError
		}
	}

	errInfo := NewErrorInfo(op)
	if isdir {
		errInfo.errType = ErrTypeOthers
		errInfo.err = errors.New("保存路径不可以覆盖目录")
		return errInfo
	}
	return nil
}

// decodeJSONError 解析json中的远端服务器返回的错误
func decodeJSONError(op string, data io.Reader) Error {
	errInfo := NewErrorInfo(op)

	d := jsoniter.NewDecoder(data)
	err := d.Decode(errInfo)
	if err != nil {
		errInfo.jsonError(err)
		return errInfo
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}

	return nil
}
