// Package converter 格式, 类型转换包
package converter

import (
	"fmt"
	"strconv"
	"unsafe"
)

const (
	// B byte
	B = (int64)(1 << (10 * iota))
	// KB kilobyte
	KB
	// MB megabyte
	MB
	// GB gigabyte
	GB
	// TB terabyte
	TB
	// PB petabyte
	PB
)

// ConvertFileSize 文件大小格式化输出
func ConvertFileSize(size int64, precision ...int) string {
	pint := "6"
	if len(precision) == 1 {
		pint = fmt.Sprint(precision[0])
	}
	if size < 0 {
		return "0B"
	}
	if size < KB {
		return fmt.Sprintf("%dB", size)
	}
	if size < MB {
		return fmt.Sprintf("%."+pint+"fKB", float64(size)/float64(KB))
	}
	if size < GB {
		return fmt.Sprintf("%."+pint+"fMB", float64(size)/float64(MB))
	}
	if size < TB {
		return fmt.Sprintf("%."+pint+"fGB", float64(size)/float64(GB))
	}
	if size < PB {
		return fmt.Sprintf("%."+pint+"fTB", float64(size)/float64(TB))
	}
	return fmt.Sprintf("%."+pint+"fPB", float64(size)/float64(PB))
}

// ToString unsafe 转换, 将 []byte 转换为 string
func ToString(p []byte) string {
	return *(*string)(unsafe.Pointer(&p))
}

// ToBytes unsafe 转换, 将 string 转换为 []byte
func ToBytes(str string) []byte {
	return *(*[]byte)(unsafe.Pointer(&str))
}

// IntToBool int 类型转换为 bool
func IntToBool(i int) bool {
	return i != 0
}

// SliceStringToInt64 []string 转换为 []int64
func SliceStringToInt64(ss []string) (si []int64) {
	si = make([]int64, 0, len(ss))
	var (
		i   int64
		err error
	)
	for k := range ss {
		i, err = strconv.ParseInt(ss[k], 10, 64)
		if err != nil {
			continue
		}
		si = append(si, i)
	}
	return
}

// MustInt 将string转换为int, 忽略错误
func MustInt(s string) (n int) {
	n, _ = strconv.Atoi(s)
	return
}

// MustInt64 将string转换为int64, 忽略错误
func MustInt64(s string) (i int64) {
	i, _ = strconv.ParseInt(s, 10, 64)
	return
}

// ShortDisplay 缩略显示字符串s, 显示长度为num, 缩略的内容用"..."填充
func ShortDisplay(s string, num int) string {
	for k := range s {
		if k >= num {
			return string(s[:k]) + "..."
		}
	}
	return s
}
