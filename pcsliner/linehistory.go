package pcsliner

import (
	"fmt"
	"os"
)

type lineHistory struct {
	historyFilePath string
	historyFile     *os.File
}

func NewLineHistory(filePath string) (lh *lineHistory, err error) {
	lh = &lineHistory{
		historyFilePath: filePath,
	}

	lh.historyFile, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return lh, nil
}

// DoWriteHistory 执行写入历史
func (pl *PCSLiner) DoWriteHistory() (err error) {
	if pl.History == nil {
		return fmt.Errorf("history not set")
	}

	pl.History.historyFile, err = os.Create(pl.History.historyFilePath)
	if err != nil {
		return fmt.Errorf("写入历史错误, %s", err)
	}

	_, err = pl.State.WriteHistory(pl.History.historyFile)
	if err != nil {
		return fmt.Errorf("写入历史错误: %s", err)
	}

	return nil
}

// SetHistory 读取历史
func (pl *PCSLiner) ReadHistory() (err error) {
	if pl.History == nil {
		return fmt.Errorf("history not set")
	}

	_, err = pl.State.ReadHistory(pl.History.historyFile)
	return err
}
