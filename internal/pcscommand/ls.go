package pcscommand

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"github.com/iikira/BaiduPCS-Go/pcsutil/pcstime"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
)

// LsOptions 列目录可选项
type LsOptions struct {
	Total bool
}

// RunLs 执行列目录
func RunLs(path string, lsOptions *LsOptions, orderOptions *baidupcs.OrderOptions) {
	path, err := getAbsPath(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	files, err := GetBaiduPCS().FilesDirectoriesList(path, orderOptions)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("\n当前目录: %s\n----\n", path)

	tb := pcstable.NewTable(os.Stdout)

	if lsOptions == nil {
		lsOptions = &LsOptions{}
	}

	var (
		fN, dN int64
	)

	if lsOptions.Total {
		tb.SetHeader([]string{"#", "fs_id", "文件大小", "创建日期", "修改日期", "md5(截图请打码)", "文件(目录)"})
		tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
		for k, file := range files {
			if file.Isdir {
				tb.Append([]string{strconv.Itoa(k), strconv.FormatInt(file.FsID, 10), "-", pcstime.FormatTime(file.Ctime), pcstime.FormatTime(file.Mtime), file.MD5, file.Filename + "/"})
				continue
			}

			tb.Append([]string{strconv.Itoa(k), strconv.FormatInt(file.FsID, 10), converter.ConvertFileSize(file.Size, 2), pcstime.FormatTime(file.Ctime), pcstime.FormatTime(file.Mtime), file.MD5, file.Filename})
		}
		fN, dN := files.Count()
		tb.Append([]string{"", "", "总: " + converter.ConvertFileSize(files.TotalSize(), 2), "", "", "", fmt.Sprintf("文件总数: %d, 目录总数: %d", fN, dN)})
	} else {
		tb.SetHeader([]string{"#", "文件大小", "修改日期", "文件(目录)"})
		tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
		for k, file := range files {
			if file.Isdir {
				tb.Append([]string{strconv.Itoa(k), "-", pcstime.FormatTime(file.Mtime), file.Filename + "/"})
				continue
			}

			tb.Append([]string{strconv.Itoa(k), converter.ConvertFileSize(file.Size, 2), pcstime.FormatTime(file.Mtime), file.Filename})
		}
		fN, dN = files.Count()
		tb.Append([]string{"", "总: " + converter.ConvertFileSize(files.TotalSize(), 2), "", fmt.Sprintf("文件总数: %d, 目录总数: %d", fN, dN)})
	}

	tb.Render()

	if fN+dN >= 50 {
		fmt.Printf("\n当前目录: %s\n", path)
	}

	fmt.Printf("----\n")
	return
}
