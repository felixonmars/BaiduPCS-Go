package pcsutil

import (
	"compress/gzip"
	"crypto/md5"
	"flag"
	"fmt"
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

// Md5Encrypt 对 str 进行md5加密, 返回值为 str 加密后的密文
func Md5Encrypt(str interface{}) string {
	md5Ctx := md5.New()
	switch value := str.(type) {
	case string:
		md5Ctx.Write([]byte(str.(string)))
	case *string:
		md5Ctx.Write([]byte(*str.(*string)))
	case []byte:
		md5Ctx.Write(str.([]byte))
	case *[]byte:
		md5Ctx.Write(*str.(*[]byte))
	default:
		fmt.Println("MD5Encrypt: undefined type:", value)
		return ""
	}
	return fmt.Sprintf("%X", md5Ctx.Sum(nil))
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
