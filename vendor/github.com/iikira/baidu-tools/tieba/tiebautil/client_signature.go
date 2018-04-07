package tiebautil

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"github.com/iikira/Baidu-Login/bdcrypto"
	"github.com/iikira/baidu-tools/ramdominfo"
	"sort"
	"strconv"
)

// TiebaClientSignature 根据给定贴吧客户端的post数据进行签名, 以通过百度服务器验证. 返回值为签名后的 post
func TiebaClientSignature(post map[string]string) {
	if post == nil {
		post = map[string]string{}
	}

	// 已经签名, 则重新签名
	if _, ok := post["sign"]; ok {
		delete(post, "sign")
	}

	var (
		bduss        = post["BDUSS"]
		phoneIMEIStr = strconv.FormatUint(ramdominfo.SumIMEI(bduss), 10)
	)

	// 预设
	post["_client_type"] = "2"
	post["_client_version"] = "7.0.0.0"
	post["_phone_imei"] = phoneIMEIStr
	post["from"] = "mini_ad_wandoujia"
	post["model"] = ramdominfo.GetPhoneModel(bduss)
	post["cuid"] = ramdominfo.RamdomMD5UpperString() + "|" + bdcrypto.StringReverse(phoneIMEIStr)

	keys := make([]string, 0, len(post))
	for key := range post {
		keys = append(keys, key)
	}
	sort.Sort(sort.StringSlice(keys))

	m := md5.New()
	for _, key := range keys {
		m.Write([]byte(key + "=" + post[key]))
	}
	m.Write([]byte("tiebaclient!!!"))

	post["sign"] = hex.EncodeToString(m.Sum(nil))
}

// TiebaClientRawQuerySignature 给 rawQuery 进行贴吧客户端签名, 返回值为签名后的 rawQuery
func TiebaClientRawQuerySignature(rawQuery string) (signedRawQuery string) {
	m := md5.New()
	m.Write(bytes.Replace([]byte(rawQuery), []byte("&"), nil, -1))
	m.Write([]byte("tiebaclient!!!"))

	signedRawQuery = rawQuery + "&sign=" + hex.EncodeToString(m.Sum(nil))
	return
}
