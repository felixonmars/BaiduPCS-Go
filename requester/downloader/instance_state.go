package downloader

import (
	"errors"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/json-iterator/go"
	"os"
	"sync"
)

//InstanceState 状态, 断点续传信息
type InstanceState struct {
	saveFile *os.File
	buf      []byte
	ii       *instanceInfo
	mu       sync.Mutex
}

type rangeInfo struct {
	Begin int64 `json:"begin"`
	End   int64 `json:"end"`
}

//InstanceInfo 状态详细信息, 用于导出状态文件
type InstanceInfo struct {
	DlStatus *DownloadStatus
	Ranges   RangeList
}

type instanceInfo struct {
	TotalSize int64        `json:"total_size"` // 总大小
	Ranges    []*rangeInfo `json:"ranges"`
}

func (ii *instanceInfo) Convert() (eii *InstanceInfo) {
	eii = &InstanceInfo{}
	eii.Ranges = make([]*Range, 0, len(ii.Ranges))
	for k := range ii.Ranges {
		if ii.Ranges[k] == nil {
			continue
		}

		eii.Ranges = append(eii.Ranges, &Range{
			Begin: ii.Ranges[k].Begin,
			End:   ii.Ranges[k].End,
		})
	}

	downloaded := ii.TotalSize - eii.Ranges.Len()
	eii.DlStatus = &DownloadStatus{
		totalSize:        ii.TotalSize,
		downloaded:       downloaded,
		speedsDownloaded: downloaded,
		oldDownloaded:    downloaded,
	}
	return eii
}

func (ii *instanceInfo) Render(eii *InstanceInfo) {
	if eii == nil {
		return
	}
	if eii.DlStatus != nil {
		ii.TotalSize = eii.DlStatus.TotalSize()
	}
	if eii.Ranges != nil {
		ii.Ranges = make([]*rangeInfo, 0, len(eii.Ranges))
		for k := range eii.Ranges {
			if eii.Ranges[k] == nil {
				continue
			}

			ii.Ranges = append(ii.Ranges, &rangeInfo{
				Begin: eii.Ranges[k].LoadBegin(),
				End:   eii.Ranges[k].LoadEnd(),
			})
		}
	}
}

//NewInstanceState 初始化InstanceState
func NewInstanceState(saveFile *os.File) *InstanceState {
	return &InstanceState{
		saveFile: saveFile,
		ii:       &instanceInfo{},
	}
}

func (is *InstanceState) checkSaveFile() bool {
	return is.saveFile != nil
}

func (is *InstanceState) getSaveFileContents() []byte {
	if !is.checkSaveFile() {
		return nil
	}

	finfo, err := is.saveFile.Stat()
	if err != nil {
		panic(err)
	}

	size := finfo.Size()
	if size > 0xffffffff {
		panic("savePath too large")
	}
	intSize := int(size)

	if is.buf != nil {
		is.buf = make([]byte, intSize)
	}

	capacity := len(is.buf)
	if intSize > capacity {
		is.buf = append(is.buf, make([]byte, intSize-capacity)...)
	}

	n, _ := is.saveFile.ReadAt(is.buf, 0)
	return is.buf[:n]
}

//Get 拉取信息
func (is *InstanceState) Get() (eii *InstanceInfo) {
	if !is.checkSaveFile() {
		return nil
	}

	is.mu.Lock()
	defer is.mu.Unlock()

	if is.ii == nil {
		is.ii = &instanceInfo{}
	}

	contents := is.getSaveFileContents()
	if len(contents) <= 0 {
		return
	}

	err := jsoniter.Unmarshal(contents, is.ii)
	if err != nil {
		pcsverbose.Verbosef("DEBUG: unmarshal json error: %s\n", err)
		return
	}

	eii = is.ii.Convert()
	return
}

//Put 提交信息
func (is *InstanceState) Put(eii *InstanceInfo) {
	if !is.checkSaveFile() {
		return
	}

	if is.ii == nil { // 忽略
		return
	}

	is.mu.Lock()
	defer is.mu.Unlock()

	var err error

	is.ii.Render(eii)
	is.buf, err = jsoniter.Marshal(is.ii)
	if err != nil {
		panic(err)
	}

	_, err = is.saveFile.WriteAt(is.buf, 0)
	if err != nil {
		pcsverbose.Verbosef("DEBUG: write json error: %s\n", err)
	}
}

//Close 关闭
func (is *InstanceState) Close() error {
	if !is.checkSaveFile() {
		return nil
	}

	return is.saveFile.Close()
}

func (der *Downloader) initInstanceState() (err error) {
	if der.instanceState != nil {
		return errors.New("already initInstanceState")
	}

	var saveFile *os.File
	if !der.config.IsTest && der.config.InstanceStatePath != "" {
		saveFile, err = os.OpenFile(der.config.InstanceStatePath, os.O_RDWR|os.O_CREATE, 0777)
		if err != nil {
			return err
		}
	}

	der.instanceState = NewInstanceState(saveFile)
	return nil
}

func (der *Downloader) removeInstanceState() error {
	der.instanceState.Close()
	if !der.config.IsTest && der.config.InstanceStatePath != "" {
		return os.Remove(der.config.InstanceStatePath)
	}
	return nil
}
