package downloader

import (
	"errors"
	"github.com/golang/protobuf/proto"
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
		format   InstanceStateStorageFormat
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
		GetInstanceInfo() *InstanceInfo
		SetInstanceInfo(*InstanceInfo)
	}

	// InstanceStateStorageFormat 断点续传储存类型
	InstanceStateStorageFormat int
)

const (
	// InstanceStateStorageFormatJSON json 格式
	InstanceStateStorageFormatJSON = iota
	// InstanceStateStorageFormatProto3 protobuf 格式
	InstanceStateStorageFormatProto3
)

// GetInstanceInfo 从断点信息获取下载状态
func (m *InstanceInfoExport) GetInstanceInfo() (eii *InstanceInfo) {
	eii = &InstanceInfo{
		Ranges: m.Ranges,
	}

	var downloaded int64
	switch m.RangeGenMode {
	case RangeGenMode_BlockSize:
		downloaded = m.GenBegin - eii.Ranges.Len()
	default:
		downloaded = m.TotalSize - eii.Ranges.Len()
	}
	eii.DlStatus = &DownloadStatus{
		nowTime:          time.Now(),
		totalSize:        m.TotalSize,
		downloaded:       downloaded,
		speedsDownloaded: downloaded,
		oldDownloaded:    downloaded,
		gen:              NewRangeListGenBlockSize(m.TotalSize, m.GenBegin, m.BlockSize),
	}
	switch m.RangeGenMode {
	case RangeGenMode_BlockSize:
		eii.DlStatus.gen = NewRangeListGenBlockSize(m.TotalSize, m.GenBegin, m.BlockSize)
	default:
		eii.DlStatus.gen = NewRangeListGenDefault(m.TotalSize, m.TotalSize, len(m.Ranges), len(m.Ranges))
	}
	return eii
}

// SetInstanceInfo 从下载状态导出断点信息
func (m *InstanceInfoExport) SetInstanceInfo(eii *InstanceInfo) {
	if eii == nil {
		return
	}

	if eii.DlStatus != nil {
		m.TotalSize = eii.DlStatus.TotalSize()
		if eii.DlStatus.gen != nil {
			m.GenBegin = eii.DlStatus.gen.LoadBegin()
			m.BlockSize = eii.DlStatus.gen.LoadBlockSize()
			m.RangeGenMode = eii.DlStatus.gen.RangeGenMode()
		} else {
			m.RangeGenMode = RangeGenMode_Default
		}
	}
	m.Ranges = eii.Ranges
}

//NewInstanceState 初始化InstanceState
func NewInstanceState(saveFile *os.File, format InstanceStateStorageFormat) *InstanceState {
	return &InstanceState{
		saveFile: saveFile,
		format:   format,
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

	is.ii = &InstanceInfoExport{}
	var err error
	switch is.format {
	case InstanceStateStorageFormatProto3:
		err = proto.Unmarshal(contents, is.ii.(*InstanceInfoExport))
	default:
		err = jsoniter.Unmarshal(contents, is.ii)
	}

	if err != nil {
		pcsverbose.Verbosef("DEBUG: InstanceInfo unmarshal error: %s\n", err)
		return
	}

	eii = is.ii.GetInstanceInfo()
	return
}

//Put 提交信息
func (is *InstanceState) Put(eii *InstanceInfo) {
	if !is.checkSaveFile() {
		return
	}

	is.mu.Lock()
	defer is.mu.Unlock()

	if is.ii == nil {
		is.ii = &InstanceInfoExport{}
	}
	is.ii.SetInstanceInfo(eii)
	var (
		data []byte
		err  error
	)
	switch is.format {
	case InstanceStateStorageFormatProto3:
		data, err = proto.Marshal(is.ii.(*InstanceInfoExport))
	default:
		data, err = jsoniter.Marshal(is.ii)
	}
	if err != nil {
		panic(err)
	}

	err = is.saveFile.Truncate(int64(len(data)))
	if err != nil {
		pcsverbose.Verbosef("DEBUG: truncate file error: %s\n", err)
	}

	_, err = is.saveFile.WriteAt(data, 0)
	if err != nil {
		pcsverbose.Verbosef("DEBUG: write instance state error: %s\n", err)
	}
}

//Close 关闭
func (is *InstanceState) Close() error {
	if !is.checkSaveFile() {
		return nil
	}

	return is.saveFile.Close()
}

func (der *Downloader) initInstanceState(format InstanceStateStorageFormat) (err error) {
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

	der.instanceState = NewInstanceState(saveFile, format)
	return nil
}

func (der *Downloader) removeInstanceState() error {
	der.instanceState.Close()
	if !der.config.IsTest && der.config.InstanceStatePath != "" {
		return os.Remove(der.config.InstanceStatePath)
	}
	return nil
}
