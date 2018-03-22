package pcsweb

import (
	"html/template"
)

// boxTmplParse ricebox 载入文件内容, 并进行模板解析
func boxTmplParse(name string, fileNames ...string) (tmpl *template.Template) {
	var (
		err error
	)
	tmpl = template.New(name)
	for k := range fileNames {
		_, err = tmpl.Parse(templatesBox.MustString(fileNames[k]))
		checkErr(err)
	}
	return
}
