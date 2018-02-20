package pcspath

import (
	"path"
	"strings"
)

// EscapeBracketOne 转义中括号, 加一个反斜杠
func EscapeBracketOne(s string) string {
	if !strings.Contains(s, "[") && !strings.Contains(s, "]") {
		return s
	}

	builder := &strings.Builder{}
	for k := range s {
		if s[k] != '[' && s[k] != ']' {
			builder.WriteByte(s[k])
			continue
		}

		if k >= 1 && s[k-1] == '\\' {
			builder.WriteByte(s[k])
			continue
		}
		builder.WriteString(`\`)
		builder.WriteByte(s[k])
	}
	return builder.String()
}

// EscapeBracketTwo 转义中括号, 加两个反斜杠
func EscapeBracketTwo(s string) string {
	if !strings.Contains(s, "[") && !strings.Contains(s, "]") {
		return s
	}

	builder := &strings.Builder{}
	for k := range s {
		if s[k] != '[' && s[k] != ']' {
			builder.WriteByte(s[k])
			continue
		}

		if k >= 2 && s[k-1] == '\\' && s[k-2] == '\\' {
			builder.WriteByte(s[k])
			continue
		}
		builder.WriteString(`\\`)
		builder.WriteByte(s[k])
	}
	return builder.String()
}

// SplitAll 分割路径, "/"为分隔符
func SplitAll(pcspath string) (elem []string) {
	pcspath = path.Clean(pcspath)

	raw := strings.Split(pcspath, "/")

	if !path.IsAbs(pcspath) {
		elem = append(elem, raw[0])
	}

	for k := range raw[1:] {
		elem = append(elem, "/"+raw[k+1])
	}

	return
}
