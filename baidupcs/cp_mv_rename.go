package baidupcs

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/iikira/BaiduPCS-Go/requester"
	"strings"
)

// CpMvJSON 源文件目录的地址和目标文件目录的地址
type CpMvJSON struct {
	From string `json:"from"` // 源文件或目录
	To   string `json:"to"`   // 目标文件或目录
}

// CpMvJSONList []CpMvJSON 对象数组
type CpMvJSONList struct {
	List []CpMvJSON `json:"list"`
}

// Rename 重命名文件/目录
func (p PCSApi) Rename(from, to string) (err error) {
	return p.cpmvOp("rename", CpMvJSON{
		From: from,
		To:   to,
	})
}

// Copy 批量拷贝文件/目录
func (p PCSApi) Copy(cpmvJSON ...CpMvJSON) (err error) {
	return p.cpmvOp("copy", cpmvJSON...)
}

// Move 批量移动文件/目录
func (p PCSApi) Move(cpmvJSON ...CpMvJSON) (err error) {
	return p.cpmvOp("move", cpmvJSON...)
}

func (p PCSApi) cpmvOp(op string, cpmvJSON ...CpMvJSON) (err error) {
	ejs, err := cpmvJSONEncode(cpmvJSON...)
	if err != nil {
		return err
	}

	method := op
	if method == "rename" {
		method = "move"
	}

	p.addItem("file", method, map[string]string{
		"param": ejs,
	})

	h := requester.NewHTTPClient()
	body, err := h.Fetch("POST", p.url.String(), nil, map[string]string{
		"Cookie": "BDUSS=" + p.bduss,
	})
	if err != nil {
		return err
	}

	json, err := simplejson.NewJson(body)
	if err != nil {
		return
	}

	code, err := CheckErr(json)
	if err != nil {
		switch op {
		case "copy":
			return fmt.Errorf("拷贝文件/目录 遇到错误, 错误代码: %d, 消息: %s", code, err)
		case "move":
			return fmt.Errorf("移动文件/目录 遇到错误, 错误代码: %d, 消息: %s", code, err)
		case "rename":
			return fmt.Errorf("重命名文件/目录 遇到错误, 错误代码: %d, 消息: %s", code, err)
		default:
			panic("Unknown op: " + op)
		}
	}

	return nil
}

//cpmvJSONEncode 生成 json 串
func cpmvJSONEncode(cpmvJSON ...CpMvJSON) (string, error) {
	pathsData := CpMvJSONList{
		List: cpmvJSON,
	}

	ej, err := json.Marshal(&pathsData)
	if err != nil {
		return "", err
	}

	return string(ej[:]), nil
}

func (cl CpMvJSONList) String() string {
	l := make([]string, len(cl.List))
	for k := range cl.List {
		l[k] = fmt.Sprintf("%d: %s -> %s", k+1, cl.List[k].From, cl.List[k].To)
	}
	return strings.Join(l, "\n")
}
