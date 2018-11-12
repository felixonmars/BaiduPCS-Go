package baidupcs

import (
	"errors"
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
	"github.com/iikira/BaiduPCS-Go/pcsutil/escaper"
	"mime"
	"net/url"
	"path"
	"strings"
)

const (
	// ShellPatternCharacters 通配符字符串
	ShellPatternCharacters = "*?[]"
)

var (
	// ErrFixMD5Isdir 目录不需要修复md5
	ErrFixMD5Isdir = errors.New("directory not support fix md5")
	// ErrFixMD5Failed 修复MD5失败, 可能服务器未刷新
	ErrFixMD5Failed = errors.New("fix md5 failed")
	// ErrFixMD5FileInfoNil 文件信息对象为空
	ErrFixMD5FileInfoNil = errors.New("file info is nil")
	// ErrMatchPathByShellPatternNotAbsPath 不是绝对路径
	ErrMatchPathByShellPatternNotAbsPath = errors.New("not absolute path")
)

// FixMD5ByFileInfo 尝试修复文件的md5, 通过文件信息对象
func (pcs *BaiduPCS) FixMD5ByFileInfo(finfo *FileDirectory) (pcsError pcserror.Error) {
	errInfo := pcserror.NewPCSErrorInfo(OperationFixMD5)
	errInfo.ErrType = pcserror.ErrTypeOthers
	if finfo == nil {
		errInfo.Err = ErrFixMD5FileInfoNil
		return errInfo
	}

	// 忽略目录
	if finfo.Isdir {
		errInfo.Err = ErrFixMD5Isdir
		return errInfo
	}

	if len(finfo.BlockList) == 1 && strings.Compare(finfo.BlockList[0], finfo.MD5) == 0 {
		// 不需要修复
		return nil
	}

	info, pcsError := pcs.LocateDownloadWithUserAgent(finfo.Path, NetdiskUA)
	if pcsError != nil {
		return
	}

	if info.URLs == nil {
		pcsError.SetJSONError(ErrNilJSONValue)
		return
	}

	for k, u := range info.URLs {
		resp, err := pcs.client.Req("HEAD", u.URL, nil, netdiskUAHeader)
		if resp != nil {
			resp.Body.Close()
		}
		if err != nil {
			baiduPCSVerbose.Warnf("[%d] request link error, link: %s, err: %s\n", k, u.URL, err)
			continue
		}

		// 检测响应状态码
		if resp.StatusCode/100 != 2 {
			baiduPCSVerbose.Warnf("[%d] link: %s, http status code: %d\n", k, u.URL, resp.StatusCode)
			continue
		}

		// 检测大小
		if finfo.Size != resp.ContentLength {
			baiduPCSVerbose.Warnf("[%d] file size match failed, origin size: %d, resp Content-Length: %d, link: %s\n", k, finfo.Size, resp.ContentLength, u.URL)
			continue
		}

		// 检测文件名是否对应
		_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
		if err != nil {
			baiduPCSVerbose.Warnf("[%d] ParseMediaType error, link: %s, err: %s\n", k, u.URL, err)
			continue
		}

		filename, err := url.QueryUnescape(params["filename"])
		if err != nil {
			baiduPCSVerbose.Warnf("[%d] unescape filename error, link: %s, err: %s\n", k, u.URL, err)
			continue
		}
		if strings.Compare(finfo.Filename, filename) != 0 {
			baiduPCSVerbose.Warnf("[%d] filename match failed, origin filename: %s, resp filename: %s, link: %s\n", k, finfo.Filename, filename, u.URL)
			continue
		}

		// 检测是否存在MD5
		md5Str := resp.Header.Get("Content-MD5")
		if md5Str == "" { // 未找到md5值, 可能是服务器未刷新
			errInfo.Err = ErrFixMD5Failed
			return errInfo
		}

		// 检测是否存在crc32 值, 一般都会存在的
		crc32Str := resp.Header.Get("x-bs-meta-crc32")
		if crc32Str == "" || crc32Str == "0" {
			errInfo.Err = ErrFixMD5Failed
			return errInfo
		}

		// 开始修复
		return pcs.RapidUploadNoCheckDir(finfo.Path, md5Str, DefaultSliceMD5, crc32Str, finfo.Size)
	}

	errInfo.Err = ErrFixMD5Failed
	return errInfo
}

// FixMD5 尝试修复文件的md5
func (pcs *BaiduPCS) FixMD5(pcspath string) (pcsError pcserror.Error) {
	finfo, pcsError := pcs.FilesDirectoriesMeta(pcspath)
	if pcsError != nil {
		return
	}

	return pcs.FixMD5ByFileInfo(finfo)
}

func (pcs *BaiduPCS) recurseMatchPathByShellPattern(index int, patternSlice *[]string, ps *[]string, pcspaths *[]string) {
	if index == len(*patternSlice) {
		*pcspaths = append(*pcspaths, strings.Join(*ps, PathSeparator))
		return
	}

	if !strings.ContainsAny((*patternSlice)[index], ShellPatternCharacters) {
		(*ps)[index] = (*patternSlice)[index]
		pcs.recurseMatchPathByShellPattern(index+1, patternSlice, ps, pcspaths)
		return
	}

	fds, pcsError := pcs.FilesDirectoriesList(strings.Join((*ps)[:index], PathSeparator), DefaultOrderOptions)
	if pcsError != nil {
		panic(pcsError) // 抛出异常
	}

	for k := range fds {
		if matched, _ := path.Match((*patternSlice)[index], fds[k].Filename); matched {
			(*ps)[index] = fds[k].Filename
			pcs.recurseMatchPathByShellPattern(index+1, patternSlice, ps, pcspaths)
		}
	}
	return
}

// MatchPathByShellPattern 通配符匹配文件路径, pattern 为绝对路径
func (pcs *BaiduPCS) MatchPathByShellPattern(pattern string) (pcspaths []string, pcsError pcserror.Error) {
	errInfo := pcserror.NewPCSErrorInfo(OperrationMatchPathByShellPattern)
	errInfo.ErrType = pcserror.ErrTypeOthers

	patternSlice := strings.Split(escaper.Escape(path.Clean(pattern), []rune{'['}), PathSeparator) // 转义中括号
	if patternSlice[0] != "" {
		errInfo.Err = ErrMatchPathByShellPatternNotAbsPath
		return nil, errInfo
	}

	ps := make([]string, len(patternSlice))
	defer func() { // 捕获异常
		if err := recover(); err != nil {
			pcspaths = nil
			pcsError = err.(pcserror.Error)
		}
	}()
	pcs.recurseMatchPathByShellPattern(1, &patternSlice, &ps, &pcspaths)
	return pcspaths, nil
}
