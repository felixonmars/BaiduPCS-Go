package uploader

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/BaiduPCS-Go/util"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
func NewUploader(url string, uploadReader io.Reader, h *requester.HTTPClient) (uploader *Uploader) {
	uploader = new(Uploader)
	uploader.URL = url
	uploader.Reader = &reader{
		reader: uploadReader,
	}

	// 设置不超时
	defer func() {
		h.SetResponseHeaderTimeout(0)
		h.SetTimeout(0)
	}()

	if h == nil {
		h = requester.NewHTTPClient()
		uploader.client = h
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
	// 缓存文件示例: /tmp/BaiduPCS-GO_uploading_0CF81040068E7893B998D20DC66FD438
	tempfilePath := filepath.Join(os.TempDir(), "BaiduPCS-GO_uploading"+pcsutil.Md5Encrypt(u.URL))
	tempfile, err := os.Create(tempfilePath)
	if err != nil {
		return nil, 1, fmt.Errorf("uploader: temp file failed, %s", err)
	}

	writer := multipart.NewWriter(tempfile)
	part, err := writer.CreateFormFile("uploadedfile", "")
	if err != nil {
		return nil, 1, err
	}

	io.Copy(part, u.Reader.reader)

	writer.Close()

	tempfile, _ = os.Open(tempfilePath) // 重新打开文件
	tempfileInfo, _ := tempfile.Stat()

	u.Reader = &reader{
		reader: tempfile,
		length: tempfileInfo.Size(),
	}

	req, err := http.NewRequest("POST", u.URL, u.Reader)
	if err != nil {
		return nil, 1, err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	// 设置 Content-Length 不然请求会卡住不动!!!
	req.ContentLength = tempfileInfo.Size()

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, 2, err
	}

	defer func() {
		resp.Body.Close()

		// 移除缓存文件
		tempfile.Close()
		os.Remove(tempfilePath)
	}()

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
