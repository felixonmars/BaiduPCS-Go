package pcsutil

import (
	"fmt"
	"time"
)

const (
	b = (int64)(1 << (10 * iota))
	kb
	mb
	gb
	tb
	pb
)

// ConvertFileSize 文件大小格式化输出
func ConvertFileSize(size int64, precision ...int) string {
	pint := "6"
	if len(precision) == 1 {
		pint = fmt.Sprint(precision[0])
	}
	if size <= 0 {
		return "0"
	}
	if size < kb {
		return fmt.Sprintf("%."+pint+"fB", float64(size)/float64(b))
	}
	if size < mb {
		return fmt.Sprintf("%."+pint+"fKB", float64(size)/float64(kb))
	}
	if size < gb {
		return fmt.Sprintf("%."+pint+"fMB", float64(size)/float64(mb))
	}
	if size < tb {
		return fmt.Sprintf("%."+pint+"fGB", float64(size)/float64(gb))
	}
	if size < pb {
		return fmt.Sprintf("%."+pint+"fTB", float64(size)/float64(tb))
	}
	return fmt.Sprintf("%."+pint+"fPB", float64(size)/float64(pb))
}

// IntToBool int 类型转换为 bool
func IntToBool(i int) bool {
	return i != 0
}

// FormatTime 将 Unix 时间戳, 转换为字符串
func FormatTime(t int64) string {
	return time.Unix(t, 0).Format("2006-01-02 03:04:05")
}
