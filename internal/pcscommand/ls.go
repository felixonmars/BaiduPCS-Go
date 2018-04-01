package pcscommand

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
)

// RunLs 执行列目录
func RunLs(path string) {
	path, err := getAbsPath(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	files, err := info.FilesDirectoriesList(path, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("\n当前目录: %s\n----\n", path)

	tb := pcstable.NewTable(os.Stdout)
	tb.SetHeader([]string{"#", "文件大小", "创建日期", "文件(目录)"})

	tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})

	for k, file := range files {
		if file.Isdir {
			tb.Append([]string{strconv.Itoa(k), "-", pcsutil.FormatTime(file.Ctime), file.Filename + "/"})
			continue
		}

		tb.Append([]string{strconv.Itoa(k), pcsutil.ConvertFileSize(file.Size), pcsutil.FormatTime(file.Ctime), file.Filename})
	}

	fN, dN := files.Count()
	tb.Append([]string{"", "总: " + pcsutil.ConvertFileSize(files.TotalSize()), "", fmt.Sprintf("文件总数: %d, 目录总数: %d", fN, dN)})

	tb.Render()

	if fN+dN >= 50 {
		fmt.Printf("\n当前目录: %s\n", path)
	}

	fmt.Printf("----\n")
	return
}
