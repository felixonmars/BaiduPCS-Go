package pcspath

import (
	"fmt"
	"testing"
)

func TestEscapeBlank(t *testing.T) {
	fmt.Println(EscapeBlank("asdfjalsfdjlsf"))
	fmt.Println(EscapeBlank("asdfjal\\\\\\ sfdj lsf"))
	fmt.Println(EscapeBlank("asdfjal\\\\ sfdj lsf"))
}
