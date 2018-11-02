package dlinkclient_test

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs/dlinkclient"
	"testing"
)

func TestAll(t *testing.T) {
	dc := dlinkclient.NewDlinkClient()
	short, err := dc.ShareReg("http://pan.baidu.com/s/1zn1W7VWuMIjTuIbGafUXMA", "")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("short: %s\n", short)

	list, err := dc.ShareList(short, "/", 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, f := range list {
		fmt.Println(f)
	}

	if len(list) <= 0 {
		return
	}

	nlink, err := dc.LinkRedirect(list[0].Link)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("redirected link: %s\n", nlink)
}
