package bdcrypto

import (
	"crypto/hmac"
	"crypto/sha1"
)

// HMacSha1 HMAC-SHA1签名认证
func HMacSha1(key, origData []byte) (cipherText []byte) {
	mac := hmac.New(sha1.New, key)
	mac.Write(origData)
	return mac.Sum(nil)
}
