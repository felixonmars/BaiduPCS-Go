package pcscommand

import (
	"fmt"
	libpath "path"
	"strings"
)

// RunTree 列出树形图
func RunTree(path string, depth int, pprefix []string) {
	path, err := getAbsPath(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	files, err := info.FilesDirectoriesList(path, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i, file := range files {
		var prefix string
		if i+1 == len(files) {
			prefix = "└──"
			pprefix = pprefix[0:len(pprefix)]
		} else {
			prefix = "├──"
		}

		if file.Isdir {
			fmt.Printf("%v%v %v/\n", strings.Join(pprefix, ""), prefix, file.Filename)
			RunTree(libpath.Join(path, file.Filename), depth+1, append(pprefix, "│   "))
			continue
		}

		fmt.Printf("%v%v %v\n", strings.Join(pprefix, ""), prefix, file.Filename)
	}

	return
}
