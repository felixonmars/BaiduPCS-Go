package baidupcs

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/requester/multipartreader"
	"github.com/json-iterator/go"
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
func (p *PCSApi) Rename(from, to string) (err error) {
	return p.cpmvOp("rename", CpMvJSON{
		From: from,
		To:   to,
	})
}

// Copy 批量拷贝文件/目录
func (p *PCSApi) Copy(cpmvJSON ...CpMvJSON) (err error) {
	return p.cpmvOp("copy", cpmvJSON...)
}

// Move 批量移动文件/目录
func (p *PCSApi) Move(cpmvJSON ...CpMvJSON) (err error) {
	return p.cpmvOp("move", cpmvJSON...)
}

func (p *PCSApi) cpmvOp(op string, cpmvJSON ...CpMvJSON) (err error) {
	var operation string
	switch op {
	case "copy":
		operation = "拷贝文件/目录"
	case "move":
		operation = "移动文件/目录"
	case "rename":
		operation = "重命名文件/目录"
	default:
		panic("Unknown op: " + op)
	}

	ejs, err := cpmvJSONEncode(cpmvJSON...)
	if err != nil {
		return err
	}

	method := op
	if method == "rename" {
		method = "move"
	}

	p.setAPI("file", method)

	// 表单上传
	mr := multipartreader.NewMultipartReader()
	mr.AddFormFeild("param", strings.NewReader(ejs))

	resp, err := p.client.Req("POST", p.url.String(), mr, nil)
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

	return nil
}

//cpmvJSONEncode 生成 json 串
func cpmvJSONEncode(cpmvJSON ...CpMvJSON) (string, error) {
	pathsData := CpMvJSONList{
		List: cpmvJSON,
	}

	ej, err := jsoniter.Marshal(&pathsData)
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
