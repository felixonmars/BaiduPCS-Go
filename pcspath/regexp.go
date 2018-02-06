package pcspath

import (
	"regexp"
)

var (
	patternRE = regexp.MustCompile(`[\*\?]`)
)
