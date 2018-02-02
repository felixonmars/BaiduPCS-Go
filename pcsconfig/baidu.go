package pcsconfig

import (
	"github.com/iikira/baidu-tools/tieba"
)

// Baidu 百度帐号对象
type Baidu struct {
	UID    uint64 `json:"uid"`
	Name   string `json:"name"`
	BDUSS  string `json:"bduss"`
	PTOKEN string `json:"ptoken"`
	STOKEN string `json:"stoken"`

	Workdir string `json:"workdir"`
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
		BDUSS:   t.Baidu.Auth.BDUSS,
		Workdir: "/",
	}
	return b, nil
}
