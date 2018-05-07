package pcsweb

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"io"
	"net/http"
)

func fileList(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fpath := r.Form.Get("path")
	dataReadCloser, err := pcsconfig.Config.ActiveUserBaiduPCS().PrepareFilesDirectoriesList(fpath, baidupcs.DefaultOrderOptions)
	if err != nil {
		w.Write((&ErrInfo{
			ErrroCode: 1,
			ErrorMsg:  err.Error(),
		}).JSON())
		return
	}

	defer dataReadCloser.Close()
	io.Copy(w, dataReadCloser)
}
