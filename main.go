package main

import (
	"bytes"
	"flag"
	"net/http"
	"strings"
	"time"
)

func buildContentCheatsheets() []URLContent {
	cheatsheets := readCheatSheets()
	csFindByURL := func(uri string) *cheatSheet {
		// match /cheatsheet/go.html => go
		uriBase := strings.ToLower(strings.TrimPrefix(uri, "/cheatsheet/"))
		if len(uri) == len(uriBase) {
			// doesn't start with /cheatsheet
			logf(ctx(), "csFindByURL: no match for '%s because doesn't start with /cheatsheet/'\n", uri)
			return nil
		}
		uriBaseNoExt := strings.TrimSuffix(uriBase, ".html")
		if len(uriBase) == len(uriBaseNoExt) {
			// doens't end with .html
			logf(ctx(), "csFindByURL: no match for '%s' because doesn't end with .html\n", uri)
			return nil
		}
		logf(ctx(), "csFindByURL: looking for '%s'\n", uriBaseNoExt)
		for _, cs := range cheatsheets {
			if uriBaseNoExt == cs.fileNameBase {
				logf(ctx(), "csFindByURL: found match for '%s'\n", uri)
				return cs
			}
		}
		logf(ctx(), "csFindByURL: no match for '%s'\n", uri)
		return nil
	}
	csMatches := func(uri string) bool {
		cs := csFindByURL(uri)
		return cs != nil
	}
	csSend := func(w http.ResponseWriter, r *http.Request) error {
		uri := r.URL.Path
		cs := csFindByURL(uri)
		panicIf(cs == nil, "no match for '%s'", uri)
		html := genCheatsheetHTML(cs)
		content := bytes.NewReader(html)
		http.ServeContent(w, r, "foo.html", time.Time{}, content)
		return nil
	}

	csIndexMatches := func(uri string) bool {
		matches := uri == "/" || uri == "/index.html"
		if matches {
			logf(ctx(), "csIndexMatches: match for '%s'\n", uri)
		}
		return matches
	}

	csIndexSend := func(w http.ResponseWriter, r *http.Request) error {
		logf(ctx(), "csIndexSend: '%s'\n", r.URL)
		uri := r.URL.Path
		panicIf(!csIndexMatches(uri), "no match for '%s'", uri)
		html := genIndexHTML(cheatsheets)
		content := bytes.NewReader([]byte(html))
		http.ServeContent(w, r, "foo.html", time.Time{}, content)
		return nil
	}
	csIndexDynamic := NewDynamicContent(csIndexMatches, csIndexSend)
	csDynamic := NewDynamicContent(csMatches, csSend)
	return []URLContent{csIndexDynamic, csDynamic}
}

func runServer() {
	staticFiles := []string{
		"/s/cheatsheet.css",
		"cheatsheet.css",

		"/s/cheatsheet.js",
		"cheatsheet.js",
	}
	var files []URLContent
	for i := 0; i < len(staticFiles); i += 2 {
		uri := staticFiles[i]
		path := staticFiles[i+1]
		f := NewFileOnDisk(path, uri)
		files = append(files, f)
	}
	cheatsheets := buildContentCheatsheets()
	files = append(files, cheatsheets...)

	serverFiles := &ServerFiles{
		Files: files,
	}
	waitFn := StartServer(serverFiles)
	waitFn()
}

func main() {
	var (
		flgRunServer bool
	)
	{
		flag.BoolVar(&flgRunServer, "run", false, "run me")
		flag.Parse()
	}
	if flgRunServer {
		runServer()
		return
	}
	flag.Usage()
}
