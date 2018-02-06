package pcspath

import (
	"path"
)

// AbsPathNoMatch 返回绝对路径, 不检测通配符
func (pp *PCSPath) AbsPathNoMatch() string {
	pp.CleanPath()
	if !path.IsAbs(pp.SubPath) {
		return path.Clean(*pp.WorkdirDestination + "/" + pp.SubPath)
	}
	return pp.SubPath
}

// Match 检测 pcspaths 的通配符, 返回匹配成功的 matchedPCSPaths
func (pp *PCSPath) Match(pcspaths ...string) (matchedPCSPaths []string) {
	pattern := pp.AbsPathNoMatch() // 获取绝对路径

	for k := range pcspaths {
		matched, _ := path.Match(pattern, pcspaths[k])
		if !matched {
			continue
		}

		matchedPCSPaths = append(matchedPCSPaths, pcspaths[k])
	}

	return matchedPCSPaths
}
