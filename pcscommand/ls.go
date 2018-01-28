package pcscommand

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"os"
	"text/template"
)

// RunLs 执行列目录
func RunLs(path string) {
	path, err := getAbsPath(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	files, err := info.FileList(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(files) == 0 {
		RunGetMeta(path)
		return
	}

	for k := range files {
		if files[k].Isdir {
			files[k].Path += "/"
		}
	}

	tmpl, err := template.New("ls").Funcs(
		template.FuncMap{
			"convertFileSize": func(size int64) string {
				res := pcsutil.ConvertFileSize(size)
				if res == "0" {
					return "-       "
				}
				return res
			},
			"timeFmt": pcsutil.FormatTime,
			"totalSize": func() string {
				return pcsutil.ConvertFileSize(files.TotalSize())
			},
			"fdCount": func() string {
				fN, dN := files.Count()
				return fmt.Sprintf("文件总数: %d,\t目录总数: %d", fN, dN)
			},
		},
	).Parse(
		`
文件大小	创建日期		文件(目录)
------------------------------------------------------------------------------{{range .}}
{{convertFileSize .Size}}	{{timeFmt .Ctime}}	{{.Path}} {{end}}
------------------------------------------------------------------------------
总大小: {{totalSize}}	{{fdCount}}
`)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(os.Stdout, files)
	if err != nil {
		panic(err)
	}
}
