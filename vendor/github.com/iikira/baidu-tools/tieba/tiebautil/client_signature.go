package tiebautil

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"sort"
)

// TiebaClientSignature 根据给定贴吧客户端的post数据进行签名, 以通过百度服务器验证. 返回值为签名后的 post
func TiebaClientSignature(post map[string]string) {
	if post == nil {
		return
	}

	// 预设
	post["_client_type"] = "2"
	post["_client_version"] = "6.9.2.1"
	post["_phone_imei"] = "860983036542682"
	post["from"] = "mini_ad_wandoujia"
	post["model"] = "HUAWEI NXT-AL10"
	post["cuid"] = "61464018582906C485355A89D105ECFB|286245630389068"

	var keys []string
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
