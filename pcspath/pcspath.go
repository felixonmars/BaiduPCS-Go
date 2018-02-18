// Package pcspath 网盘路径工具包
package pcspath

import (
	"path"
)

// PCSPath 百度 PCS 路径
type PCSPath struct {
	WorkdirDestination *string // 指向工作目录
	SubPath            string  // 相对于工作目录的子目录
}

// NewPCSPath 返回 PCSPath 指针对象
func NewPCSPath(workdirDestination *string, pcsSubPath string) *PCSPath {
	pp := &PCSPath{
		WorkdirDestination: workdirDestination,
		SubPath:            pcsSubPath,
	}
	pp.CleanPath()
	return pp
}

// CleanPath 过滤处理目录
func (pp *PCSPath) CleanPath() {
	pp.SubPath = path.Clean(pp.SubPath)
	*pp.WorkdirDestination = path.Clean(*pp.WorkdirDestination)
}

// EscapeBracket 转义文件名中的中括号
func (pp *PCSPath) EscapeBracket() {
	pp.SubPath = EscapeBracketOne(pp.SubPath)
	*pp.WorkdirDestination = EscapeBracketOne(*pp.WorkdirDestination)
}

// SetSubPath 设置子目录
func (pp *PCSPath) SetSubPath(pcsSubPath string) {
	pp.SubPath = path.Clean(pcsSubPath)
}
