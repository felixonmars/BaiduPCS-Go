// 控制bg，fg命令时的终端输出信息（不包含DEBUG和ERROR）
package downloader

import (
	"sync"
	"fmt"
)

// OutputController 控制终端输出
type OutputController struct {
	*sync.Mutex
	// 终端输出开关
	trigger bool
}

func NewOutputController() *OutputController {
	o := new(OutputController)
	o.Mutex = &sync.Mutex{}
	o.trigger = true
	
	return o
}

func (o *OutputController)Printf(args ...interface{}) {
	o.Lock()
	defer o.Unlock()
	
	if !o.trigger {
		fmt.Printf(args[0].(string), args[1:]...)
	}
}

func (o *OutputController)Println(args ...interface{}) {
	o.Lock()
	defer o.Unlock()
	
	if !o.trigger {
		fmt.Println(args...)
	}
}

// SetTrigger 设置是否输出信息
func (o *OutputController)SetTrigger(b bool) {
	o.Lock()
	defer o.Unlock()
	
	o.trigger = b
}
