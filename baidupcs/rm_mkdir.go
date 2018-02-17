package baidupcs

import (
	"fmt"
	"github.com/json-iterator/go"
)

// Remove 批量删除文件/目录
func (p *PCSApi) Remove(paths ...string) (err error) {
	operation := "删除文件/目录"

	pathsData := struct {
		List []struct {
			Path string `json:"path"`
		} `json:"list"`
	}{}

	for k := range paths {
		pathsData.List = append(pathsData.List, struct {
			Path string `json:"path"`
		}{
			Path: paths[k],
		})
	}

	ej, err := jsoniter.Marshal(&pathsData)
	if err != nil {
		return
	}

	p.setAPI("file", "delete", map[string]string{
		"param": string(ej[:]),
	})

	resp, err := p.client.Req("POST", p.url.String(), nil, nil)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	errInfo := NewErrorInfo(operation)

	d := jsoniter.NewDecoder(resp.Body)
	err = d.Decode(errInfo)
	if err != nil {
		return fmt.Errorf("%s, json 数据解析失败, %s", operation, err)
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}

	return nil
}

// Mkdir 创建目录
func (p *PCSApi) Mkdir(path string) (err error) {
	operation := "创建目录"

	p.setAPI("file", "mkdir", map[string]string{
		"path": path,
	})

	resp, err := p.client.Req("POST", p.url.String(), nil, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	errInfo := NewErrorInfo(operation)

	d := jsoniter.NewDecoder(resp.Body)
	err = d.Decode(errInfo)
	if err != nil {
		return fmt.Errorf("%s, json 数据解析失败, %s", operation, err)
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}
	return
}
