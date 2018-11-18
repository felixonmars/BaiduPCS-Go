package uploader

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
)

type (
	// SplitUnit 将 io.ReaderAt 分割单元
	SplitUnit interface {
		Readed64
		io.Seeker
		Range() ReadRange
		Left() int64
	}

	// ReadRange 读取io.ReaderAt范围
	ReadRange struct {
		Begin int64 `json:"begin"`
		End   int64 `json:"end"`
	}

	fileBlock struct {
		readRange ReadRange
		readed    int64
		readerAt  io.ReaderAt
		mu        sync.Mutex
	}

	bufioFileBlock struct {
		*fileBlock
		bufio *bufio.Reader
	}
)

// SplitBlock 文件分块
func SplitBlock(fileSize, blockSize int64) (blockList []*BlockState) {
	blocksNum := int(fileSize / blockSize)
	if fileSize%blockSize != 0 {
		blocksNum++
	}

	blockList = make([]*BlockState, 0, blocksNum)
	var (
		id         int
		begin, end int64
	)

	for i := 0; i < blocksNum-1; i++ {
		end += blockSize
		blockList = append(blockList, &BlockState{
			ID: id,
			Range: ReadRange{
				Begin: begin,
				End:   end,
			},
		})
		id++
		begin = end
	}

	blockList = append(blockList, &BlockState{
		ID: id,
		Range: ReadRange{
			Begin: begin,
			End:   fileSize,
		},
	})
	return
}

// NewSplitUnit io.ReaderAt实现SplitUnit接口
func NewSplitUnit(readerAt io.ReaderAt, readRange ReadRange) SplitUnit {
	return &fileBlock{
		readerAt:  readerAt,
		readRange: readRange,
	}
}

// NewBufioSplitUnit io.ReaderAt实现SplitUnit接口
func NewBufioSplitUnit(readerAt io.ReaderAt, readRange ReadRange) SplitUnit {
	su := &fileBlock{
		readerAt:  readerAt,
		readRange: readRange,
	}
	return &bufioFileBlock{
		fileBlock: su,
		bufio:     bufio.NewReaderSize(su, BufioReadSize),
	}
}

func (bfb *bufioFileBlock) Read(b []byte) (n int, err error) {
	return bfb.bufio.Read(b)
}

func (fb *fileBlock) Read(b []byte) (n int, err error) {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if fb.readed+fb.readRange.Begin >= fb.readRange.End {
		return 0, io.EOF
	}

	left := int(fb.Left())
	if len(b) > left {
		n, err = fb.readerAt.ReadAt(b[:left], fb.readed+fb.readRange.Begin)
	} else {
		n, err = fb.readerAt.ReadAt(b, fb.readed+fb.readRange.Begin)
	}

	fb.readed += int64(n)
	return
}

func (fb *fileBlock) Seek(offset int64, whence int) (int64, error) {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	switch whence {
	case os.SEEK_SET:
		fb.readed = offset
	case os.SEEK_CUR:
		fb.readed += offset
	case os.SEEK_END:
		fb.readed = fb.readRange.End - fb.readRange.Begin + offset
	default:
		return 0, fmt.Errorf("unsupport whence: %d", whence)
	}
	if fb.readed < 0 {
		fb.readed = 0
	}
	return fb.readed, nil
}

func (fb *fileBlock) Len() int64 {
	return fb.readRange.End - fb.readRange.Begin
}

func (fb *fileBlock) Left() int64 {
	return fb.readRange.End - fb.readRange.Begin - fb.readed
}

func (fb *fileBlock) Range() ReadRange {
	return fb.readRange
}

func (fb *fileBlock) Readed() int64 {
	return fb.readed
}
