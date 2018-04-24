package pcsverbose

import (
	"fmt"
	"io"
	"io/ioutil"
)

//PrintReader 输出Reader
func PrintReader(r io.Reader) {
	b, _ := ioutil.ReadAll(r)
	fmt.Printf("%s\n", b)
}
