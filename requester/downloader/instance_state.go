package downloader

import (
	"errors"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/json-iterator/go"
	"os"
	"sync"
	"time"
)

type (
	//InstanceState 状态, 断点续传信息
	InstanceState struct {
		saveFile *os.File
		ii       InstanceInfoExporter
		mu       sync.Mutex
	}

	//InstanceInfo 状态详细信息, 用于导出状态文件
	InstanceInfo struct {
		DlStatus *DownloadStatus
		Ranges   RangeList
	}

	// InstanceInfoExporter 断点续传类型接口
	InstanceInfoExporter interface {
		Convert() *InstanceInfo
		Render(*InstanceInfo)
	}

	// instanceInfoExport 断点续传
	instanceInfoExport struct {
		RangeGenMode RangeGenMode `json:"mode"`
	}

	// instanceInfoExportDefault 断点续传用到的struct 1
	instanceInfoExportDefault struct {
		instanceInfoExport
		TotalSize int64     `json:"total_size"` // 总大小
		Ranges    RangeList `json:"ranges"`
	}

	// instanceInfoExportBlockSize 断点续传用到的struct 2
	instanceInfoExportBlockSize struct {
		instanceInfoExport
		TotalSize int64     `json:"total_size"` // 总大小
		GenBegin  int64     `json:"gen_begin"`
		BlockSize int64     `json:"block_size"`
		Ranges    RangeList `json:"ranges"`
	}
)

func (ii1 *instanceInfoExportDefault) Convert() (eii *InstanceInfo) {
	eii = &InstanceInfo{
		Ranges: ii1.Ranges,
	}

	downloaded := ii1.TotalSize - eii.Ranges.Len()
	eii.DlStatus = &DownloadStatus{
		nowTime:          time.Now(),
		totalSize:        ii1.TotalSize,
		downloaded:       downloaded,
		speedsDownloaded: downloaded,
		oldDownloaded:    downloaded,
		gen:              NewRangeListGenDefault(ii1.TotalSize, ii1.TotalSize, len(ii1.Ranges), len(ii1.Ranges)), // 无效gen
	}
	return eii
}

func (ii1 *instanceInfoExportDefault) Render(eii *InstanceInfo) {
	ii1.RangeGenMode = RangeGenModeDefault
	if eii == nil {
		return
	}
	if eii.DlStatus != nil {
		ii1.TotalSize = eii.DlStatus.TotalSize()
	}
	ii1.Ranges = eii.Ranges
}

func (ii2 *instanceInfoExportBlockSize) Convert() (eii *InstanceInfo) {
	eii = &InstanceInfo{
		Ranges: ii2.Ranges,
	}

	downloaded := ii2.GenBegin - eii.Ranges.Len()
	eii.DlStatus = &DownloadStatus{
		nowTime:          time.Now(),
		totalSize:        ii2.TotalSize,
		downloaded:       downloaded,
		speedsDownloaded: downloaded,
		oldDownloaded:    downloaded,
		gen:              NewRangeListGenBlockSize(ii2.TotalSize, ii2.GenBegin, ii2.BlockSize),
	}
	return eii
}

func (ii2 *instanceInfoExportBlockSize) Render(eii *InstanceInfo) {
	ii2.RangeGenMode = RangeGenModeBlockSize
	if eii == nil {
		return
	}
	if eii.DlStatus != nil {
		ii2.TotalSize = eii.DlStatus.TotalSize()
	}
	ii2.GenBegin = eii.DlStatus.gen.LoadBegin()
	ii2.BlockSize = eii.DlStatus.gen.LoadBlockSize()
	ii2.Ranges = eii.Ranges
}

//NewInstanceState 初始化InstanceState
func NewInstanceState(saveFile *os.File) *InstanceState {
	return &InstanceState{
		saveFile: saveFile,
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

	buf := make([]byte, intSize)

	n, _ := is.saveFile.ReadAt(buf, 0)
	return buf[:n]
}

//Get 拉取信息
func (is *InstanceState) Get() (eii *InstanceInfo) {
	if !is.checkSaveFile() {
		return nil
	}

	is.mu.Lock()
	defer is.mu.Unlock()

	contents := is.getSaveFileContents()
	if len(contents) <= 0 {
		return
	}

	iiBase := instanceInfoExport{}
	err := jsoniter.Unmarshal(contents, &iiBase)
	if err != nil {
		pcsverbose.Verbosef("DEBUG: [base] InstanceInfo unmarshal json error: %s\n", err)
		return
	}

	switch iiBase.RangeGenMode {
	case RangeGenModeBlockSize:
		is.ii = &instanceInfoExportBlockSize{}
	default:
		is.ii = &instanceInfoExportDefault{}
	}

	err = jsoniter.Unmarshal(contents, is.ii)
	if err != nil {
		pcsverbose.Verbosef("DEBUG: [%d] InstanceInfo unmarshal json error: %s\n", iiBase.RangeGenMode, err)
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

	if eii.DlStatus.gen == nil {
		is.ii = &instanceInfoExportDefault{}
	} else {
		switch eii.DlStatus.gen.Mode() {
		case RangeGenModeBlockSize:
			is.ii = &instanceInfoExportBlockSize{}
		default:
			is.ii = &instanceInfoExportDefault{}
		}
	}

	is.mu.Lock()
	defer is.mu.Unlock()

	var err error

	is.ii.Render(eii)
	data, err := jsoniter.Marshal(is.ii)
	if err != nil {
		panic(err)
	}

	err = is.saveFile.Truncate(int64(len(data)))
	if err != nil {
		pcsverbose.Verbosef("DEBUG: truncate file error: %s\n", err)
	}

	_, err = is.saveFile.WriteAt(data, 0)
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
