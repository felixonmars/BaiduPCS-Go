package pcsconfig

import (
	"github.com/iikira/baidu-tools/tieba"
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

// NewWithBDUSS 检测BDUSS有效性, 同时获取百度详细信息 (无法获取 ptoken 和 stoken)
func NewWithBDUSS(bduss string) (b *Baidu, err error) {
	t, err := tieba.NewWithBDUSS(bduss)
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
