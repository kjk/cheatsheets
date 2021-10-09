package main

import (
	"context"
	"io/ioutil"

	"github.com/kjk/common/u"
)

var (
	must                     = u.Must
	panicIf                  = u.PanicIf
	isWindows                = u.IsWindows
	openBrowser              = u.OpenBrowser
	dirExists                = u.DirExists
	normalizeNewlinesInPlace = u.NormalizeNewlinesInPlace
	formatSize               = u.FormatSize
	formatDuration           = u.FormatDuration
	perc                     = u.Percent
)

func ctx() context.Context {
	return context.Background()
}

func readFileMust(path string) []byte {
	d, err := ioutil.ReadFile(path)
	must(err)
	return d
}
