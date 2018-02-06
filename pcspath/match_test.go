package pcspath

import (
	"fmt"
	"testing"
)

func TestMatch(t *testing.T) {
	workdir := "/"
	pp := NewPCSPath(&workdir, "[123?")
	fmt.Println(pp.AbsPathNoMatch())
	fmt.Println(pp.Match([]string{"123", "/[1234", "12345"}...))
}

func TestTest(t *testing.T) {
	fmt.Println(SplitAll("/1/2/3"))
}
