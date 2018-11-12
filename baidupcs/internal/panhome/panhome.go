package panhome

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs/expires"
	"github.com/iikira/BaiduPCS-Go/requester"
	"net/url"
)

const (
	// OperationSignature signature
	OperationSignature = "signature"
)

var (
	panBaiduComURL = &url.URL{
		Scheme: "https",
		Host:   "pan.baidu.com",
	}
	AndroidUserAgent = "Mozilla/5.0 (Linux; Android 7.0; HUAWEI NXT-AL10 Build/HUAWEINXT-AL10) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.137 Mobile Safari/537.36"
)

type (
	PanHome struct {
		client *requester.HTTPClient
		ua     string
		bduss  string

		sign1, sign3 []rune
		timestamp    string

		signRes     SignRes
		signExpires expires.Expires
	}
)

func NewPanHome(client *requester.HTTPClient) *PanHome {
	ph := PanHome{}
	if client != nil {
		newC := *client
		ph.client = &newC
	}
	return &ph
}

func (ph *PanHome) lazyInit() {
	if ph.client == nil {
		ph.client = requester.NewHTTPClient()
	}
}
