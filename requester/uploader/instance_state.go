package uploader

import (
	"io"
)

type (
	// BlockState 文件区块信息
	BlockState struct {
		ID       int       `json:"id"`
		Range    ReadRange `json:"range"`
		CheckSum string    `json:"checksum"`
	}

	// InstanceState 上传断点续传信息
	InstanceState struct {
		BlockList []*BlockState `json:"block_list"`
	}
)

func workerListToInstanceState(workers workerList) *InstanceState {
	blockStates := make([]*BlockState, 0, len(workers))
	for _, wer := range workers {
		blockStates = append(blockStates, &BlockState{
			ID:       wer.id,
			Range:    wer.splitUnit.Range(),
			CheckSum: wer.checksum,
		})
	}
	return &InstanceState{
		BlockList: blockStates,
	}
}

func instanceStateToWorkerList(is *InstanceState, file io.ReaderAt) workerList {
	workers := make(workerList, 0, len(is.BlockList))
	for _, blockState := range is.BlockList {
		if blockState.CheckSum == "" {
			workers = append(workers, &worker{
				id:         blockState.ID,
				partOffset: blockState.Range.Begin,
				splitUnit:  NewBufioSplitUnit(file, blockState.Range),
				checksum:   blockState.CheckSum,
			})
		} else {
			workers = append(workers, &worker{
				id:         blockState.ID,
				partOffset: blockState.Range.Begin,
				splitUnit: &fileBlock{
					readRange: blockState.Range,
					readed:    blockState.Range.End - blockState.Range.Begin,
					readerAt:  file,
				},
				checksum: blockState.CheckSum,
			})
		}
	}
	return workers
}
