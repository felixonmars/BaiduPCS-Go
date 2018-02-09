package pcsliner

import (
	"fmt"
	"github.com/iikira/liner"
	"os"
)

// PCSLiner 封装 *liner.State, 提供更简便的操作
type PCSLiner struct {
	State  *liner.State
	Config *linerConfig

	paused bool
}

type linerConfig struct {
	CtrlCAborts bool

	mainCompleter func(line string) []string // 主命令自动补全
	historyFile   *os.File
}

// NewLiner 返回 *PCSLiner, 默认设置允许 Ctrl+C 结束
func NewLiner() *PCSLiner {
	pcsliner := liner.NewLiner()

	pl := &PCSLiner{
		State: pcsliner,
		Config: &linerConfig{
			CtrlCAborts: true,
		},
	}

	pcsliner.SetCtrlCAborts(pl.Config.CtrlCAborts)

	return pl
}

// DoWriteHistory 执行写入历史
func (pl *PCSLiner) DoWriteHistory() error {
	if pl.Config.historyFile == nil {
		return fmt.Errorf("history file not set")
	}

	pl.Config.historyFile, _ = os.Create(pl.Config.historyFile.Name())
	_, err := pl.State.WriteHistory(pl.Config.historyFile)
	if err != nil {
		return fmt.Errorf("Error writing history file: %s", err)
	}

	return nil
}

// Pause 暂停服务
func (pl *PCSLiner) Pause() error {
	if pl.paused {
		panic("PCSLiner already paused")
	}

	pl.paused = true
	pl.DoWriteHistory()
	return pl.State.Close()
}

// Resume 恢复服务
func (pl *PCSLiner) Resume() {
	if !pl.paused {
		panic("PCSLiner is not paused")
	}

	pl.paused = false

	*pl = *resetPCSLiner(pl) // 拷贝
}

// Close 关闭服务
func (pl *PCSLiner) Close() error {
	pl.DoWriteHistory()
	pl.Config.historyFile.Close()
	return pl.State.Close()
}
