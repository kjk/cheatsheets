package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kjk/common/server"
)

const (
	dirWwwGenerated = "www_generated"
	httpPort        = 9033
)

func logHTTPReqShort(r *http.Request, code int, size int64, dur time.Duration) {
	if strings.HasPrefix(r.URL.Path, "/ping") {
		return
	}
	if code >= 400 {
		// make 400 stand out more in logs
		logf(ctx(), "%s %d %s %s in %s\n", "   ", code, r.RequestURI, formatSize(size), dur)
	} else {
		logf(ctx(), "%s %d %s %s in %s\n", r.Method, code, r.RequestURI, formatSize(size), dur)
	}
	ref := r.Header.Get("Referer")
	if ref != "" && !strings.Contains(ref, r.Host) {
		logf(ctx(), "ref: %s \n", ref)
	}
}

func makeHTTPServer(srv *server.Server) *http.Server {
	panicIf(srv == nil, "must provide srv")
	httpPort := 8080
	if srv.Port != 0 {
		httpPort = srv.Port
	}
	httpAddr := fmt.Sprintf(":%d", httpPort)
	if isWindows() {
		httpAddr = "localhost" + httpAddr
	}

	mainHandler := func(w http.ResponseWriter, r *http.Request) {
		//logf(ctx(), "mainHandler: '%s'\n", r.RequestURI)
		timeStart := time.Now()
		cw := server.CapturingResponseWriter{ResponseWriter: w}

		defer func() {
			if p := recover(); p != nil {
				logf(ctx(), "mainHandler: panicked with with %v\n", p)
				http.Error(w, fmt.Sprintf("Error: %v", r), http.StatusInternalServerError)
				logHTTPReqShort(r, http.StatusInternalServerError, 0, time.Since(timeStart))
				LogHTTPReq(r, http.StatusInternalServerError, 0, time.Since(timeStart))
			} else {
				logHTTPReqShort(r, cw.StatusCode, cw.Size, time.Since(timeStart))
				LogHTTPReq(r, cw.StatusCode, cw.Size, time.Since(timeStart))
			}
		}()

		uri := r.URL.Path
		serve, _ := srv.FindHandler(uri)
		if serve != nil {
			serve(&cw, r)
			return
		}
		http.NotFound(&cw, r)
	}

	httpSrv := &http.Server{
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second, // introduced in Go 1.8
		Handler:      http.HandlerFunc(mainHandler),
	}
	httpSrv.Addr = httpAddr
	return httpSrv
}

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

	closeHTTPLog := OpenHTTPLog("cheatsheets")
	defer closeHTTPLog()

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
	h.TryServeCompressed = true
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
