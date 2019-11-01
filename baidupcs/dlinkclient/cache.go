package dlinkclient

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs/expires"
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
	"strconv"
	"time"
)

func (dc *DlinkClient) CacheShareReg(shareURL, passwd string) (short string, dlinkError pcserror.Error) {
	data := dc.cacheOpMap.CacheOperation(OperationReg, shareURL+"_"+passwd, func() expires.DataExpires {
		short, dlinkError = dc.ShareReg(shareURL, passwd)
		if dlinkError != nil {
			return nil
		}
		return expires.NewDataExpires(short, 10*time.Minute)
	})
	if dlinkError != nil {
		return
	}
	return data.Data().(string), nil
}

func (dc *DlinkClient) CacheShareList(short, dir string, page int) (fds []*FileDirectory, dlinkError pcserror.Error) {
	data := dc.cacheOpMap.CacheOperation(OperationList, short+"_"+dir+"_"+strconv.Itoa(page), func() expires.DataExpires {
		fds, dlinkError = dc.ShareList(short, dir, page)
		if dlinkError != nil {
			return nil
		}
		return expires.NewDataExpires(fds, 1*time.Minute)
	})
	if dlinkError != nil {
		return
	}
	return data.Data().([]*FileDirectory), nil
}

func (dc *DlinkClient) cacheLinkRedirect(op string, link string) (nlink string, dlinkError pcserror.Error) {
	data := dc.cacheOpMap.CacheOperation(OperationList, link, func() expires.DataExpires {
		nlink, dlinkError = dc.linkRedirect(op, link)
		if dlinkError != nil {
			return nil
		}
		return expires.NewDataExpires(nlink, 2*time.Minute)
	})
	if dlinkError != nil {
		return
	}
	return data.Data().(string), nil
}

func (dc *DlinkClient) CacheLinkRedirect(link string) (nlink string, dlinkError pcserror.Error) {
	return dc.cacheLinkRedirect(OperationRedirect, link)
}
func (dc *DlinkClient) CacheLinkRedirectPr(link string) (nlink string, dlinkError pcserror.Error) {
	return dc.cacheLinkRedirect(OperationRedirectPr, link)
}
