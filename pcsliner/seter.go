package pcsliner

import (
	"fmt"
	"os"
)

// resetPCSLiner 重置 PCSLiner
func resetPCSLiner(oldLiner *PCSLiner) (newLiner *PCSLiner) {
	newLiner = NewLiner()

	if oldLiner == nil {
		return
	}

	newLiner.Config = oldLiner.Config

	// 重新设置历史
	if newLiner.Config.historyFile != nil {
		newLiner.SetHistory(newLiner.Config.historyFile.Name())
	}

	newLiner.State.SetCtrlCAborts(newLiner.Config.CtrlCAborts)
	newLiner.State.SetCompleter(newLiner.Config.mainCompleter)

	oldLiner.Config.historyFile.Close()
	return
}

// SetHistory 设置历史记录保存文件
func (pl *PCSLiner) SetHistory(filePath string) (err error) {
	if filePath == "" {
		return fmt.Errorf("history file not set")
	}

	pl.Config.historyFile, err = os.Open(filePath)
	if err != nil {
		return err
	}

	_, err = pl.State.ReadHistory(pl.Config.historyFile)
	return err
}

func (pl *PCSLiner) SetMainCompleter(mc func(line string) []string) {
	pl.Config.mainCompleter = mc
	pl.State.SetCompleter(mc)
}
