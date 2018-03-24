package baidupcs

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/json-iterator/go"
	"strconv"
	"strings"
)

// CloudDlFileInfo 离线下载的文件信息
type CloudDlFileInfo struct {
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
}

// CloudDlTaskInfo 离线下载的任务信息
type CloudDlTaskInfo struct {
	TaskID       int64
	Status       int // 0下载成功, 1下载进行中, 2系统错误, 3资源不存在, 4下载超时, 5资源存在但下载失败, 6存储空间不足, 7任务取消
	StatusText   string
	FileSize     int64  // 文件大小
	FinishedSize int64  // 文件大小
	CreateTime   int64  // 创建时间
	StartTime    int64  // 开始时间
	FinishTime   int64  // 结束时间
	SavePath     string // 保存的路径
	SourceURL    string // 资源地址
	TaskName     string // 任务名称, 一般为文件名
	OdType       int
	FileList     []*CloudDlFileInfo
	Result       int // 0查询成功，结果有效，1要查询的task_id不存在
}

// CloudDlTaskList 离线下载的任务信息列表
type CloudDlTaskList []*CloudDlTaskInfo

// cloudDlTaskInfo 用于解析远程返回的JSON
type cloudDlTaskInfo struct {
	Status       string `json:"status"`
	FileSize     string `json:"file_size"`
	FinishedSize string `json:"finished_size"`
	CreateTime   string `json:"create_time"`
	StartTime    string `json:"start_time"`
	FinishTime   string `json:"finish_time"`
	SavePath     string `json:"save_path"`
	SourceURL    string `json:"source_url"`
	TaskName     string `json:"task_name"`
	OdType       string `json:"od_type"`
	FileList     []*struct {
		FileName string `json:"file_name"`
		FileSize string `json:"file_size"`
	} `json:"file_list"`
	Result int `json:"result"`
}

func (ci *cloudDlTaskInfo) convert() *CloudDlTaskInfo {
	ci2 := &CloudDlTaskInfo{
		Status:       pcsutil.MustInt(ci.Status),
		FileSize:     pcsutil.MustInt64(ci.FileSize),
		FinishedSize: pcsutil.MustInt64(ci.FinishedSize),
		CreateTime:   pcsutil.MustInt64(ci.CreateTime),
		StartTime:    pcsutil.MustInt64(ci.StartTime),
		FinishTime:   pcsutil.MustInt64(ci.FinishTime),
		SavePath:     ci.SavePath,
		SourceURL:    ci.SourceURL,
		TaskName:     ci.TaskName,
		OdType:       pcsutil.MustInt(ci.OdType),
		Result:       ci.Result,
	}

	ci2.FileList = make([]*CloudDlFileInfo, 0, len(ci.FileList))
	for _, v := range ci.FileList {
		if v == nil {
			continue
		}

		ci2.FileList = append(ci2.FileList, &CloudDlFileInfo{
			FileName: v.FileName,
			FileSize: pcsutil.MustInt64(v.FileSize),
		})
	}

	return ci2
}

// CloudDlAddTask 添加离线下载任务
func (pcs *BaiduPCS) CloudDlAddTask(sourceURL, savePath string) (taskID int64, err error) {
	dataReadCloser, err := pcs.PrepareCloudDlAddTask(sourceURL, savePath)
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	taskInfo := &struct {
		TaskID int64 `json:"task_id"`
		*ErrInfo
	}{
		ErrInfo: NewErrorInfo(OperationCloudDlAddTask),
	}

	d := jsoniter.NewDecoder(dataReadCloser)
	err = d.Decode(taskInfo)
	if err != nil {
		return 0, fmt.Errorf("%s, %s, %s", OperationCloudDlAddTask, StrJSONParseError, err)
	}

	if taskInfo.ErrCode != 0 {
		return 0, taskInfo.ErrInfo
	}

	return taskInfo.TaskID, nil
}

