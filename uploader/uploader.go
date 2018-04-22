// Package uploader 上传包
package uploader

import (
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/BaiduPCS-Go/requester/multipartreader"
	"github.com/iikira/BaiduPCS-Go/requester/rio"
	"net/http"
)

// Uploader 上传
type Uploader struct {
	URL  string                      // 上传地址
	Body multipartreader.ReadedLen64 // 要上传的对象

	Options *Options

	UploadStatus <-chan UploadStatus // 上传状态
	finished     bool

	onExecute func()
	onFinish  func()
}

// Options are the options for creating a new Uploader
type Options struct {
	IsMultiPart bool                  // 是否表单上传
	Client      *requester.HTTPClient // http 客户端
}

// NewUploader 返回 uploader 对象, url: 上传地址, readedlen64: 实现 multipartreader.ReadedLen64 接口的对象, 例如文件
func NewUploader(url string, readedlen64 multipartreader.ReadedLen64, o *Options) (uploader *Uploader) {
	uploader = &Uploader{
		URL:     url,
		Body:    readedlen64,
		Options: o,
	}

	if uploader.Options == nil {
		uploader.Options = &Options{
			IsMultiPart: false,
			Client:      requester.NewHTTPClient(),
		}
	}

	if uploader.Options.Client == nil {
		uploader.Options.Client = requester.NewHTTPClient()
	}

	// 设置不超时
	uploader.Options.Client.SetTimeout(0)
	uploader.Options.Client.SetResponseHeaderTimeout(0)
	return
}

// Execute 执行上传, 收到返回值信号则为上传结束
func (u *Uploader) Execute(checkFunc func(resp *http.Response, err error)) <-chan struct{} {
	finish := make(chan struct{}, 0)
	u.startStatus()
	go func() {
		u.touch(u.onExecute)

		// 开始上传
		resp, _, err := u.execute()

		// 上传结束
		u.finished = true

		if checkFunc != nil {
			checkFunc(resp, err)
		}

		u.touch(u.onFinish) // 触发上传结束的事件

		finish <- struct{}{}
	}()
	return finish
}

func (u *Uploader) execute() (resp *http.Response, code int, err error) {
	var (
		contentType string
		obody       rio.ReaderLen64
	)

	if u.Options.IsMultiPart {
		mr := multipartreader.NewMultipartReader()
		mr.AddFormFile("uploadedfile", "", u.Body)

		contentType = mr.ContentType()
		obody = mr
	} else {
		contentType = "application/x-www-form-urlencoded"
		obody = u.Body
	}

	req, err := http.NewRequest("POST", u.URL, obody)
	if err != nil {
		return nil, 1, err
	}

	req.Header.Add("Content-Type", contentType)

	// 设置 Content-Length 不然请求会卡住不动!!!
	req.ContentLength = obody.Len()

	resp, err = u.Options.Client.Do(req)
	if err != nil {
		return nil, 2, err
	}

	return resp, 0, nil
}

// touch 用于触发事件
func (u *Uploader) touch(fn func()) {
	if fn != nil {
		go fn()
	}
}

// OnExecute 任务开始时触发的事件
func (u *Uploader) OnExecute(fn func()) {
	u.onExecute = fn
}

// OnFinish 任务完成时触发的事件
func (u *Uploader) OnFinish(fn func()) {
	u.onFinish = fn
}
