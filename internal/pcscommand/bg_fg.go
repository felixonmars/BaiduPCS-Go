package pcscommand

import (
	"sync"
	"fmt"
	"os"
	"time"
	"strings"
	"crypto/md5"
	
	"github.com/iikira/BaiduPCS-Go/requester/downloader"
	"github.com/iikira/BaiduPCS-Go/pcstable"
)

var (
	bgMap = bgTasks{
		tasks: sync.Map{},
		ticker: time.NewTicker(1*time.Minute),
	}
)

func init() {
	bgMap.cycleCheck()
}

type bgTasks struct {
	tasks sync.Map
	ticker *time.Ticker
	checkStarted bool
}

func (b *bgTasks)checkDoneTask() {
	b.tasks.Range(func(id, v interface{}) bool {
		task := v.(*bgDTaskItem)
		select {
		case <- task.Done:
			fmt.Printf("任务：%v 已完成\n", id.(string))
			b.tasks.Delete(id)
			return true
		default:
			return true
		}
	})
}

// 周期性检查是否有后台任务完成
func (b *bgTasks)cycleCheck() {
	if b.checkStarted {
		return
	}
	
	b.checkStarted = true
	go func() {
		for {
			select {
			case <- b.ticker.C:
				b.checkDoneTask()
			}
		}
	}()
}

type bgDTaskItem struct {
	paths []string
	outputcontrol *downloader.OutputController
	Done chan struct{}
}

// TaskID 以paths生成md5，取前10位做taskID
func (t *bgDTaskItem)TaskID() string {
	data := strings.Join(t.paths, "")
	has := md5.Sum([]byte(data))
	id := fmt.Sprintf("%x", has)
	return string(id[:10])
}

func PrintAllBgTask() {
	tb := pcstable.NewTable(os.Stdout)
	tb.SetHeader([]string{"task_id", "downloading files"})
	bgMap.tasks.Range(func(id, v interface{}) bool {
		tb.Append([]string{id.(string), strings.Join(v.(*bgDTaskItem).paths, ",")})
		return true
	})
	tb.Render()
}

func RunBgDownload(paths []string, option DownloadOption) {
	task := new(bgDTaskItem)
	task.paths = make([]string, 0, len(paths))
	task.paths = append(task.paths, paths...)
	task.outputcontrol = downloader.NewOutputController()
	task.Done = make(chan struct{})
	
	taskID := task.TaskID()
	_, exists := bgMap.tasks.Load(taskID)
	if exists {
		fmt.Printf("下载任务 ID：%v 已在后台下载任务中\n", taskID)
		return
	}
	
	bgMap.tasks.Store(taskID, task)
	
	option.BanOutput = task.outputcontrol
	go func(done chan struct{}) {
		RunDownload(paths, option, taskID)
		close(done)
	}(task.Done)
}

func RunFgDownload(taskID string) {
	t, ok := bgMap.tasks.Load(taskID)
	if !ok {
		fmt.Printf("任务：%v 不存在\n", taskID)
		return
	}
	
	task := t.(*bgDTaskItem)
	bgMap.checkDoneTask()
	task.outputcontrol.SetTrigger(false)
}

