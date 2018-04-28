// Package getip 获取 ip 信息包
package getip

import (
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/iikira/BaiduPCS-Go/requester"
	"regexp"
	"unsafe"
)

const (
	//StrUnknown unknown
	StrUnknown = "unknown"
)

var (
	ipExp = regexp.MustCompile("{ip:'(.*?)',address:'(.*?)'}")
)

//IPInfo 获取IP地址和IP位置
func IPInfo() (ipAddr, location string) {
	body, err := requester.Fetch("GET", "http://ip.chinaz.com/getip.aspx", nil, nil)
	if err != nil {
		pcsverbose.Verbosef("DEBUG: getip: %s\n", err)
		return StrUnknown, StrUnknown
	}

	raw := ipExp.FindSubmatch(body)
	if len(raw) < 3 {
		pcsverbose.Verboseln("DEBUG: getip: regexp match failed")
		return StrUnknown, StrUnknown
	}

	return *(*string)(unsafe.Pointer(&raw[1])), *(*string)(unsafe.Pointer(&raw[2]))
}
