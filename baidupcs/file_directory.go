package baidupcs

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/iikira/BaiduPCS-Go/downloader"
	"github.com/iikira/BaiduPCS-Go/util"
)

// FileDirectory 文件或目录的详细信息
type FileDirectory struct {
	FsID        int64
	Path        string
	Filename    string
	Ctime       int64
	MD5         string
	Size        int64
	Isdir       bool
	Ifhassubdir bool
}

// FileDirectoryList FileDirectory 的 数组
type FileDirectoryList []FileDirectory

// FilesDirectoriesMeta 获取单个文件/目录的元信息
//
// 可用信息: 是否是目录isdir 是否含有子目录ifhassubdir 修改时间mtime 文件大小size
func (p PCSApi) FilesDirectoriesMeta(path string) (data FileDirectory, err error) {
	p.addItem("file", "meta", map[string]string{
		"path": path,
	})

	h := downloader.NewHTTPClient()
	body, err := h.Fetch("GET", p.url.String(), nil, map[string]string{
		"Cookie": "BDUSS=" + p.bduss,
	})
	if err != nil {
		return
	}

	json, err := simplejson.NewJson(body)

	code, err := checkErr(json)
	if err != nil {
		err = fmt.Errorf("获取单个文件/目录的元信息遇到错误, 错误代码: %d, 消息: %s", code, err)
		return
	}

	json = json.Get("list").GetIndex(0)

	data = FileDirectory{
		FsID:        json.Get("fs_id").MustInt64(),
		Path:        json.Get("path").MustString(),
		Filename:    json.Get("server_filename").MustString(),
		Ctime:       json.Get("ctime").MustInt64(),
		MD5:         json.Get("md5").MustString(),
		Size:        json.Get("size").MustInt64(),
		Isdir:       pcsutil.IntToBool(json.Get("isdir").MustInt()),
		Ifhassubdir: pcsutil.IntToBool(json.Get("ifhassubdir").MustInt()),
	}

	return
}

// FileList 获取目录下的文件列表
func (p PCSApi) FileList(path string) (data FileDirectoryList, err error) {
	p.addItem("file", "list", map[string]string{
		"path":  path,
		"by":    "name",
		"order": "asc",
		"limit": "0-2147483647",
	})

	h := downloader.NewHTTPClient()
	body, err := h.Fetch("GET", p.url.String(), nil, map[string]string{
		"Cookie": "BDUSS=" + p.bduss,
	})
	if err != nil {
		return
	}

	json, err := simplejson.NewJson(body)
	if err != nil {
		return
	}

	code, err := checkErr(json)
	if err != nil {
		return nil, fmt.Errorf("获取目录下的文件列表遇到错误, 错误代码: %d, 消息: %s", code, err)
	}

	json = json.Get("list")

	for i := 0; ; i++ {
		index := json.GetIndex(i)
		fsID := index.Get("fs_id").MustInt64()
		if fsID == 0 {
			break
		}
		data = append(data, FileDirectory{
			FsID:     fsID,
			Path:     index.Get("path").MustString(),
			Filename: index.Get("server_filename").MustString(),
			Isdir:    pcsutil.IntToBool(index.Get("isdir").MustInt()),
			Ctime:    index.Get("ctime").MustInt64(),
			Size:     index.Get("size").MustInt64(),
		})
	}
	return
}

func (f FileDirectory) String() string {
	if f.Isdir {
		return fmt.Sprintf("类型: 目录 \n目录名称: %s \n目录路径: %s \nfs_id: %d \n创建时间: %s \n是否含有子目录: %t\n",
			f.Filename,
			f.Path,
			f.FsID,
			pcsutil.FormatTime(f.Ctime),
			f.Ifhassubdir,
		)
	}

	return fmt.Sprintf("类型: 文件 \n文件名: %s \n文件路径: %s \n文件大小: %d \nmd5: %s \nfs_id: %d \n创建时间: %s \n",
		f.Filename,
		f.Path,
		f.Size,
		f.MD5,
		f.FsID,
		pcsutil.FormatTime(f.Ctime),
	)
}

// TotalSize 获取总文件大小
func (f *FileDirectoryList) TotalSize() int64 {
	var size int64
	for k := range *f {
		size += (*f)[k].Size
	}
	return size
}

// Count 获取文件总数和目录总数
func (f *FileDirectoryList) Count() (fileN int64, directoryN int64) {
	for k := range *f {
		if (*f)[k].Isdir {
			directoryN++
		} else {
			fileN++
		}
	}
	return
}
