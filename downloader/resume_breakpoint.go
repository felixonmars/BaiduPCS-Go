package downloader

import (
	"encoding/json"
	"io/ioutil"
)

var (
	// DownloadingFileSuffix 断点续传临时文件后缀
	DownloadingFileSuffix = ".baidupcs_go_downloading"
)

type downloadStatus struct {
	Downloaded int64     `json:"downloaded"`
	BlockList  blockList `json:"block_list"`
}

// recordBreakPoint 保存下载断点到文件, 用于断点续传
func (der *Downloader) recordBreakPoint() error {
	byt, err := json.Marshal(downloadStatus{
		Downloaded: der.status.Downloaded,
		BlockList:  der.BlockList,
	})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(der.File.Name()+DownloadingFileSuffix, byt, 0644)
}

// loadBreakPoint 尝试从文件载入下载断点
func (der *Downloader) loadBreakPoint() error {
	byt, err := ioutil.ReadFile(der.File.Name() + DownloadingFileSuffix)
	if err != nil {
		return err
	}
	downloadStatus := new(downloadStatus)
	err = json.Unmarshal(byt, downloadStatus)
	if err != nil {
		return err
	}
	der.status.Downloaded = downloadStatus.Downloaded
	der.BlockList = downloadStatus.BlockList
	return nil
}
