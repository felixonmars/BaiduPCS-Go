package baidupcs

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs/expires"
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
	"time"
)

type (
	filesDirectoriesListValidate struct {
		fds FileDirectoryList
		expires.Expires
	}
)

// updateFilesDirectoriesCache 更新缓存
func (pcs *BaiduPCS) updateFilesDirectoriesCache(dirs []string) {
	cache := pcs.cacheMap.LazyInitCachePoolOp(OperationFilesDirectoriesList)
	for _, v := range dirs {
		filesDirectoriesListValidateItf, ok := cache.Load(v + "_" + DefaultOrderOptionsStr)
		if ok {
			filesDirectoriesListValidateItf.(*filesDirectoriesListValidate).SetExpires(false)
		}
	}
}

func (pcs *BaiduPCS) CacheFilesDirectoriesList(path string, options *OrderOptions) (data FileDirectoryList, pcsError pcserror.Error) {
	var (
		cache                               = pcs.cacheMap.LazyInitCachePoolOp(OperationFilesDirectoriesList)
		key                                 = path + "_" + fmt.Sprint(options)
		filesDirectoriesListValidateItf, ok = cache.Load(key)
	)
	if !ok {
		data, pcsError = pcs.FilesDirectoriesList(path, options)
		if pcsError != nil {
			return
		}
		cache.Store(key, &filesDirectoriesListValidate{
			fds:     data,
			Expires: expires.NewExpires(1 * time.Minute),
		})
		return
	}
	return filesDirectoriesListValidateItf.(*filesDirectoriesListValidate).fds, nil
}
