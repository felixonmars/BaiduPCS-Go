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
	URL         string // 上传地址
	IsMultiPart bool   // 是否表单上传

	Body *reader // 要上传的对象

	client *requester.HTTPClient

	onExecute func()
	onFinish  func()
}

// NewUploader 返回 uploader 对象, url: 上传地址, isMultipart: 是否表单上传,uploadReaderLen: 实现 uploader.ReaderLen 接口的对象, 例如文件
func NewUploader(url string, isMultipart bool, uploadReaderLen ReaderLen, h *requester.HTTPClient) (uploader *Uploader) {
	uploader = &Uploader{
		URL:         url,
		IsMultiPart: isMultipart,
		Body: &reader{
			uploadReaderLen: uploadReaderLen,
		},
	}

	if h == nil {
		uploader.client = requester.NewHTTPClient()
	} else {
		uploader.client = h
	}

	// 设置不超时
	uploader.client.SetTimeout(0)
	uploader.client.SetResponseHeaderTimeout(0)
	return
}

// Execute 执行上传
func (u *Uploader) Execute(checkFunc func(resp *http.Response, err error)) {
	go func() {
		u.touch(u.onExecute)

		// 开始上传
		resp, _, err := u.execute()

		if checkFunc != nil {
			checkFunc(resp, err)
		}
		u.touch(u.onFinish)
	}()
}

func (u *Uploader) execute() (resp *http.Response, code int, err error) {
	var contentType string
	if u.IsMultiPart {
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

	resp, err = u.client.Do(req)
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
