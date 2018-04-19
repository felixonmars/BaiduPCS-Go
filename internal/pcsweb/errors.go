package pcsweb

import (
	"fmt"
	"github.com/json-iterator/go"
)

// ErrInfo web 错误详情
type ErrInfo struct {
	ErrroCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

func (ei *ErrInfo) Error() string {
	return fmt.Sprintf("error code: %d, error message: %s", ei.ErrroCode, ei.ErrorMsg)
}

// JSON 将错误信息打包成 json
func (ei *ErrInfo) JSON() (data []byte) {
	var err error
	data, err = jsoniter.MarshalIndent(ei, "", " ")
	checkErr(err)

	return
}

// checkErr 遇到错误就退出
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
