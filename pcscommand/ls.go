package pcscommand

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/olekukonko/tablewriter"
	"os"
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

	for k := range files {
		if files[k].Isdir {
			files[k].Filename += "/"
		}
	}

	fmt.Printf("\n当前工作目录: %s\n----\n", path)

	tb := tablewriter.NewWriter(os.Stdout)
	tb.SetHeader([]string{"文件大小", "创建日期", "文件(目录)"})
	tb.SetBorder(false)
	tb.SetHeaderLine(false)
	tb.SetColumnSeparator("")

	tb.Append([]string{"", "", ""})

	var sizeStr string
	for _, file := range files {
		if file.Isdir {
			sizeStr = "-"
		} else {
			sizeStr = pcsutil.ConvertFileSize(file.Size)
		}

		tb.Append([]string{sizeStr, pcsutil.FormatTime(file.Ctime), file.Filename})
	}

	tb.Append([]string{"", "", ""})

	fN, dN := files.Count()
	tb.Append([]string{"总: " + pcsutil.ConvertFileSize(files.TotalSize()), "", fmt.Sprintf("文件总数: %d, 目录总数: %d", fN, dN)})

	tb.Render()

	if fN+dN >= 50 {
		fmt.Printf("\n当前工作目录: %s\n", path)
	}

	fmt.Printf("----\n")
	return
}
