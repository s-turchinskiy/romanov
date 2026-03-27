package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	shiftItem     = "├───"
	shiftLastItem = "└───"
	shiftTopLevel = "│"
)

func main() {
	out := os.Stdout
	if len(os.Args) != 2 && len(os.Args) != 3 {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, dir string, printFiles bool) error {

	return printDir(out, dir, printFiles, "")
}

func printDir(out io.Writer, dir string, printFiles bool, shift string) error {
	list, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error read dir %s, error : %w", dir, err)
	}

	selectionList := selectionFiles(printFiles, list)

	err = printItems(out, dir, printFiles, shift, selectionList)
	if err != nil {
		return err
	}

	return nil
}

func selectionFiles(printFiles bool, list []os.DirEntry) []os.DirEntry {
	if printFiles {
		return list
	}

	newList := make([]os.DirEntry, 0)
	for _, item := range list {
		if item.IsDir() {
			newList = append(newList, item)
		}

	}
	return newList
}

func printItems(out io.Writer, dir string, printFiles bool, shift string, list []os.DirEntry) error {
	lastIndex := len(list) - 1
	for i, item := range list {

		fileFullName := filepath.Join(dir, item.Name())

		var builder strings.Builder

		builder.WriteString(shift)
		if i == lastIndex {
			builder.WriteString(shiftLastItem)
		} else {
			builder.WriteString(shiftItem)
		}

		builder.WriteString(item.Name())

		err := addSizeInfoFirFile(item, &builder, fileFullName)
		if err != nil {
			return err
		}

		builder.WriteString("\n")
		_, err = out.Write([]byte(builder.String()))
		if err != nil {
			return err
		}

		if item.IsDir() {
			newShift := shift
			if i != lastIndex {
				newShift = newShift + shiftTopLevel
			}
			newShift = newShift + "\t"

			err = printDir(out, fileFullName, printFiles, newShift)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func addSizeInfoFirFile(item os.DirEntry, builder *strings.Builder, fileFullName string) error {

	if item.IsDir() {
		return nil
	}

	finfo, err := item.Info()
	if err != nil {
		return fmt.Errorf("error get file info %s, error : %w", fileFullName, err)
	}

	builder.WriteString(" (")
	if finfo.Size() == 0 {
		builder.WriteString("empty")
	} else {
		builder.WriteString(strconv.FormatInt(finfo.Size(), 10) + "b")
	}
	builder.WriteString(")")

	return nil
}
