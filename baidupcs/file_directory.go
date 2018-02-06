package baidupcs

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
)

// FileDirectory 文件或目录的详细信息
type FileDirectory struct {
	FsID        int64  // fs_id
	Path        string // 路径
	Filename    string // 文件名 或 目录名
	Ctime       int64  // 创建日期
	Mtime       int64  // 修改日期
	MD5         string // md5 值
	Size        int64  // 文件大小 (目录为0)
	Isdir       bool   // 是否为目录
	Ifhassubdir bool   // 是否含有子目录 (只对目录有效)

	Parent   *FileDirectory    // 父目录信息
	Children FileDirectoryList // 子目录信息
}

// FileDirectoryList FileDirectory 的 指针数组
type FileDirectoryList []*FileDirectory

// FilesDirectoriesMeta 获取单个文件/目录的元信息
func (p *PCSApi) FilesDirectoriesMeta(path string) (data *FileDirectory, err error) {
	if path == "" {
		path = "/"
	}

	p.setApi("file", "meta", map[string]string{
		"path": path,
	})

	resp, err := p.client.Req("GET", p.url.String(), nil, nil)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	json, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		return
	}

	code, msg := CheckErr(json)
	if msg != "" {
		err = fmt.Errorf("获取单个文件/目录的元信息遇到错误, 路径: %s, 错误代码: %d, 消息: %s", path, code, msg)
		return
	}

	json = json.Get("list").GetIndex(0)

	data = &FileDirectory{
		FsID:        json.Get("fs_id").MustInt64(),
		Path:        json.Get("path").MustString(),
		Filename:    json.Get("server_filename").MustString(),
		Ctime:       json.Get("ctime").MustInt64(),
		Mtime:       json.Get("mtime").MustInt64(),
		MD5:         json.Get("md5").MustString(),
		Size:        json.Get("size").MustInt64(),
		Isdir:       pcsutil.IntToBool(json.Get("isdir").MustInt()),
		Ifhassubdir: pcsutil.IntToBool(json.Get("ifhassubdir").MustInt()),
	}

	return
}

// FilesDirectoriesList 获取目录下的文件和目录列表, 可选是否递归
func (p *PCSApi) FilesDirectoriesList(path string, recurse bool) (data FileDirectoryList, err error) {
	if path == "" {
		path = "/"
	}

	p.setApi("file", "list", map[string]string{
		"path":  path,
		"by":    "name",
		"order": "asc", // 升序
		"limit": "0-2147483647",
	})

	resp, err := p.client.Req("GET", p.url.String(), nil, nil)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	json, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		return
	}

	code, msg := CheckErr(json)
	if msg != "" {
		return nil, fmt.Errorf("获取目录下的文件列表遇到错误, 路径: %s, 错误代码: %d, 消息: %s", path, code, msg)
	}

	json = json.Get("list")

	for i := 0; ; i++ {
		index := json.GetIndex(i)
		fsID := index.Get("fs_id").MustInt64()
		if fsID == 0 {
			break
		}

		sub := &FileDirectory{
			FsID:     fsID,
			Path:     index.Get("path").MustString(),
			Filename: index.Get("server_filename").MustString(),
			Ctime:    index.Get("ctime").MustInt64(),
			Mtime:    index.Get("mtime").MustInt64(),
			MD5:      index.Get("md5").MustString(),
			Size:     index.Get("size").MustInt64(),
			Isdir:    pcsutil.IntToBool(index.Get("isdir").MustInt()),
		}

		// 递归获取子目录信息
		if recurse && sub.Isdir {
			sub.Children, err = p.FilesDirectoriesList(sub.Path, true)
			if err != nil {
				pcsverbose.Verboseln(err)
			}
		}

		data = append(data, sub)
	}
	return
}

func (f *FileDirectory) String() string {
	if f.Isdir {
		return fmt.Sprintf("类型: 目录 \n目录名称: %s \n目录路径: %s \nfs_id: %d \n创建日期: %s \n修改日期: %s \n是否含有子目录: %t\n",
			f.Filename,
			f.Path,
			f.FsID,
			pcsutil.FormatTime(f.Ctime),
			pcsutil.FormatTime(f.Mtime),
			f.Ifhassubdir,
		)
	}

	return fmt.Sprintf("类型: 文件 \n文件名: %s \n文件路径: %s \n文件大小: %d, (%s) \nmd5: %s \nfs_id: %d \n创建日期: %s \n修改日期: %s \n",
		f.Filename,
		f.Path,
		f.Size, pcsutil.ConvertFileSize(f.Size),
		f.MD5,
		f.FsID,
		pcsutil.FormatTime(f.Ctime),
		pcsutil.FormatTime(f.Mtime),
	)
}

// TotalSize 获取目录下文件的总大小
func (fl FileDirectoryList) TotalSize() int64 {
	var size int64
	for k := range fl {
		if fl[k] == nil {
			continue
		}

		size += fl[k].Size

		// 递归获取
		if fl[k].Children != nil {
			size += fl[k].Children.TotalSize()
		}
	}
	return size
}

// Count 获取文件总数和目录总数
func (fl FileDirectoryList) Count() (fileN, directoryN int64) {
	for k := range fl {
		if fl[k] == nil {
			continue
		}

		if fl[k].Isdir {
			directoryN++
		} else {
			fileN++
		}

		// 递归获取
		if fl[k].Children != nil {
			fN, dN := fl[k].Children.Count()
			fileN += fN
			directoryN += dN
		}
	}
	return
}

// AllFilePaths 返回所有的网盘路径, 包括子目录
func (fl FileDirectoryList) AllFilePaths() (pcspaths []string) {
	fN, dN := fl.Count()
	pcspaths = make([]string, fN+dN)
	for k := range fl {
		if fl[k] == nil {
			continue
		}

		pcspaths = append(pcspaths, fl[k].Path)

		if fl[k].Children != nil {
			pcspaths = append(pcspaths, fl[k].Children.AllFilePaths()...)
		}
	}
	return
}
