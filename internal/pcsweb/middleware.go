package pcsweb

import (
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

	// TODO web登录

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

		tmpl := boxTmplParse("index", "index.html", "404.html")
		checkErr(tmpl.Execute(w, nil))
	}
}
