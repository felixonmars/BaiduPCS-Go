// Package pcsutil 工具包
package pcsutil

import (
	"compress/gzip"
	"flag"
	"io"
	"io/ioutil"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
)

var (
	// PipeInput 命令中是否为管道输入
	PipeInput bool
)

func init() {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return
	}
	PipeInput = (fileInfo.Mode() & os.ModeNamedPipe) == os.ModeNamedPipe
}

// ContainsString 检测字符串是否在字符串数组里
func ContainsString(ss []string, s string) bool {
	for k := range ss {
		if strings.Compare(ss[k], s) == 0 {
			return true
		}
	}
	return false
}

// GetURLCookieString 返回cookie字串
func GetURLCookieString(urlString string, jar *cookiejar.Jar) string {
	url, _ := url.Parse(urlString)
	cookies := jar.Cookies(url)
	cookieString := ""
	for _, v := range cookies {
		cookieString += v.String() + "; "
	}
	cookieString = strings.TrimRight(cookieString, "; ")
	return cookieString
}

// DecompressGZIP 对 io.Reader 数据, 进行 gzip 解压
func DecompressGZIP(r io.Reader) ([]byte, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	gzipReader.Close()
	return ioutil.ReadAll(gzipReader)
}

// FlagProvided 检测命令行是否提供名为 name 的 flag, 支持多个name(names)
func FlagProvided(names ...string) bool {
	if len(names) == 0 {
		return false
	}
	var targetFlag *flag.Flag
	for _, name := range names {
		targetFlag = flag.Lookup(name)
		if targetFlag == nil {
			return false
		}
		if targetFlag.DefValue == targetFlag.Value.String() {
			return false
		}
	}
	return true
}
