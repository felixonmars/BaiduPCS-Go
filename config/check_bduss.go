package pcsconfig

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/BaiduPCS-Go/util"
	"strconv"
)

// Baidu 百度帐号对象
type Baidu struct {
	UID    uint64 `json:"uid"`
	Name   string `json:"name"`
	BDUSS  string `json:"bduss"`
	PTOKEN string `json:"ptoken"`
	STOKEN string `json:"stoken"`

	Workdir string `json:"workdir"`
}

// NewWithBDUSS 检测BDUSS有效性, 同时获取百度详细信息
func NewWithBDUSS(bduss string) (*Baidu, error) {
	h := requester.NewHTTPClient()
	timestamp := pcsutil.BeijingTimeOption("")
	post := map[string]string{
		"bdusstoken":  bduss + "|null",
		"channel_id":  "",
		"channel_uid": "",
		"stErrorNums": "0",
		"subapp_type": "mini",
		"timestamp":   timestamp + "922",
	}
	pcsutil.TiebaClientSignature(post)

	header := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"Cookie":       "ka=open",
		"net":          "1",
		"User-Agent":   "bdtb for Android 6.9.2.1",
		"client_logid": timestamp + "416",
		"Connection":   "Keep-Alive",
	}

	body, err := h.Fetch("POST", "http://tieba.baidu.com/c/s/login", post, header) // 获取百度ID的TBS，UID，BDUSS等
	if err != nil {
		return nil, fmt.Errorf("检测BDUSS有效性失败, %s", err)
	}

	json, err := simplejson.NewJson(body)
	if err != nil {
		return nil, fmt.Errorf("检测BDUSS有效性json解析出错: %s", err)
	}

	errCode := json.Get("error_code").MustString()
	errMsg := json.Get("error_msg").MustString()

	switch errCode {
	case "0":
	case "1":
		return nil, fmt.Errorf("检测BDUSS有效性错误, 百度BDUSS格式不正确或者已过期")
	default:
		return nil, fmt.Errorf("检测BDUSS有效性错误代码: %s, 消息: %s", errCode, errMsg)
	}

	uidStr := json.GetPath("user", "id").MustString()
	uid, _ := strconv.ParseUint(uidStr, 10, 64)

	username, err := GetUserNameByUID(uid)
	if err != nil {
		return nil, err
	}

	return &Baidu{
		UID:     uid,
		Name:    username,
		BDUSS:   bduss,
		Workdir: "/",
	}, nil
}

// GetUserNameByUID 通过百度uid获取百度用户名
func GetUserNameByUID(uid uint64) (username string, err error) {
	rawQuery := "has_plist=0&need_post_count=1&rn=1&uid=" + fmt.Sprint(uid)
	urlStr := "http://c.tieba.baidu.com/c/u/user/profile?" + pcsutil.TiebaClientRawQuerySignature(rawQuery)

	body, err := requester.HTTPGet(urlStr)
	if err != nil {
		return "", err
	}

	json, err := simplejson.NewJson(body)
	if err != nil {
		return "", err
	}

	userJSON := json.GetPath("user")
	username = userJSON.Get("name").MustString()
	return
}
