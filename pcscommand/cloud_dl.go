package pcscommand

import (
	"fmt"
)

// RunCloudDlAddTask 执行添加离线任务
func RunCloudDlAddTask(sourceURL, savePath string) {
	taskid, err := info.CloudDlAddTask(sourceURL, savePath)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	fmt.Printf("添加离线任务成功, taskid: %d, 源地址: %s, 保存路径: %s\n", taskid, sourceURL, savePath)
}
