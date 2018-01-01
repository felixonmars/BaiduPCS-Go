package baidupcscmd

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/util"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
)

const requiredSliceLen = 256 * pcsutil.KB // 256 KB

type upload struct {
	dir   string   // 目录路径
	files []string // 目录下所有子文件
}

type localPathInfo struct {
	path     string
	length   int64
	sliceMD5 string
	md5      string
	crc32    string
}

func (lp *localPathInfo) check() bool {
	info, err := os.Stat(lp.path)
	if err != nil {
		return false
	}
	lp.length = info.Size()
	return true
}

func (lp *localPathInfo) getSum() {
	file, err := os.Open(lp.path)
	if err != nil {
		fmt.Printf("open file %s error, %s\n", lp.path, err)
		return
	}

	bf := bufio.NewReader(file)

	// 获取前 256KB 文件切片的 md5
	buf := make([]byte, requiredSliceLen)
	file.ReadAt(buf[:], requiredSliceLen)
	lp.sliceMD5 = pcsutil.Md5Encrypt(buf[:])

	// 获取 文件 md5
	m := md5.New()
	bf.WriteTo(m)
	lp.md5 = fmt.Sprintf("%x", m.Sum(nil))

	// reset
	file, _ = os.Open(lp.path)
	bf.Reset(file)

	// 获取 文件 crc32
	c := crc32.NewIEEE()
	bf.WriteTo(c)
	lp.crc32 = fmt.Sprint(c.Sum32())
}

// RunUpload 执行文件上传
func RunUpload(localPaths []string, targetPath string) {
	targetPath, err := getAbsPath(targetPath)
	if err != nil {
		fmt.Printf("上传文件, 获取网盘路径 %s 错误, %s\n", targetPath, err)
	}

	switch len(localPaths) {
	case 0:
		fmt.Printf("本地路径为空\n")
		return
	}

	var _localPaths []upload
	for k := range localPaths {
		_paths, err := filepath.Glob(localPaths[k])
		if err != nil {
			fmt.Printf("上传文件, 匹配本地路径失败, %s\n", err)
			continue
		}

		for k2 := range _paths {
			_files, err := pcsutil.WalkDir(_paths[k2], "")
			if err != nil {
				fmt.Println(err)
				continue
			}

			_localPaths = append(_localPaths, upload{
				dir:   filepath.Dir(_paths[k2]),
				files: _files,
			})
		}
	}

	filesTotalNum := len(_localPaths)

	for ftN, uploadInfo := range _localPaths {

		filesNum := len(uploadInfo.files)

		for fN, file := range uploadInfo.files {
			fmt.Printf("[%d/%d - %d/%d] - [%s]: 任务开始\n", ftN+1, filesTotalNum, fN+1, filesNum, file)

			localPathInfo := localPathInfo{
				path: file,
			}

			if !localPathInfo.check() {
				fmt.Printf("文件不可读, 跳过...\n")
				continue
			}

			if localPathInfo.length < requiredSliceLen {
				continue
			}

			fmt.Printf("检测秒传中, 请稍候...\n")

			localPathInfo.getSum()

			err := info.RapidUpload(targetPath+"/"+strings.TrimLeft(localPathInfo.path, uploadInfo.dir), localPathInfo.md5, localPathInfo.sliceMD5, localPathInfo.crc32, localPathInfo.length)
			if err == nil {
				fmt.Printf("秒传成功\n")
				continue
			}
			fmt.Printf("秒传失败, 开始上传文件...\n")
		}
	}
}
