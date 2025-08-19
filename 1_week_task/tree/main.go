package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, pF bool) error {
	return dirTreeRec(out, path, "", pF)
}

func dirTreeRec(out io.Writer, path string, level string, pF bool) error {
	files, err := os.ReadDir(path)
	lenFiles := 0
	if !pF {
		for _, file := range files {
			if file.Type() == fs.ModeDir {
				lenFiles++
			}
		}
	} else {
		lenFiles = len(files)
	}
	if err != nil {
		return err
	}
	sc := 1
	for _, file := range files {
		var branchStr string
		var indent string
		if sc == lenFiles {
			branchStr = "└───"
			indent = "	"
		} else {
			branchStr = "├───"
			indent = "│	"
		}

		t := file.Type()
		if t != fs.ModeDir && pF {
			fI, err := file.Info()
			if err != nil {
				return err
			}
			var sizeStr string
			if fI.Size() == 0 {
				sizeStr = " (empty)"
			} else {
				sizeStr = " (" + strconv.FormatInt(fI.Size(), 10) + "b)"
			}
			result := level + branchStr + fI.Name() + sizeStr
			fmt.Fprintln(out, result)
			sc++
		} else if t == fs.ModeDir {
			result := level + branchStr + file.Name()
			fmt.Fprintln(out, result)
			err = dirTreeRec(out, path+"/"+file.Name(), level+indent, pF)
			if err != nil {
				return err
			}
			sc++
		}

	}
	return nil
}
