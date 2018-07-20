// Package pcsupload 上传包
package pcsupload

import (
	"context"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/requester/multipartreader"
	"github.com/iikira/BaiduPCS-Go/requester/rio"
	"github.com/iikira/BaiduPCS-Go/requester/uploader"
	"net/http"
	"net/http/cookiejar"
)

const (
	UploadingFileName = "pcs_uploading.json"
)

type (
	PCSUpload struct {
		pcs        *baidupcs.BaiduPCS
		targetPath string
	}
)

func NewPCSUpload(pcs *baidupcs.BaiduPCS, targetPath string) uploader.MultiUpload {
	return &PCSUpload{
		pcs:        pcs,
		targetPath: targetPath,
	}
}

func (pu *PCSUpload) lazyInit() {
	if pu.pcs == nil {
		pu.pcs = &baidupcs.BaiduPCS{}
	}
}

func (pu *PCSUpload) TmpFile(ctx context.Context, r rio.ReaderLen64) (checksum string, uperr error) {
	pu.lazyInit()
	return pu.pcs.UploadTmpFile(func(uploadURL string, jar *cookiejar.Jar) (resp *http.Response, err error) {
		client := pcsconfig.Config.HTTPClient()
		client.SetCookiejar(jar)
		client.SetTimeout(0)

		mr := multipartreader.NewMultipartReader()
		mr.AddFormFile("uploadedfile", "", r)
		mr.CloseMultipart()

		doneChan := make(chan struct{}, 1)
		go func() {
			resp, err = client.Req("POST", uploadURL, mr, nil)
			doneChan <- struct{}{}
		}()
		select {
		case <-ctx.Done():
			return resp, ctx.Err()
		case <-doneChan:
			// return
		}
		return
	})
}

func (pu *PCSUpload) CreateSuperFile(checksumList ...string) (err error) {
	pu.lazyInit()
	return pu.pcs.UploadCreateSuperFile(pu.targetPath, checksumList...)
}
