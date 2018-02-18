package pcsconfig

import (
	"bytes"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/iikira/baidu-tools/tieba"
	"strconv"
)

// Baidu 百度帐号对象
type Baidu struct {
	UID  uint64  `json:"uid"`  // 百度ID对应的uid
	Name string  `json:"name"` // 真实ID
	Sex  string  `json:"sex"`  // 性别
	Age  float64 `json:"age"`  // 帐号年龄

	BDUSS  string `json:"bduss"`
	PTOKEN string `json:"ptoken"`
	STOKEN string `json:"stoken"`

	Workdir string `json:"workdir"` // 工作目录
}

// BaiduUserList 百度帐号列表
type BaiduUserList []*Baidu

// NewUserInfoByBDUSS 检测BDUSS有效性, 同时获取百度详细信息 (无法获取 ptoken 和 stoken)
func NewUserInfoByBDUSS(bduss string) (b *Baidu, err error) {
	t, err := tieba.NewUserInfoByBDUSS(bduss)
	if err != nil {
		return nil, err
	}

	b = &Baidu{
		UID:     t.Baidu.UID,
		Name:    t.Baidu.Name,
		Sex:     t.Baidu.Sex,
		Age:     t.Baidu.Age,
		BDUSS:   t.Baidu.Auth.BDUSS,
		Workdir: "/",
	}
	return b, nil
}

// String 格式输出百度帐号列表
func (bl *BaiduUserList) String() string {
	buf := bytes.NewBuffer(nil)

	tb := pcstable.NewTable(buf)
	tb.SetHeader([]string{"#", "uid", "用户名"})

	for k, baiduInfo := range *bl {
		tb.Append([]string{strconv.Itoa(k), strconv.FormatUint(baiduInfo.UID, 10), baiduInfo.Name})
	}

	tb.Render()

	return buf.String()
}
