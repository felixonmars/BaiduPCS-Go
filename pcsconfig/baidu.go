package pcsconfig

import (
	"bytes"
	"fmt"
	"github.com/iikira/baidu-tools/tieba"
	"strings"
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
	var s bytes.Buffer
	s.WriteString("\nindex\t\tuid\t用户名\n")
	s.WriteString(strings.Repeat("-", 50) + "\n")

	for k, baiduInfo := range *bl {
		s.WriteString(fmt.Sprintf("%4d|", k) + "\t" + fmt.Sprintf("%11d|", baiduInfo.UID) + "\t" + baiduInfo.Name + "\n")
	}
	s.WriteRune('\n')
	return s.String()
}
