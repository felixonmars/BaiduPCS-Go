package baidupcs

import (
	"fmt"
	"github.com/json-iterator/go"
)

// CloudDlAddTask 添加离线下载任务
func (pcs *BaiduPCS) CloudDlAddTask(sourceURL, savePath string) (taskid int64, err error) {
	dataReadCloser, err := pcs.PrepareCloudDlAddTask(sourceURL, savePath)
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	taskInfo := &struct {
		TaskID int64 `json:"task_id"`
		ErrInfo
	}{}

	d := jsoniter.NewDecoder(dataReadCloser)
	err = d.Decode(taskInfo)
	if err != nil {
		return 0, fmt.Errorf("%s, %s, %s", operationCloudDlAddTask, StrJSONParseError, err)
	}

	if taskInfo.ErrCode != 0 {
		return 0, &taskInfo.ErrInfo
	}

	return taskInfo.TaskID, nil
}
