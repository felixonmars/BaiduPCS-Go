package uploader

import (
	"bytes"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/requester"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
)

// Uploader 上传
type Uploader struct {
	URL    string  // 上传地址
	Reader *reader // 要上传的对象

	client *requester.HTTPClient

	onExecute func()
	onFinish  func()
}

// NewUploader 返回 uploader 对象, url: 上传地址, uploadReader: 实现 io.Reader 接口的对象, 例如文件
func NewUploader(url string, uploadReader io.Reader, size int64, h *requester.HTTPClient) (uploader *Uploader) {
	uploader = new(Uploader)
	uploader.URL = url
	uploader.Reader = &reader{
		uploadReader: uploadReader,
		size:         size,
	}

	// 设置不超时
	defer h.SetTimeout(0)

	if h == nil {
		uploader.client = requester.NewHTTPClient()
		return
	}
	uploader.client = h
	return
}

// Execute 执行上传
func (u *Uploader) Execute(checkFunc func(respBodyContents []byte, err error)) {
	go func() {
		u.touch(u.onExecute)

		// 开始上传
		respBodyContents, _, err := u.execute()

		u.touch(u.onFinish)
		if checkFunc != nil {
			checkFunc(respBodyContents, err)
		}
	}()
}

func (u *Uploader) execute() (respBodyContents []byte, code int, err error) {
	multipartWriter := &bytes.Buffer{}
	writer := multipart.NewWriter(multipartWriter)
	writer.CreateFormFile("uploadedfile", "")

	u.Reader.multipart = multipartWriter
	u.Reader.multipartEnd = strings.NewReader(fmt.Sprintf("\r\n--%s--\r\n", writer.Boundary()))

	req, err := http.NewRequest("POST", u.URL, u.Reader)
	if err != nil {
		return nil, 1, err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	// 设置 Content-Length 不然请求会卡住不动!!!
	req.ContentLength = u.Reader.Len()

	resp, err := u.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, 2, err
	}

	defer resp.Body.Close()

	respBodyContents, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return respBodyContents, 3, fmt.Errorf("uploader: read response body error, %s", err)
	}

	return respBodyContents, 0, nil
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
