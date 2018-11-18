package dlinkclient

import (
	"github.com/iikira/Baidu-Login/bdcrypto"
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"net/url"
)

func (dc *DlinkClient) linkRedirect(op, link string) (nlink string, dlinkError pcserror.Error) {
	dc.lazyInit()

	var (
		u           *url.URL
		redirectRes = RedirectRes{
			DlinkErrInfo: pcserror.NewDlinkErrInfo(op),
		}
		uv = url.Values{}
	)

	switch op {
	case OperationRedirect:
		u = dc.genCgiBinURL("redirect", nil)
		uv.Set("link", link)
	case OperationRedirectPr:
		u = dc.genCgiBinURL("redirect/pr", nil)
		uv.Set("link", converter.ToString(bdcrypto.Base64Encode(converter.ToBytes(link))))
	}

	resp, err := dc.client.Req("POST", u.String(), uv.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		redirectRes.SetNetError(err)
		return "", redirectRes.DlinkErrInfo
	}

	dlinkError = pcserror.HandleJSONParse(op, resp.Body, &redirectRes)
	if dlinkError != nil {
		return
	}

	return redirectRes.Link, nil
}

func (dc *DlinkClient) LinkRedirect(link string) (nlink string, dlinkError pcserror.Error) {
	return dc.linkRedirect(OperationRedirect, link)
}

func (dc *DlinkClient) LinkRedirectPr(link string) (nlink string, dlinkError pcserror.Error) {
	return dc.linkRedirect(OperationRedirectPr, link)
}
