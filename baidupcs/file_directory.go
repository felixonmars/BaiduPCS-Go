package baidupcs

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"github.com/iikira/BaiduPCS-Go/pcsutil/pcstime"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/json-iterator/go"
	"github.com/olekukonko/tablewriter"
	"strconv"
	"strings"
)

// HandleFileDirectoryFunc 处理文件或目录的元信息
type HandleFileDirectoryFunc func(depth int, fd *FileDirectory)

// FileDirectory 文件或目录的元信息
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

// fdJSON 用于解析远程JSON数据
type fdJSON struct {
	FsID           int64  `json:"fs_id"`           // fs_id
	Path           string `json:"path"`            // 路径
	Filename       string `json:"server_filename"` // 文件名 或 目录名
	Ctime          int64  `json:"ctime"`           // 创建日期
	Mtime          int64  `json:"mtime"`           // 修改日期
	MD5            string `json:"md5"`             // md5 值
	Size           int64  `json:"size"`            // 文件大小 (目录为0)
	IsdirInt       int    `json:"isdir"`
	IfhassubdirInt int    `json:"ifhassubdir"`
}

// convert 将解析的远程JSON数据, 转换为 *FileDirectory
func (fj *fdJSON) convert() *FileDirectory {
	return &FileDirectory{
		FsID:        fj.FsID,
		Path:        fj.Path,
		Filename:    fj.Filename,
		Ctime:       fj.Ctime,
		Mtime:       fj.Mtime,
		MD5:         fj.MD5,
		Size:        fj.Size,
		Isdir:       converter.IntToBool(fj.IsdirInt),
		Ifhassubdir: converter.IntToBool(fj.IfhassubdirInt),
	}
}

type fdData struct {
	*ErrInfo
	List []*fdJSON `json:"list"`
}

// FilesDirectoriesMeta 获取单个文件/目录的元信息
func (pcs *BaiduPCS) FilesDirectoriesMeta(path string) (data *FileDirectory, pcsError Error) {
	if path == "" {
		path = "/"
	}

	fds, err := pcs.FilesDirectoriesBatchMeta(path)
	if err != nil {
		return nil, err
	}

	// 返回了多条元信息
	if len(fds) != 1 {
		return nil, &ErrInfo{
			operation: OperationFilesDirectoriesMeta,
			errType:   ErrTypeOthers,
			err:       fmt.Errorf("未知返回数据"),
		}
	}

	return fds[0], nil
}

// FilesDirectoriesBatchMeta 获取多个文件/目录的元信息
func (pcs *BaiduPCS) FilesDirectoriesBatchMeta(paths ...string) (data FileDirectoryList, pcsError Error) {
	dataReadCloser, pcsError := pcs.PrepareFilesDirectoriesBatchMeta(paths...)
	if pcsError != nil {
		return nil, pcsError
	}

	defer dataReadCloser.Close()

	errInfo := NewErrorInfo(OperationFilesDirectoriesMeta)
	// 服务器返回数据进行处理
	jsonData := &fdData{
		ErrInfo: errInfo,
	}

	d := jsoniter.NewDecoder(dataReadCloser)
	err := d.Decode(jsonData)
	if err != nil {
		errInfo.jsonError(err)
		return nil, errInfo
	}

	// 错误处理
	errCode, _ := jsonData.ErrInfo.FindErr()
	if errCode != 0 {
		return nil, jsonData.ErrInfo
	}

	// 结果处理
	data = make(FileDirectoryList, len(jsonData.List))
	for k := range jsonData.List {
		data[k] = jsonData.List[k].convert()
	}

	return
}

// FilesDirectoriesList 获取目录下的文件和目录列表
func (pcs *BaiduPCS) FilesDirectoriesList(path string) (data FileDirectoryList, pcsError Error) {
	dataReadCloser, pcsError := pcs.PrepareFilesDirectoriesList(path)
	if pcsError != nil {
		return nil, pcsError
	}

	defer dataReadCloser.Close()

	jsonData := &fdData{
		ErrInfo: NewErrorInfo(OperationFilesDirectoriesList),
	}

	d := jsoniter.NewDecoder(dataReadCloser)
	err := d.Decode(jsonData)
	if err != nil {
		jsonData.ErrInfo.jsonError(err)
		return nil, jsonData.ErrInfo
	}

	// 错误处理
	errCode, _ := jsonData.ErrInfo.FindErr()
	if errCode != 0 {
		return nil, jsonData.ErrInfo
	}

	// 可能是一个文件
	if len(jsonData.List) == 0 {
		var fd *FileDirectory
		fd, pcsError = pcs.FilesDirectoriesMeta(path)
		if pcsError != nil {
			return
		}

		if fd.Isdir {
			return
		}

		return FileDirectoryList{fd}, nil
	}

	data = make(FileDirectoryList, len(jsonData.List))
	for k := range jsonData.List {
		data[k] = jsonData.List[k].convert()
	}
	return
}

func (pcs *BaiduPCS) recurseList(path string, depth int, handleFileDirectoryFunc HandleFileDirectoryFunc) (data FileDirectoryList, pcsError Error) {
	fdl, pcsError := pcs.FilesDirectoriesList(path)
	if pcsError != nil {
		return nil, pcsError
	}

	for k := range fdl {
		handleFileDirectoryFunc(depth+1, fdl[k])
		if !fdl[k].Isdir {
			continue
		}

		fdl[k].Children, pcsError = pcs.recurseList(fdl[k].Path, depth+1, handleFileDirectoryFunc)
		if pcsError != nil {
			pcsverbose.Verboseln(pcsError)
		}
	}

	return fdl, nil
}

// FilesDirectoriesRecurseList 递归获取目录下的文件和目录列表
func (pcs *BaiduPCS) FilesDirectoriesRecurseList(path string, handleFileDirectoryFunc HandleFileDirectoryFunc) (data FileDirectoryList, pcsError Error) {
	return pcs.recurseList(path, 0, handleFileDirectoryFunc)
}

func (f *FileDirectory) String() string {
	builder := &strings.Builder{}
	tb := pcstable.NewTable(builder)
	tb.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})

	if f.Isdir {
		tb.AppendBulk([][]string{
			[]string{"类型", "目录"},
			[]string{"目录路径", f.Path},
			[]string{"目录名称", f.Filename},
		})
	} else {
		tb.AppendBulk([][]string{
			[]string{"类型", "文件"},
			[]string{"文件路径", f.Path},
			[]string{"文件名称", f.Filename},
			[]string{"文件大小", strconv.FormatInt(f.Size, 10) + ", " + converter.ConvertFileSize(f.Size)},
			[]string{"md5 (截图请打码)", f.MD5},
		})
	}

	tb.Append([]string{"fs_id", strconv.FormatInt(f.FsID, 10)})
	tb.AppendBulk([][]string{
		[]string{"创建日期", pcstime.FormatTime(f.Ctime)},
		[]string{"修改日期", pcstime.FormatTime(f.Mtime)},
	})

	if f.Ifhassubdir {
		tb.Append([]string{"是否含有子目录", "true"})
	}

	tb.Render()
	return builder.String()
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
	pcspaths = make([]string, 0, fN+dN)
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
