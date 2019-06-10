package getip

import (
	"github.com/iikira/BaiduPCS-Go/pcsutil/jsonhelper"
	"github.com/iikira/BaiduPCS-Go/requester"
	"net"
)

type (
	// IPResNetease 网易服务器获取ip返回的结果
	IPResNetease struct {
		Result  string `json:"result"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
)

// IPInfoFromNetease 从网易服务器获取ip
func IPInfoFromNetease() (ipAddr string, err error) {
	c := requester.NewHTTPClient()
	resp, err := c.Req("GET", "http://mam.netease.com/api/config/getClientIp", nil, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}

	res := &IPResNetease{}
	err = jsonhelper.UnmarshalData(resp.Body, res)
	if err != nil {
		return
	}

	ip := net.ParseIP(res.Result)
	if ip == nil {
		err = ErrParseIP
		return
	}

	ipAddr = res.Result
	return
}
