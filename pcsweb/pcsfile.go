package pcsweb

import (
	"io"
	"net/http"
)

func fileList(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fpath := r.Form.Get("path")
	dataReadCloser, err := activeAPI.PrepareFilesDirectoriesList(fpath, false)
	if err != nil {
		w.Write((&ErrInfo{
			ErrroCode: 1,
			ErrorMsg:  err.Error(),
		}).JSON())
	}

	io.Copy(w, dataReadCloser)

	dataReadCloser.Close()
}
