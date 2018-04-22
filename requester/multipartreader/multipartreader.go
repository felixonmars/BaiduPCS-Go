// Package multipartreader helps you encode large files in MIME multipart format
// without reading the entire content into memory.
package multipartreader

import (
	"bytes"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/requester/rio"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

// MultipartReader MIME multipart format
type MultipartReader struct {
	contentType string
	boundary    string

	formBody  string
	parts     []*part
	part64s   []*part64
	formClose string

	once        sync.Once
	multiReader io.Reader

	_      bool  // alignmemt
	readed int64 // 已读取的数据量
}

type part struct {
	form      string
	readerlen rio.ReaderLen
}

type part64 struct {
	form        string
	readerlen64 rio.ReaderLen64
}

// NewMultipartReader 返回初始化的 *MultipartReader
func NewMultipartReader() (mr *MultipartReader) {
	buf := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buf)

	mr = &MultipartReader{
		contentType: writer.FormDataContentType(),
		boundary:    writer.Boundary(),
	}

	mr.formBody = buf.String()
	mr.formClose = "\r\n--" + mr.boundary + "--\r\n"
	return
}

// ContentType 返回 Content-Type
func (mr *MultipartReader) ContentType() string {
	return mr.contentType
}

// AddFormFeild 增加 form 表单
func (mr *MultipartReader) AddFormFeild(fieldname string, readerlen rio.ReaderLen) {
	mr.parts = append(mr.parts, &part{
		form:      fmt.Sprintf("--%s\r\nContent-Disposition: form-data; name=\"%s\"\r\n\r\n", mr.boundary, fieldname),
		readerlen: readerlen,
	})
}

// AddFormFile 增加 form 文件表单
func (mr *MultipartReader) AddFormFile(fieldname, filename string, readerlen64 rio.ReaderLen64) {
	mr.part64s = append(mr.part64s, &part64{
		form:        fmt.Sprintf("--%s\r\nContent-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n\r\n", mr.boundary, fieldname, filename),
		readerlen64: readerlen64,
	})
}

// SetupHTTPRequest 为 *http.Request 配置
func (mr *MultipartReader) SetupHTTPRequest(req *http.Request) {
	req.Header.Add("Content-Type", mr.contentType)

	// 设置 Content-Length 不然请求会卡住不动!!!
	req.ContentLength = mr.Len()
}

func (mr *MultipartReader) Read(p []byte) (n int, err error) {
	mr.once.Do(func() {
		readers := []io.Reader{
			strings.NewReader(mr.formBody),
		}

		for _, part := range mr.parts {
			if part == nil {
				continue
			}
			readers = append(readers, strings.NewReader(part.form), part.readerlen)
		}

		for _, part64 := range mr.part64s {
			if part64 == nil {
				continue
			}
			readers = append(readers, strings.NewReader(part64.form), part64.readerlen64)
		}

		readers = append(readers, strings.NewReader(mr.formClose))

		mr.multiReader = io.MultiReader(readers...)
	})

	n, err = mr.multiReader.Read(p)
	mr.addReaded(int64(n))
	return n, err
}

// Len 返回表单内容总长度
func (mr *MultipartReader) Len() int64 {
	var (
		i32 int
		i64 int64
	)

	i32 += len(mr.formBody)
	for _, part := range mr.parts {
		if part == nil {
			continue
		}
		i32 += len(part.form) + part.readerlen.Len()
	}

	for _, part64 := range mr.part64s {
		if part64 == nil {
			continue
		}
		i32 += len(part64.form)
		i64 += part64.readerlen64.Len()
	}

	i32 += len(mr.formClose)

	return int64(i32) + i64
}

// Readed 返回 form 表单已读取的数据量, 用于计算上传速度等
func (mr *MultipartReader) Readed() int64 {
	return atomic.LoadInt64(&mr.readed)
}

func (mr *MultipartReader) addReaded(i int64) int64 {
	return atomic.AddInt64(&mr.readed, i)
}
