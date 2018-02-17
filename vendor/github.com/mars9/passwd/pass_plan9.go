package passwd

import (
	"bufio"
	"fmt"
	"os"
)

func init() {
	passFunc = getPasswd
}

func getPasswd(prompt string) ([]byte, error) {
	cons, err := os.OpenFile("/dev/cons", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	defer cons.Close()
	consctl, err := os.OpenFile("/dev/consctl", os.O_WRONLY, 0)
	if err != nil {
		return nil, err
	}
	defer consctl.Close()

	fmt.Fprint(consctl, "rawon")
	defer fmt.Fprint(consctl, "rawoff")

	fmt.Fprint(cons, prompt)
	r := bufio.NewReader(cons)
	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	fmt.Fprint(cons, "\n")
	return line[:len(line)-1], nil
}
