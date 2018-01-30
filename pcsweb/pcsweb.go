package pcsweb

import (
	"bytes"
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/iikira/BaiduPCS-Go/pcscommand"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"html/template"
	"net/http"
	"path/filepath"
)

var (
	staticBox    = new(rice.Box) // go.rice 文件盒子
	templatesBox = new(rice.Box) // go.rice 文件盒子
)

func boxInit() error {
	hb, err := rice.FindBox("static")
	if err != nil {
		return err
	}
	staticBox = hb

	hb, err = rice.FindBox("template")
	if err != nil {
		return err
	}
	templatesBox = hb
	return nil
}

// StartServer 开启web服务
func StartServer(port uint) error {
	if port <= 0 || port >= 0x10000 {
		return fmt.Errorf("invalid port %d", port)
	}

	err := boxInit()
	if err != nil {
		return err
	}

	http.Handle("/lib/", http.StripPrefix("/lib/", http.FileServer(staticBox.HTTPBox())))
	http.HandleFunc("/about.html", aboutPage)
	http.HandleFunc("/", indexPage)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func aboutPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index.html").Funcs(
		template.FuncMap{
			"include": tplInclude,
		},
	).Parse(templatesBox.MustString("index.html"))

	tmpl.Parse(templatesBox.MustString("about.html"))
	if err != nil {
		fmt.Println(err)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tmpl, err := template.New("index.html").Funcs(
		template.FuncMap{
			"include": tplInclude,
		},
	).Parse(templatesBox.MustString("index.html"))
	if err != nil {
		fmt.Println(err)
		return
	}

	files, err := pcscommand.GetPCSInfo().FileList(r.Form.Get("path"))
	if err != nil {
		fmt.Println(err)
		return
	}

	tmpl.Funcs(
		template.FuncMap{
			"getPath": func() string {
				return r.Form.Get("path")
			},
			"convertFileSize": func(size int64) string {
				res := pcsutil.ConvertFileSize(size)
				if res == "0" {
					return "-"
				}
				return res
			},
			"timeFmt": pcsutil.FormatTime,
		},
	).Parse(templatesBox.MustString("baidu/userinfo.html"))

	err = tmpl.Execute(w, files)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func tplInclude(file string, dot interface{}) template.HTML {
	var buffer = &bytes.Buffer{}

	// get file contents as string
	contents, err := templatesBox.String(file)
	if err != nil {
		fmt.Printf("get rice.box contents(%s) error: %s\n", file, err)
		return ""
	}

	tpl, err := template.New(filepath.Base(file)).Funcs(
		template.FuncMap{
			"include": tplInclude,
		},
	).Parse(contents)
	if err != nil {
		fmt.Printf("parse template file(%s) error:%v\n", file, err)
		return ""
	}
	err = tpl.Execute(buffer, dot)
	if err != nil {
		fmt.Printf("template file(%s) syntax error:%v", file, err)
		return ""
	}
	return template.HTML(buffer.String())
}

func render() {}
