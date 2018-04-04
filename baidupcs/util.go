package baidupcs

import (
	"fmt"
	"path"
)

// Isdir 检查路径在网盘中是否为目录
func (pcs *BaiduPCS) Isdir(pcspath string) (isdir bool, err error) {
	if path.Clean(pcspath) == "/" {
		return true, nil
	}

	f, err := pcs.FilesDirectoriesMeta(pcspath)
	if err != nil {
		return false, err
	}

	return f.Isdir, nil
}

func (pcs *BaiduPCS) checkIsdir(op string, targetPath string) error {
	// 检测文件是否存在于网盘路径
	// 很重要, 如果文件存在会直接覆盖!!! 即使是根目录!
	isdir, err := pcs.Isdir(targetPath)
	if err != nil {
		errInfo, ok := err.(*ErrInfo)
		if !ok {
			return err
		}

		// 忽略远程服务端返回的错误
		if errInfo.ErrType != ErrTypeRemoteError {
			return errInfo
		}
	}

	errInfo := NewErrorInfo(op)
	if isdir {
		errInfo.ErrType = ErrTypeOthers
		errInfo.Err = fmt.Errorf("保存路径不可以覆盖目录")
		return errInfo
	}
	return nil
}
