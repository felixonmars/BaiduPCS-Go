package uploader

import (
	"bytes"
	"github.com/iikira/BaiduPCS-Go/requester"
	"mime/multipart"
	"net/http"
	"strings"
)

// Uploader 上传
type Uploader struct {
	URL  string  // 上传地址
	Body *reader // 要上传的对象

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

// NewUploader 返回 uploader 对象, url: 上传地址, uploadReaderLen: 实现 uploader.ReaderLen 接口的对象, 例如文件
func NewUploader(url string, uploadReaderLen ReaderLen, o *Options) (uploader *Uploader) {
	uploader = &Uploader{
		URL: url,
		Body: &reader{
			uploadReaderLen: uploadReaderLen,
		},
	}

	if o == nil {
		uploader.Options = &Options{
			IsMultiPart: false,
			Client:      requester.NewHTTPClient(),
		}

		return
	}

	if o.Client == nil {
		o.Client = requester.NewHTTPClient()
	}

	// 设置不超时
	o.Client.SetTimeout(0)
	o.Client.SetResponseHeaderTimeout(0)

	uploader.Options = o
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

		if checkFunc != nil {
			checkFunc(resp, err)
		}

		u.finished = true

		u.touch(u.onFinish) // 触发上传结束的事件
		finish <- struct{}{}
	}()
	return finish
}

func (u *Uploader) execute() (resp *http.Response, code int, err error) {
	var contentType string
	if u.Options.IsMultiPart {
		multipartWriter := &bytes.Buffer{}
		writer := multipart.NewWriter(multipartWriter)
		writer.CreateFormFile("uploadedfile", "")

		u.Body.multipart = multipartWriter
		u.Body.multipartEnd = strings.NewReader("\r\n--" + writer.Boundary() + "--\r\n")
		contentType = writer.FormDataContentType()
	} else {
		contentType = "application/x-www-form-urlencoded"
	}

	req, err := http.NewRequest("POST", u.URL, u.Body)
	if err != nil {
		return nil, 1, err
	}

	req.Header.Add("Content-Type", contentType)

	// 设置 Content-Length 不然请求会卡住不动!!!
	req.ContentLength = u.Body.totalLen()

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
