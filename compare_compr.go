package main

import (
	"bytes"
	"compress/gzip"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
)

func compareCompr() {
	logf(ctx(), "compareCompr\n")
	var buf bytes.Buffer
	nFiles := 0
	filepath.WalkDir("cheatsheets", func(path string, de fs.DirEntry, err error) error {
		must(err)
		//logf(ctx(), "path: '%s'\n", path)
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		nFiles++
		d, err := os.ReadFile(path)
		must(err)
		buf.Write(d)
		return nil
	})
	d := buf.Bytes()
	logf(ctx(), "compareCompr: %d files of size %s\n", nFiles, formatSize(int64(len(d))))

	origSize := int64(len(d))
	logf(ctx(), "un: %d %s\n", origSize, formatSize(origSize))

	{
		var bufCompr bytes.Buffer
		timeStart := time.Now()
		w, err := gzip.NewWriterLevel(&bufCompr, gzip.BestCompression)
		must(err)
		_, err = w.Write(d)
		must(err)
		err = w.Close()
		must(err)
		dur := time.Since(timeStart)
		compSize := int64(bufCompr.Len())
		p := perc(origSize, compSize)
		logf(ctx(), "gz: %d %s in %s %.2f%%\n", compSize, formatSize(compSize), dur, p)
	}

	{
		var bufCompr bytes.Buffer
		timeStart := time.Now()
		w := brotli.NewWriterLevel(&bufCompr, gzip.BestCompression)
		_, err := w.Write(d)
		must(err)
		err = w.Close()
		must(err)
		dur := time.Since(timeStart)
		compSize := int64(bufCompr.Len())
		p := perc(origSize, compSize)
		logf(ctx(), "br: %d %s in %s %.2f%%\n", compSize, formatSize(compSize), dur, p)
	}
}
