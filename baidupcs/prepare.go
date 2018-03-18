package baidupcs

import (
	"bytes"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/requester/multipartreader"
	"github.com/json-iterator/go"
	"io"
)

// PrepareFilesDirectoriesBatchMeta 获取多个文件/目录的元信息, 只返回服务器响应数据和错误信息
func (p *PCSApi) PrepareFilesDirectoriesBatchMeta(paths ...string) (dataReadCloser io.ReadCloser, err error) {
	type listStr struct {
		Path string `json:"path"`
	}

	type postStr struct {
		List []listStr `json:"list"`
	}

	// json 数据构造
	post := &postStr{
		List: make([]listStr, len(paths)),
	}

	for k := range paths {
		post.List[k].Path = paths[k]
	}

	sendData, err := jsoniter.Marshal(post)
	if err != nil {
		panic(operationFilesDirectoriesBatchMeta + ", json 数据构造失败, " + err.Error())
	}

	p.setAPI("file", "meta")

	// 表单上传
	mr := multipartreader.NewMultipartReader()
	mr.AddFormFeild("param", bytes.NewReader(sendData))

	resp, err := p.client.Req("POST", p.url.String(), mr, nil)
	if err != nil {
		return nil, fmt.Errorf("%s, 网络错误, %s", operationFilesDirectoriesBatchMeta, err)
	}

	return resp.Body, nil
}

// PrepareFilesDirectoriesList 获取目录下的文件和目录列表, 可选是否递归, 只返回服务器响应数据和错误信息
func (p *PCSApi) PrepareFilesDirectoriesList(path string, recurse bool) (dataReadCloser io.ReadCloser, err error) {
	if path == "" {
		path = "/"
	}

	p.setAPI("file", "list", map[string]string{
		"path":  path,
		"by":    "name",
		"order": "asc", // 升序
		"limit": "0-2147483647",
	})

	resp, err := p.client.Req("GET", p.url.String(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%s, 网络错误, %s", operationFilesDirectoriesList, err)
	}

	return resp.Body, nil
}
