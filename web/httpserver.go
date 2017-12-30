package pcsweb

import (
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/iikira/BaiduPCS-Go/command"
	"github.com/iikira/BaiduPCS-Go/util"
	"html/template"
	"net/http"
)

var (
	staticBox    *rice.Box // go.rice 文件盒子
	templatesBox *rice.Box // go.rice 文件盒子
)

func init() {
	hb, err := rice.FindBox("static")
	if err != nil {
		fmt.Println(err)
		return
	}
	staticBox = hb

	hb, err = rice.FindBox("template")
	if err != nil {
		fmt.Println(err)
		return
	}
	templatesBox = hb
}

func StartServer() error {
	http.Handle("/lib/", http.StripPrefix("/lib/", http.FileServer(staticBox.HTTPBox())))
	http.HandleFunc("/", indexPage)
	return http.ListenAndServe(":8080", nil)
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// get file contents as string
	indexContents, err := templatesBox.String("index.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	directoryListContents, err := templatesBox.String("directory-list.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	tmpl, err := template.New("index").Parse(indexContents)
	if err != nil {
		fmt.Println(err)
		return
	}

	files, err := baidupcscmd.GetPCSInfo().FileList(r.Form.Get("path"))
	if err != nil {
		fmt.Println(err)
		return
	}

	tmpl.New("directory-list").Funcs(
		template.FuncMap{
			"convertFileSize": func(size int64) string {
				res := pcsutil.ConvertFileSize(size)
				if res == "0" {
					return "-"
				}
				return res
			},
			"timeFmt": pcsutil.FormatTime,
		},
	).Parse(directoryListContents)

	err = tmpl.Execute(w, files)
	if err != nil {
		fmt.Println(err)
		return
	}
}
