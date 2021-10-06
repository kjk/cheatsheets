package main

import (
	"bytes"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kjk/cheatsheets/pkg/server"
)

const (
	dirWwwGenerated = "www_generated"
	httpPort        = 9033
)

func buildContentCheatsheets() []server.Handler {
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
	csIndexDynamic := server.NewDynamicHandler(csIndexMatches, csIndexURLS)
	csDynamic := server.NewDynamicHandler(csMatches, csURLS)
	return []server.Handler{csIndexDynamic, csDynamic}
}

func buildServerDynamic() *server.Server {
	staticFiles := []string{
		"/s/cheatsheet.css",
		"cheatsheet.css",

		"/s/cheatsheet.js",
		"cheatsheet.js",

		"/404.html",
		"404.html",

		"/ping.txt",
		"ping.txt",
	}
	for i := 0; i < len(staticFiles); i += 2 {
		name := staticFiles[i+1]
		staticFiles[i+1] = filepath.Join("www", name)
	}
	h := server.NewFilesHandler(staticFiles...)
	handlers := []server.Handler{h}
	cheatsheets := buildContentCheatsheets()
	handlers = append(handlers, cheatsheets...)

	return &server.Server{
		Handlers:  handlers,
		CleanURLS: true,
		Port:      httpPort,
	}
}

func runServerDynamic() {
	printLoggingStats()
	logf(ctx(), "runServerDynamic starting\n")

	srv := buildServerDynamic()

	closeHTTPLog := openHTTPLog()
	defer closeHTTPLog()

	waitFn := StartServer(srv)
	waitFn()
}

func runServerStatic() {
	printLoggingStats()
	logf(ctx(), "runServerStatic starting\n")
	panicIf(!dirExists(dirWwwGenerated))
	h := server.NewDirHandler(dirWwwGenerated, "/", nil)
	logf(ctx(), "  %d urls\n", len(h.URLS()))
	srv := &server.Server{
		Handlers:  []server.Handler{h},
		CleanURLS: true,
		Port:      httpPort,
	}
	closeHTTPLog := openHTTPLog()
	defer closeHTTPLog() // TODO: this actually doesn't take in prod
	RunServerProd(srv)
}

func generateStatic() {
	timeStart := time.Now()
	defer func() {
		logf(ctx(), "generateStatic() finished in %s\n", formatDuration(time.Since(timeStart)))
	}()
	srv := buildServerDynamic()
	must(os.RemoveAll(dirWwwGenerated))
	WriteServerFilesToDir(dirWwwGenerated, srv.Handlers)
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
		flgRunServer       bool
		flgRunServerStatic bool
		flgGen             bool
		flgDeploy          bool
	)
	{
		flag.BoolVar(&flgRunServer, "run", false, "run dev server")
		flag.BoolVar(&flgRunServerStatic, "run-static", false, "run prod server serving www_generated")
		flag.BoolVar(&flgGen, "gen", false, "generate static files in www_generated dir")
		flag.BoolVar(&flgDeploy, "deploy", false, "deploy to render.com")
		flag.Parse()
	}

	if false {
		compareCompr()
		return
	}

	if flgRunServer {
		runServerDynamic()
		return
	}

	if flgRunServerStatic {
		runServerStatic()
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
