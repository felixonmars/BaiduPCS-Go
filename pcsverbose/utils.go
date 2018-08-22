package pcsverbose

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil/pcstime"
	"io"
	"io/ioutil"
)

//PrintReader 输出Reader
func PrintReader(r io.Reader) {
	b, _ := ioutil.ReadAll(r)
	fmt.Printf("%s\n", b)
}

func TimePrefix() string {
	return "[" + pcstime.BeijingTimeOption("Refer") + "]"
}
