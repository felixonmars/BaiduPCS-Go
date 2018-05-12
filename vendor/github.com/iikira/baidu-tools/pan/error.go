package pan

import (
	"errors"
	"fmt"
)

var (
	// ErrSharedInfoNotInit 未初始化分享信息
	ErrSharedInfoNotInit = errors.New("未初始化分享信息")
	// ErrSharedInfoNotSetSharedURL 未设置分享链接
	ErrSharedInfoNotSetSharedURL = errors.New("未设置分享链接")
)

// RemoteErrInfo 错误详情
type RemoteErrInfo struct {
	ErrNo  int    `json:"errno"`
	ErrMsg string `json:"err_msg"`
}

func (rei *RemoteErrInfo) Error() string {
	rei.ParseErrMsg()

	if rei.ErrNo == 0 {
		return fmt.Sprint("操作成功")
	}

	if rei.ErrMsg == "" {
		return fmt.Sprintf("错误代码: %d", rei.ErrNo)
	}

	return fmt.Sprintf("错误代码: %d, 消息: %s", rei.ErrNo, rei.ErrMsg)
}

// ParseErrMsg 根据 ErrNo, 解析网盘错误信息
func (rei *RemoteErrInfo) ParseErrMsg() {
	if rei.ErrMsg != "" || rei.ErrNo == 0 {
		return
	}

	switch rei.ErrNo {
	case -1:
		rei.ErrMsg = "由于您分享了违反相关法律法规的文件，分享功能已被禁用，之前分享出去的文件不受影响。"
	case -2:
		rei.ErrMsg = "用户不存在,请刷新页面后重试"
	case -3:
		rei.ErrMsg = "文件不存在,请刷新页面后重试"
	case -4:
		rei.ErrMsg = "登录信息有误，请重新登录试试"
	case -5:
		rei.ErrMsg = "host_key和user_key无效"
	case -6:
		rei.ErrMsg = "请重新登录"
	case -7:
		rei.ErrMsg = "该分享已删除或已取消"
	case -8:
		rei.ErrMsg = "该分享已经过期"
	case -9:
		rei.ErrMsg = "文件不存在"
	case -10:
		rei.ErrMsg = "分享外链已经达到最大上限100000条，不能再次分享"
	case -11:
		rei.ErrMsg = "验证cookie无效"
	case -12:
		rei.ErrMsg = "访问密码错误"
	case -14:
		rei.ErrMsg = "对不起，短信分享每天限制20条，你今天已经分享完，请明天再来分享吧！"
	case -15:
		rei.ErrMsg = "对不起，邮件分享每天限制20封，你今天已经分享完，请明天再来分享吧！"
	case -16:
		rei.ErrMsg = "对不起，该文件已经限制分享！"
	case -17:
		rei.ErrMsg = "文件分享超过限制"
	case -19:
		rei.ErrMsg = "需要输入验证码"
	case -30:
		rei.ErrMsg = "文件已存在"
	case -31:
		rei.ErrMsg = "文件保存失败"
	case -33:
		rei.ErrMsg = "一次支持操作999个，减点试试吧"
	case -62:
		rei.ErrMsg = "可能需要输入验证码"
	case -70:
		rei.ErrMsg = "你分享的文件中包含病毒或疑似病毒，为了你和他人的数据安全，换个文件分享吧"
	case 2:
		rei.ErrMsg = "参数错误"
	case 3:
		rei.ErrMsg = "未登录或帐号无效"
	case 4:
		rei.ErrMsg = "存储好像出问题了，请稍候再试"
	case 108:
		rei.ErrMsg = "文件名有敏感词，优化一下吧"
	case 110:
		rei.ErrMsg = "分享次数超出限制，可以到“我的分享”中查看已分享的文件链接"
	case 112:
		rei.ErrMsg = "页面已过期，请刷新后重试"
	case 113:
		rei.ErrMsg = "签名错误"
	case 114:
		rei.ErrMsg = "当前任务不存在，保存失败"
	case 115:
		rei.ErrMsg = "该文件禁止分享"
	default:
		rei.ErrMsg = "未知错误"
	}
}
