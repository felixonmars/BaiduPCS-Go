package baidupcs

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/iikira/BaiduPCS-Go/requester"
)

// Remove 批量删除文件/目录
func (p PCSApi) Remove(paths ...string) (err error) {
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

	ej, err := json.Marshal(&pathsData)
	if err != nil {
		return
	}

	p.addItem("file", "delete", map[string]string{
		"param": string(ej[:]),
	})

	h := requester.NewHTTPClient()
	body, err := h.Fetch("POST", p.url.String(), nil, map[string]string{
		"Cookie": "BDUSS=" + p.bduss,
	})
	if err != nil {
		return
	}

	json, err := simplejson.NewJson(body)
	if err != nil {
		return
	}

	code, err := CheckErr(json)
	if err != nil {
		return fmt.Errorf("删除文件/目录 遇到错误, 错误代码: %d, 消息: %s", code, err)
	}

	return nil
}

// Mkdir 创建目录
func (p PCSApi) Mkdir(path string) (err error) {
	p.addItem("file", "mkdir", map[string]string{
		"path": path,
	})

	h := requester.NewHTTPClient()
	body, err := h.Fetch("POST", p.url.String(), nil, map[string]string{
		"Cookie": "BDUSS=" + p.bduss,
	})
	if err != nil {
		return
	}

	json, err := simplejson.NewJson(body)
	if err != nil {
		return
	}

	code, err := CheckErr(json)
	if err != nil {
		return fmt.Errorf("创建目录 遇到错误, 错误代码: %d, 消息: %s", code, err)
	}
	return
}
