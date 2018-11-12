package args

import (
	"strings"
	"unicode"
)

const (
	CharEscape      = '\\'
	CharSingleQuote = '\''
	CharDoubleQuote = '"'
	CharBackQuote   = '`'
)

// IsQuote 是否为引号
func IsQuote(r rune) bool {
	return r == CharSingleQuote || r == CharDoubleQuote || r == CharBackQuote
}

// Parse 解析line, 忽略括号
func Parse(line string) (lineArgs []string) {
	var (
		rl        = []rune(line + " ")
		buf       = strings.Builder{}
		quoteChar rune
		escaped   bool
		in        bool
	)

	var (
		isSpace bool
	)

	for _, r := range rl {
		isSpace = unicode.IsSpace(r)
		if !isSpace && !in {
			in = true
		}

		switch {
		case escaped: // 已转义, 跳过
			escaped = false
			//pass
		case r == CharEscape: // 转义模式
			escaped = true
			continue
		case IsQuote(r):
			if quoteChar == 0 { //未引
				quoteChar = r
				continue
			}

			if quoteChar == r { //取消引
				quoteChar = 0
				continue
			}
		case isSpace:
			if !in { // 忽略多余的空格
				continue
			}
			if quoteChar == 0 { // 未在引号内
				lineArgs = append(lineArgs, buf.String())
				buf.Reset()
				in = false
				continue
			}
		}

		buf.WriteRune(r)
	}

	return
}
