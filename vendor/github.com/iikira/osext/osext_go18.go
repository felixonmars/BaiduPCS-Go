//+build go1.8,!openbsd

package osext

import (
	"os"
	"path/filepath"
)

func executable() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(p)
}
