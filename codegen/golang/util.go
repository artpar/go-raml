package golang

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// try to set root import path based on target dir
// - do nothing if user specify import path
// - if target dir is under GOPATH, set it to the target dir
func setRootImportPath(importPath, targetDir string) string {
	// use import path if not empty
	if len(importPath) > 0 {
		return importPath
	}

	// get GOPATH dir
	gopath, err := filepath.Abs(os.Getenv("GOPATH"))
	if err != nil {
		return ""
	}
	gopath = filepath.Join(gopath, "src")

	// get absolute target dir
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		panic("invalid targetDir:" + err.Error())
	}

	// panic if user doesn't specify import path and target dir
	// not under GOPATH.
	if !strings.HasPrefix(absTargetDir, gopath) {
		panic("please specify '--import-path' or set '--dir' under your GOPATH")
	}

	// set import path
	newImportPath, err := filepath.Rel(gopath, absTargetDir)
	if err != nil {
		panic("failed to set import path automatically:" + err.Error())
	}

	return newImportPath
}

// escape identifier to meet this spec
// https://golang.org/ref/spec#Identifiers
// - firstChar need to be letter, prepend "The_" if it is not
// - replace "-" with "_"
func escapeIdentifier(str string) string {
	r := []rune(str)
	if !unicode.IsLetter(r[0]) {
		r = append([]rune("The_"), r...)
	}
	str = string(r)
	return strings.Replace(str, "-", "_", -1)
}
