package pan

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const (
	// BDKey 百度 HMAC-SHA1 密钥
	BDKey = "B8ec24caf34ef7227c66767d29ffd3fb"
)

var (
	// YunDataExp 解析网盘分享首页的数据的正则表达式
	YunDataExp = regexp.MustCompile(`window\.yunData[\s]?=[\s]?(.*?);`)
)

// MustParseInt64 将字符串转换为 int64, 忽略错误
func MustParseInt64(s string) (i int64) {
	i, _ = strconv.ParseInt(s, 10, 64)
	return
}

// MustParseInt 将字符串转换为 int, 忽略错误
func MustParseInt(s string) (i int) {
	i, _ = strconv.Atoi(s)
	return
}

// signature 签名
func (si *SharedInfo) signature() {
	si.Timestamp = time.Now().Unix()
	orig := fmt.Sprintf("%d_%d__%d", si.ShareID, si.UK, si.Timestamp)

	mac := hmac.New(sha1.New, []byte(BDKey))
	mac.Write([]byte(orig))
	si.Sign = mac.Sum(nil)
}