// CloudDlQueryTask 精确查询离线下载任务
func (pcs *BaiduPCS) CloudDlQueryTask(taskIDs []int64) (cl CloudDlTaskList, err error) {
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("%s, no input any task_ids", OperationCloudDlQueryTask)
	}

	taskStrIDs := make([]string, len(taskIDs))
	for k := range taskStrIDs {
		taskStrIDs[k] = strconv.FormatInt(taskIDs[k], 10)
	}

	dataReadCloser, err := pcs.PrepareCloudDlQueryTask(strings.Join(taskStrIDs, ","))
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	taskInfo := &struct {
		TaskInfo map[string]*cloudDlTaskInfo `json:"task_info"`
		*ErrInfo
	}{
		ErrInfo: NewErrorInfo(OperationCloudDlQueryTask),
	}

	d := jsoniter.NewDecoder(dataReadCloser)
	err = d.Decode(taskInfo)
	if err != nil {
		return nil, fmt.Errorf("%s, %s, %s", OperationCloudDlQueryTask, StrJSONParseError, err)
	}

	if taskInfo.ErrCode != 0 {
		return nil, taskInfo.ErrInfo
	}

	var v2 *CloudDlTaskInfo
	cl = make(CloudDlTaskList, 0, len(taskStrIDs))
	for k := range taskStrIDs {
		v := taskInfo.TaskInfo[taskStrIDs[k]]
		if v == nil {
			continue
		}

		v2 = v.convert()

		v2.TaskID, err = strconv.ParseInt(taskStrIDs[k], 10, 64)
		if err != nil {
			continue
		}

		v2.ParseText()
		cl = append(cl, v2)
	}

	return cl, nil
}

// CloudDlListTask 查询离线下载任务列表
func (pcs *BaiduPCS) CloudDlListTask() (cl CloudDlTaskList, err error) {
	dataReadCloser, err := pcs.PrepareCloudDlListTask()
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	taskInfo := &struct {
		TaskInfo []*struct {
			TaskID string `json:"task_id"`
		} `json:"task_info"`
		*ErrInfo
	}{
		ErrInfo: NewErrorInfo(OperationCloudDlListTask),
	}

	d := jsoniter.NewDecoder(dataReadCloser)
	err = d.Decode(taskInfo)
	if err != nil {
		return nil, fmt.Errorf("%s, %s, %s", OperationCloudDlListTask, StrJSONParseError, err)
	}

	if taskInfo.ErrCode != 0 {
		return nil, taskInfo.ErrInfo
	}

	var (
		taskID  int64
		taskIDs = make([]int64, 0, len(taskInfo.TaskInfo))
	)
	for _, v := range taskInfo.TaskInfo {
		if v == nil {
			continue
		}

		if taskID, err = strconv.ParseInt(v.TaskID, 10, 64); err == nil {
			taskIDs = append(taskIDs, taskID)
		}
	}

	return pcs.CloudDlQueryTask(taskIDs)
}

// CloudDlCancelTask 取消离线下载任务
func (pcs *BaiduPCS) CloudDlCancelTask(taskID int64) (err error) {
	dataReadCloser, err := pcs.PrepareCloudDlCancelTask(taskID)
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := NewErrorInfo(OperationCloudDlCancelTask)

	d := jsoniter.NewDecoder(dataReadCloser)
	err = d.Decode(errInfo)
	if err != nil {
		return fmt.Errorf("%s, %s, %s", OperationCloudDlCancelTask, StrJSONParseError, err)
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}

	return nil
}

// CloudDlDeleteTask 删除离线下载任务
func (pcs *BaiduPCS) CloudDlDeleteTask(taskID int64) (err error) {
	dataReadCloser, err := pcs.PrepareCloudDlDeleteTask(taskID)
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := NewErrorInfo(OperationCloudDlDeleteTask)

	d := jsoniter.NewDecoder(dataReadCloser)
	err = d.Decode(errInfo)
	if err != nil {
		return fmt.Errorf("%s, %s, %s", OperationCloudDlDeleteTask, StrJSONParseError, err)
	}

	if errInfo.ErrCode != 0 {
		return errInfo
	}

	return nil
}

// ParseText 解析状态码
func (ci *CloudDlTaskInfo) ParseText() {
	switch ci.Status {
	case 0:
		ci.StatusText = "下载成功"
	case 1:
		ci.StatusText = "下载进行中"
	case 2:
		ci.StatusText = "系统错误"
	case 3:
		ci.StatusText = "资源不存在"
	case 4:
		ci.StatusText = "下载超时"
	case 5:
		ci.StatusText = "资源存在但下载失败"
	case 6:
		ci.StatusText = "存储空间不足"
	case 7:
		ci.StatusText = "任务取消"
	default:
		ci.StatusText = "未知状态码: " + strconv.Itoa(ci.Status)
	}
}

func (cl CloudDlTaskList) String() string {
	builder := &strings.Builder{}
	tb := pcstable.NewTable(builder)
	tb.SetHeader([]string{"#", "任务ID", "任务名称", "文件大小", "创建日期", "保存路径", "资源地址", "状态"})
	for k, v := range cl {
		tb.Append([]string{strconv.Itoa(k), strconv.FormatInt(v.TaskID, 10), v.TaskName, pcsutil.ConvertFileSize(v.FileSize), pcsutil.FormatTime(v.CreateTime), v.SavePath, v.SourceURL, v.StatusText})
	}
	tb.Render()
	return builder.String()
}
