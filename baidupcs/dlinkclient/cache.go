package dlinkclient

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs/expires"
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
	"strconv"
	"time"
)

type (
	regValidate struct {
		short string
		expires.Expires
	}

	listValidate struct {
		fds []*FileDirectory
		expires.Expires
	}

	linkRedirectValidate struct {
		nlink string
		expires.Expires
	}
)

func (dc *DlinkClient) CacheShareReg(shareURL, passwd string) (short string, dlinkError pcserror.Error) {
	var (
		cache                = dc.cacheMap.LazyInitCachePoolOp(OperationReg)
		key                  = shareURL + "_" + passwd
		shortValidateItf, ok = cache.Load(key)
	)
	if !ok {
		short, dlinkError = dc.ShareReg(shareURL, passwd)
		if dlinkError != nil {
			return
		}
		cache.Store(key, &regValidate{
			short:   short,
			Expires: expires.NewExpires(10 * time.Minute),
		})
		return
	}
	return shortValidateItf.(*regValidate).short, nil
}

func (dc *DlinkClient) CacheShareList(short, dir string, page int) (fds []*FileDirectory, dlinkError pcserror.Error) {
	var (
		cache               = dc.cacheMap.LazyInitCachePoolOp(OperationList)
		key                 = short + "_" + dir + "_" + strconv.Itoa(page)
		listValidateItf, ok = cache.Load(key)
	)
	if !ok {
		fds, dlinkError = dc.ShareList(short, dir, page)
		if dlinkError != nil {
			return
		}
		cache.Store(key, &listValidate{
			fds:     fds,
			Expires: expires.NewExpires(1 * time.Minute),
		})
		return
	}
	return listValidateItf.(*listValidate).fds, nil
}

func (dc *DlinkClient) cacheLinkRedirect(op string, link string) (nlink string, dlinkError pcserror.Error) {
	var (
		cache                       = dc.cacheMap.LazyInitCachePoolOp(op)
		linkRedirectValidateItf, ok = cache.Load(link)
	)
	if !ok {
		nlink, dlinkError = dc.linkRedirect(op, link)
		if dlinkError != nil {
			return
		}
		cache.Store(link, &linkRedirectValidate{
			nlink:   nlink,
			Expires: expires.NewExpires(2 * time.Hour),
		})
		return
	}
	return linkRedirectValidateItf.(*linkRedirectValidate).nlink, nil
}

func (dc *DlinkClient) CacheLinkRedirect(link string) (nlink string, dlinkError pcserror.Error) {
	return dc.cacheLinkRedirect(OperationRedirect, link)
}
func (dc *DlinkClient) CacheLinkRedirectPr(link string) (nlink string, dlinkError pcserror.Error) {
	return dc.cacheLinkRedirect(OperationRedirectPr, link)
}
