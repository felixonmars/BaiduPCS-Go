package pcsweb

import (
	"html/template"
	"net/http"
)

func middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// next handler
		next.ServeHTTP(w, r)
	}
}

func activeAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	next2 := middleware(next)

	err := activeAPIInit()
	checkErr(err) // TODO web登录

	return func(w http.ResponseWriter, r *http.Request) {
		next2.ServeHTTP(w, r)
	}
}

// rootMiddleware 根目录中间件
func rootMiddleware(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// 跳转到 /index.html
		w.Header().Set("Location", "/index.html")
		http.Error(w, "", 301)
	} else {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(404)

		tmpl, err := template.New("index").Parse(templatesBox.MustString("index.html"))
		checkErr(err)

		_, err = tmpl.Parse(templatesBox.MustString("404.html"))
		checkErr(err)

		err = tmpl.Execute(w, nil)
		checkErr(err)
	}
}
