package downloader

import (
	"encoding/json"
	"io/ioutil"
)

var (
	downloadingFileSuffix = ".baidupcs_go_downloading"
)

type downloadStatus struct {
	Downloaded int64     `json:"downloaded"`
	BlockList  blockList `json:"block_list"`
}

// recordBreakPoint 保存下载断点到文件, 用于断点续传
func (f *FileDl) recordBreakPoint() error {
	byt, err := json.Marshal(downloadStatus{
		Downloaded: f.status.Downloaded,
		BlockList:  f.BlockList,
	})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(f.File.Name()+downloadingFileSuffix, byt, 0644)
}

// loadBreakPoint 尝试从文件载入下载断点
func (f *FileDl) loadBreakPoint() error {
	byt, err := ioutil.ReadFile(f.File.Name() + downloadingFileSuffix)
	if err != nil {
		return err
	}
	downloadStatus := new(downloadStatus)
	err = json.Unmarshal(byt, downloadStatus)
	if err != nil {
		return err
	}
	f.status.Downloaded = downloadStatus.Downloaded
	f.BlockList = downloadStatus.BlockList
	return nil
}
