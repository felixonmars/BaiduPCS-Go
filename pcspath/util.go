package pcspath

import (
	"bytes"
	"path"
	"strings"
)

// EscapeBracketOne 转义中括号, 加一个反斜杠
func EscapeBracketOne(s string) string {
	if !strings.Contains(s, "[") && !strings.Contains(s, "]") {
		return s
	}

	buf := bytes.NewBuffer(nil)
	for k := range s {
		if s[k] != '[' && s[k] != ']' {
			buf.WriteByte(s[k])
			continue
		}

		if k >= 1 && s[k-1] == '\\' {
			buf.WriteByte(s[k])
			continue
		}
		buf.WriteString(`\`)
		buf.WriteByte(s[k])
	}
	return buf.String()
}

// EscapeBracketTwo 转义中括号, 加两个反斜杠
func EscapeBracketTwo(s string) string {
	if !strings.Contains(s, "[") && !strings.Contains(s, "]") {
		return s
	}

	buf := bytes.NewBuffer(nil)
	for k := range s {
		if s[k] != '[' && s[k] != ']' {
			buf.WriteByte(s[k])
			continue
		}

		if k >= 2 && s[k-1] == '\\' && s[k-2] == '\\' {
			buf.WriteByte(s[k])
			continue
		}
		buf.WriteString(`\\`)
		buf.WriteByte(s[k])
	}
	return buf.String()
}

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
