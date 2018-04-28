package pcspath

import (
	"fmt"
	"testing"
)

func TestEscapeBlank(t *testing.T) {
	fmt.Println(Escape("asdfj(alsf)djlsf"))
	fmt.Println(Escape("asdfjal\\\\\\ sfdj lsf"))
	fmt.Println(Escape("asdfjal\\\\ sfdj lsf"))
	fmt.Println(Escape("asdfjal\\ s (asdfa) [asf] fdj lsf"))
}
