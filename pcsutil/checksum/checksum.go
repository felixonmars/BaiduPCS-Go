// Package checksum 校验本地文件包
package checksum

import (
	"crypto/md5"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"hash"
	"hash/crc32"
	"io"
	"os"
)

const (
	defaultBufSize = 256 * converter.KB
)

type (
	// LocalFileMeta 本地文件元信息
	LocalFileMeta struct {
		Path     string `json:"path"`     // 本地路径
		Length   int64  `json:"length"`   // 文件大小
		SliceMD5 []byte `json:"slicemd5"` // 文件前 requiredSliceLen (256KB) 切片的 md5 值
		MD5      []byte `json:"md5"`      // 文件的 md5
		CRC32    uint32 `json:"crc32"`    // 文件的 crc32
		ModTime  int64  `json:"modtime"`  // 修改日期
	}

	// LocalFile LocalFile
	LocalFile struct {
		LocalFileMeta

		bufSize int
		buf     []byte
		File    *os.File // 文件
	}

	// SumConfig 计算文件摘要值配置
	SumConfig struct {
		IsMD5Sum      bool
		IsSliceMD5Sum bool
		IsCRC32Sum    bool
	}
)

func NewLocalFileInfo(localPath string, bufSize int) *LocalFile {
	return &LocalFile{
		LocalFileMeta: LocalFileMeta{
			Path: localPath,
		},
		bufSize: bufSize,
	}
}

// OpenPath 检查文件状态并获取文件的大小 (Length)
func (lf *LocalFile) OpenPath() error {
	if lf.File != nil {
		lf.File.Close()
	}

	var err error
	lf.File, err = os.Open(lf.Path)
	if err != nil {
		return err
	}

	info, err := lf.File.Stat()
	if err != nil {
		return err
	}

	lf.Length = info.Size()
	lf.ModTime = info.ModTime().Unix()
	return nil
}

// Close 关闭文件
func (lf *LocalFile) Close() error {
	if lf.File == nil {
		return fmt.Errorf("file is nil")
	}

	return lf.File.Close()
}

func (lf *LocalFile) initBuf() {
	if lf.buf == nil {
		if lf.bufSize != 0 {
			lf.buf = make([]byte, lf.bufSize)
			return
		}

		lf.buf = make([]byte, defaultBufSize)
	}
}

func (lf *LocalFile) repeatRead(ws ...io.Writer) {
	if lf.File == nil {
		return
	}

	lf.initBuf()

	var (
		begin int64
		n     int
		err   error
	)

	handle := func() {
		begin += int64(n)
		for k := range ws {
			ws[k].Write(lf.buf[:n])
		}
	}

	// 读文件
	for {
		n, err = lf.File.ReadAt(lf.buf, begin)
		if err != nil {
			if err == io.EOF {
				handle()
			} else {
				fmt.Printf("%s\n", err)
			}
			break
		}

		handle()
	}
}

// Sum 计算文件摘要值
func (lf *LocalFile) Sum(cfg SumConfig) {
	var (
		md5w   hash.Hash
		crc32w hash.Hash32
	)

	ws := make([]io.Writer, 0, 2)
	if cfg.IsMD5Sum {
		md5w = md5.New()
		ws = append(ws, md5w)
	}
	if cfg.IsCRC32Sum {
		crc32w = crc32.NewIEEE()
		ws = append(ws, crc32w)
	}
	if cfg.IsSliceMD5Sum {
		lf.SliceMD5Sum()
	}

	lf.repeatRead(ws...)

	if cfg.IsMD5Sum {
		lf.MD5 = md5w.Sum(nil)
	}
	if cfg.IsCRC32Sum {
		lf.CRC32 = crc32w.Sum32()
	}
}

// Md5Sum 获取文件的 md5 值
func (lf *LocalFile) Md5Sum() {
	lf.Sum(SumConfig{
		IsMD5Sum: true,
	})
}

// SliceMD5Sum 获取文件前 requiredSliceLen (256KB) 切片的 md5 值
func (lf *LocalFile) SliceMD5Sum() {
	if lf.File == nil {
		return
	}

	// 获取前 256KB 文件切片的 md5
	lf.initBuf()

	m := md5.New()
	n, err := lf.File.ReadAt(lf.buf, 0)
	if err != nil {
		if err == io.EOF {
			goto md5sum
		} else {
			fmt.Printf("SliceMD5Sum: %s\n", err)
			return
		}
	}

md5sum:
	m.Write(lf.buf[:n])
	lf.SliceMD5 = m.Sum(nil)
}

// Crc32Sum 获取文件的 crc32 值
func (lf *LocalFile) Crc32Sum() {
	lf.Sum(SumConfig{
		IsCRC32Sum: true,
	})
}
