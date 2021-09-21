package main

import (
	"bytes"
	"flag"
	"net/http"
	"os"
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
	csSend := func(w http.ResponseWriter, r *http.Request, uri string) error {
		if uri == "" {
			uri = r.URL.Path
		}
		cs := csFindByURL(uri)
		panicIf(cs == nil, "no match for '%s'", uri)
		html := genCheatsheetHTML(cs)
		content := bytes.NewReader(html)
		http.ServeContent(w, r, "foo.html", time.Time{}, content)
		return nil
	}
	csContent := func() []*Content {
		var res []*Content
		for _, cs := range cheatsheets {
			uri := "/cheatsheet/" + cs.fileNameBase + ".html"
			d := genCheatsheetHTML(cs)
			res = append(res, &Content{
				URL:     uri,
				Content: d,
			})
		}
		return res
	}

	csIndexMatches := func(uri string) bool {
		matches := uri == "/index.html"
		if matches {
			logf(ctx(), "csIndexMatches: match for '%s'\n", uri)
		}
		return matches
	}
	csIndexSend := func(w http.ResponseWriter, r *http.Request, uri string) error {
		logf(ctx(), "csIndexSend: '%s'\n", r.URL)
		if uri == "" {
			uri = r.URL.Path
		}
		panicIf(!csIndexMatches(uri), "no match for '%s'", uri)
		html := genIndexHTML(cheatsheets)
		content := bytes.NewReader([]byte(html))
		http.ServeContent(w, r, "foo.html", time.Time{}, content)
		return nil
	}
	csIndexContent := func() []*Content {
		var res []*Content
		d := genIndexHTML(cheatsheets)
		res = append(res, &Content{
			URL:     "/index.html",
			Content: []byte(d),
		})
		return res
	}
	csIndexDynamic := NewDynamicContent(csIndexMatches, csIndexSend, csIndexContent)
	csDynamic := NewDynamicContent(csMatches, csSend, csContent)
	return []URLContent{csIndexDynamic, csDynamic}
}

func buildServerFiles() *ServerConfig {
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

	return &ServerConfig{
		Files:     files,
		CleanURLS: true,
	}
}

func runServer() {
	waitFn := StartServer(buildServerFiles())
	waitFn()
}

func generateStatic() {
	timeStart := time.Now()
	defer func() {
		logf(ctx(), "generateStatic() finished in %s\n", formatDuration(time.Since(timeStart)))
	}()
	sf := buildServerFiles()
	WriteServerFilesToDir("www_generated", sf.Files)
}

func deployToRender() {
	deployURL := os.Getenv("CHEATSHEETS_DEPLOY_HOOK")
	panicIf(deployURL == "", "need env variable CHEATSHEETS_DEPLOY_HOOK")
	d, err := httpGet(deployURL)
	must(err)
	logf(ctx(), "deployed to render.com:\n%s\n", string(d))
}

func main() {
	var (
		flgRunServer bool
		flgGen       bool
		flgDeploy    bool
	)
	{
		flag.BoolVar(&flgRunServer, "run", false, "run dev server")
		flag.BoolVar(&flgGen, "gen", false, "generate static files in www_generated dir")
		flag.BoolVar(&flgDeploy, "deploy", false, "deploy to render.com")
		flag.Parse()
	}
	if flgRunServer {
		runServer()
		return
	}
	if flgGen {
		generateStatic()
		return
	}
	if flgDeploy {
		deployToRender()
		return
	}

	flag.Usage()
}
