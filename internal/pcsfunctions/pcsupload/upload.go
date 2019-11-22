package pcsupload

import (
	"context"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/requester/multipartreader"
	"github.com/iikira/BaiduPCS-Go/requester/rio"
	"github.com/iikira/BaiduPCS-Go/requester/uploader"
	"net/http"
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

// Precreate do nothing
func (pu *PCSUpload) Precreate() (err error) {
	return nil
}

func (pu *PCSUpload) TmpFile(ctx context.Context, partseq int, partOffset int64, r rio.ReaderLen64) (checksum string, uperr error) {
	pu.lazyInit()

	var respErr *uploader.MultiError
	checksum, pcsError := pu.pcs.UploadTmpFile(func(uploadURL string, jar http.CookieJar) (resp *http.Response, err error) {
		client := pcsconfig.Config.PCSHTTPClient()
		client.SetCookiejar(jar)
		client.SetTimeout(0)

		mr := multipartreader.NewMultipartReader()
		mr.AddFormFile("uploadedfile", "", r)
		mr.CloseMultipart()

		doneChan := make(chan struct{}, 1)
		go func() {
			resp, err = client.Req(http.MethodPost, uploadURL, mr, nil)
			doneChan <- struct{}{}

			if resp != nil {
				// 不可恢复的错误
				switch resp.StatusCode {
				case 400, 401, 403, 413:
					respErr = &uploader.MultiError{
						Terminated: true,
					}
				}
			}
		}()
		select {
		case <-ctx.Done(): // 取消
			// 返回, 让那边关闭连接
			return resp, ctx.Err()
		case <-doneChan:
			// return
		}
		return
	})

	if respErr != nil {
		respErr.Err = pcsError
		return checksum, respErr
	}

	return checksum, pcsError
}

func (pu *PCSUpload) CreateSuperFile(checksumList ...string) (err error) {
	pu.lazyInit()
	return pu.pcs.UploadCreateSuperFile(pu.targetPath, checksumList...)
}
