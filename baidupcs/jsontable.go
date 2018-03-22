package baidupcs

import (
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/json-iterator/go"
	"strconv"
	"strings"
)

// PathJSON 网盘路径
type PathJSON struct {
	Path string `json:"path"`
}

// PathsListJSON 网盘路径列表
type PathsListJSON struct {
	List []*PathJSON `json:"list"`
}

// CpMvJSON 源文件目录的地址和目标文件目录的地址
type CpMvJSON struct {
	From string `json:"from"` // 源文件或目录
	To   string `json:"to"`   // 目标文件或目录
}

// CpMvListJSON []*CpMvJSON 对象数组
type CpMvListJSON struct {
	List []*CpMvJSON `json:"list"`
}

// JSON json 数据构造
func (plj *PathsListJSON) JSON(paths ...string) (data []byte, err error) {
	plj.List = make([]*PathJSON, len(paths))

	for k := range paths {
		plj.List[k] = &PathJSON{
			Path: paths[k],
		}
	}

	data, err = jsoniter.Marshal(plj)
	return
}

// JSON json 数据构造
func (cj *CpMvJSON) JSON() (data []byte, err error) {
	data, err = jsoniter.Marshal(cj)
	return
}

// JSON json 数据构造
func (clj *CpMvListJSON) JSON() (data []byte, err error) {
	data, err = jsoniter.Marshal(clj)
	return
}

func (clj *CpMvListJSON) String() string {
	builder := &strings.Builder{}

	tb := pcstable.NewTable(builder)
	tb.SetHeader([]string{"#", "原路径", "目标路径"})

	for k := range clj.List {
		if clj.List[k] == nil {
			continue
		}
		tb.Append([]string{strconv.Itoa(k), clj.List[k].From, clj.List[k].To})
	}

	tb.Render()
	return builder.String()
}
