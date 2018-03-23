package pcscommand

import (
	"fmt"
)

// RunCloudDlAddTask 执行添加离线任务
func RunCloudDlAddTask(sourceURLs []string, savePath string) {
	var err error
	savePath, err = getAbsPath(savePath)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	var taskid int64
	for k := range sourceURLs {
		taskid, err = info.CloudDlAddTask(sourceURLs[k], savePath)
		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		fmt.Printf("[%d] 添加离线任务成功, taskid: %d, 源地址: %s, 保存路径: %s\n", k+1, taskid, sourceURLs[k], savePath)
	}
}
