// Package pcscaptcha 验证码处理包
// TODO: 直接打开验证码
package pcscaptcha

import (
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"os"
	"path/filepath"
)

const (
	// CaptchaName 验证码文件名称
	CaptchaName = "captcha.png"
)

// CaptchaPath 返回验证码存放路径
func CaptchaPath() string {
	return filepath.Join(os.TempDir(), CaptchaName)
}

// RemoveOldCaptchaPath 移除旧的验证码路径
func RemoveOldCaptchaPath() error {
	return os.Remove(filepath.Join(pcsconfig.GetConfigDir(), CaptchaName))
}

// RemoveCaptchaPath 移除验证码路径
func RemoveCaptchaPath() error {
	return os.Remove(CaptchaPath())
}
