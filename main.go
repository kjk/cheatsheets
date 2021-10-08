package main

import (
	"bytes"
	"flag"
	"fmt"
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
		switch uri {
		case "/index.html", "/all.html":
			// no-op
		default:
			return nil
		}
		all := uri == "/all.html"
		send := func(w http.ResponseWriter, r *http.Request) {
			logf(ctx(), "csIndexSend: '%s'\n", uri)
			a := cheatsheets
			if !all {
				a = nil
				for _, cs := range cheatsheets {
					if cs.inMain {
						a = append(a, cs)
					}
				}
			}
			html := []byte(genIndexHTML(a))
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
		return []string{"/index.html", "/all.html"}
	}
	csIndexDynamic := server.NewDynamicHandler(csIndexMatches, csIndexURLS)
	csDynamic := server.NewDynamicHandler(csMatches, csURLS)
	return []server.Handler{csIndexDynamic, csDynamic}
}

func makeServerDynamic() *server.Server {
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

	srv := makeServerDynamic()
	httpSrv := makeHTTPServer(srv)
	logf(ctx(), "Starting server on http://%s'\n", httpSrv.Addr)
	if isWindows() {
		openBrowser(fmt.Sprintf("http://%s", httpSrv.Addr))
	}
	err := httpSrv.ListenAndServe()
	logf(ctx(), "runServerDynamic: httpSrv.ListenAndServe() returned '%s'\n", err)
}

func runServerProd() {
	printLoggingStats()
	panicIf(!dirExists(dirWwwGenerated))
	h := server.NewDirHandler(dirWwwGenerated, "/", nil)
	logf(ctx(), "runServerProd starting, hasSpacesCreds: %v, %d urls\n", hasSpacesCreds(), len(h.URLS()))

	closeHTTPLog := OpenHTTPLog("cheatsheets")
	defer closeHTTPLog()

	srv := &server.Server{
		Handlers:  []server.Handler{h},
		CleanURLS: true,
		Port:      httpPort,
	}
	httpSrv := makeHTTPServer(srv)
	logf(ctx(), "Starting server on http://%s'\n", httpSrv.Addr)
	if isWindows() {
		openBrowser(fmt.Sprintf("http://%s", httpSrv.Addr))
	}
	err := httpSrv.ListenAndServe()
	logf(ctx(), "runServerProd: httpSrv.ListenAndServe() returned '%s'\n", err)
}

func generateStatic() {
	timeStart := time.Now()
	defer func() {
		logf(ctx(), "generateStatic() finished in %s\n", formatDuration(time.Since(timeStart)))
	}()
	srv := makeServerDynamic()
	must(os.RemoveAll(dirWwwGenerated))

	nFiles := 0
	totalSize := int64(0)
	onWritten := func(path string, d []byte) {
		fsize := int64(len(d))
		totalSize += fsize
		sizeStr := formatSize(fsize)
		if nFiles%256 == 0 {
			logf(ctx(), "generateStatic: file %d '%s' of size %s\n", nFiles+1, path, sizeStr)
		}
		nFiles++
	}
	server.WriteServerFilesToDir(dirWwwGenerated, srv.Handlers, onWritten)
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
		flgRunServer     bool
		flgRunServerProd bool
		flgGen           bool
		flgDeploy        bool
	)
	{
		flag.BoolVar(&flgRunServer, "run", false, "run dev server")
		flag.BoolVar(&flgRunServerProd, "run-prod", false, "run prod server serving www_generated")
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

	if flgRunServerProd {
		runServerProd()
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
