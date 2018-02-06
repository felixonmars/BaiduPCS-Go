package baidupcs

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
)

// Remove 批量删除文件/目录
func (p *PCSApi) Remove(paths ...string) (err error) {
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

	p.setApi("file", "delete", map[string]string{
		"param": string(ej[:]),
	})

	resp, err := p.client.Req("POST", p.url.String(), nil, nil)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	json, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		return
	}

	code, msg := CheckErr(json)
	if msg != "" {
		return fmt.Errorf("删除文件/目录 遇到错误, 错误代码: %d, 消息: %s", code, msg)
	}

	return nil
}

// Mkdir 创建目录
func (p *PCSApi) Mkdir(path string) (err error) {
	p.setApi("file", "mkdir", map[string]string{
		"path": path,
	})

	resp, err := p.client.Req("POST", p.url.String(), nil, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	json, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		return
	}

	code, msg := CheckErr(json)
	if msg != "" {
		return fmt.Errorf("创建目录 遇到错误, 错误代码: %d, 消息: %s", code, msg)
	}
	return
}
