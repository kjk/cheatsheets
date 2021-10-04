package main

import (
	"bytes"
	"flag"
	"net/http"
	"os"
	"strings"
	"time"
)

func buildContentCheatsheets() []Handler {
	cheatsheets := readCheatSheets()
	csFindByURL := func(uri string) *cheatSheet {
		// match /cheatsheet/go.html => go
		uriBase := strings.ToLower(strings.TrimPrefix(uri, "/cheatsheet/"))
		if len(uri) == len(uriBase) {
			// doesn't start with /cheatsheet
			logvf(ctx(), "csFindByURL: no match for '%s because doesn't start with /cheatsheet/'\n", uri)
			return nil
		}
		uriBaseNoExt := strings.TrimSuffix(uriBase, ".html")
		if len(uriBase) == len(uriBaseNoExt) {
			// doens't end with .html
			logf(ctx(), "csFindByURL: no match for '%s' because doesn't end with .html\n", uri)
			return nil
		}
		logvf(ctx(), "csFindByURL: looking for '%s'\n", uriBaseNoExt)
		for _, cs := range cheatsheets {
			if uriBaseNoExt == cs.fileNameBase {
				logvf(ctx(), "csFindByURL: found match for '%s'\n", uri)
				return cs
			}
		}
		logvf(ctx(), "csFindByURL: no match for '%s'\n", uri)
		return nil
	}
	csMatches := func(uri string) func(w http.ResponseWriter, r *http.Request) {
		cs := csFindByURL(uri)
		send := func(w http.ResponseWriter, r *http.Request) {
			panicIf(cs == nil, "no match for '%s'", uri)
			processCheatSheet(cs)
			html := genCheatsheetHTML(cs)
			if r == nil {
				w.Write(html)
				return
			}
			content := bytes.NewReader(html)
			http.ServeContent(w, r, "foo.html", time.Time{}, content)
		}
		if cs == nil {
			return nil
		}
		return send
	}
	csURLS := func() []string {
		var res []string
		for _, cs := range cheatsheets {
			uri := "/cheatsheet/" + cs.fileNameBase + ".html"
			res = append(res, uri)
		}
		return res
	}

	csIndexMatches := func(uri string) func(w http.ResponseWriter, r *http.Request) {
		if uri != "/index.html" {
			return nil
		}
		send := func(w http.ResponseWriter, r *http.Request) {
			logf(ctx(), "csIndexSend: '%s'\n", uri)
			html := []byte(genIndexHTML(cheatsheets))
			if r == nil {
				w.Write(html)
				return
			}
			content := bytes.NewReader(html)
			http.ServeContent(w, r, "foo.html", time.Time{}, content)
		}
		return send
	}
	csIndexURLS := func() []string {
		return []string{"/index.html"}
	}
	csIndexDynamic := NewDynamicHandler(csIndexMatches, csIndexURLS)
	csDynamic := NewDynamicHandler(csMatches, csURLS)
	return []Handler{csIndexDynamic, csDynamic}
}

func buildServerFiles() *ServerConfig {
	staticFiles := []string{
		"/s/cheatsheet.css",
		"cheatsheet.css",

		"/s/cheatsheet.js",
		"cheatsheet.js",

		"/404.html",
		"404.html",
	}
	var handlers []Handler
	for i := 0; i < len(staticFiles); i += 2 {
		uri := staticFiles[i]
		path := staticFiles[i+1]
		h := NewFilesHandler(uri, path)
		handlers = append(handlers, h)
	}
	cheatsheets := buildContentCheatsheets()
	handlers = append(handlers, cheatsheets...)

	return &ServerConfig{
		Handlers:  handlers,
		CleanURLS: true,
		Port:      9033,
	}
}

func runServer() {
	printLoggingStats()
	logf(ctx(), "runServer starting\n")
	waitFn := StartServer(buildServerFiles())
	waitFn()
}

func generateStatic() {
	timeStart := time.Now()
	defer func() {
		logf(ctx(), "generateStatic() finished in %s\n", formatDuration(time.Since(timeStart)))
	}()
	sf := buildServerFiles()
	WriteServerFilesToDir("www_generated", sf.Handlers)
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
