package pcsutil

import (
	"regexp"
)

var (
	// HTTPSRE https regexp
	HTTPSRE = regexp.MustCompile("^https")
	// ChinaPhoneRE https regexp
	ChinaPhoneRE = regexp.MustCompile("^((13[0-9])|(14[5|7])|(15([0-3]|[5-9]))|(18[0,5-9]))\\d{8}$")
)
