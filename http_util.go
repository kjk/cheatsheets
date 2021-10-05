package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/kjk/cheatsheets/pkg/server"
	"github.com/kjk/siser"
)

func WriteServerFilesToDir(dir string, handlers []server.Handler) (int, int64) {
	nFiles := 0
	totalSize := int64(0)
	dirCreated := map[string]bool{}

	writeFile := func(uri string, d []byte) {
		name := strings.TrimPrefix(uri, "/")
		name = filepath.FromSlash(name)
		path := filepath.Join(dir, name)
		// optimize for writing lots of files
		// I assume that even a no-op os.MkdirAll()
		// might be somewhat expensive
		fileDir := filepath.Dir(path)
		if !dirCreated[fileDir] {
			must(os.MkdirAll(fileDir, 0755))
			dirCreated[fileDir] = true
		}
		err := os.WriteFile(path, d, 0644)
		must(err)
		fsize := int64(len(d))
		totalSize += fsize
		sizeStr := formatSize(fsize)
		if nFiles%256 == 0 {
			logf(ctx(), "WriteServerFilesToDir: file %d '%s' of size %s\n", nFiles+1, path, sizeStr)
		}
		nFiles++
	}
	server.IterContent(handlers, writeFile)
	return nFiles, totalSize
}

func recWriteNonEmpty(rec *siser.Record, k, v string) {
	if v != "" {
		rec.Write(k, v)
	}
}

func MakeHTTPServer(srv *server.Server) *http.Server {
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
		defer func() {
			if p := recover(); p != nil {
				logf(ctx(), "mainHandler: panicked with with %v\n", p)
				http.Error(w, fmt.Sprintf("Error: %v", r), http.StatusInternalServerError)
				logHTTPReq(r, http.StatusInternalServerError, 0, time.Since(timeStart))
				panic(p)
			}
		}()
		uri := r.URL.Path
		serve, _ := srv.FindHandler(uri)
		if serve == nil {
			http.NotFound(w, r)
			logHTTPReq(r, http.StatusNotFound, 0, time.Since(timeStart))
			return
		}
		if serve != nil {
			cw := server.CapturingResponseWriter{ResponseWriter: w}
			serve(&cw, r)
			logHTTPReq(r, cw.StatusCode, cw.Size, time.Since(timeStart))
			return
		}
		http.NotFound(w, r)
		logHTTPReq(r, http.StatusNotFound, 0, time.Since(timeStart))
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

// returns function that will wait for SIGTERM signal (e.g. Ctrl-C) and
// shutdown the server
func StartHTTPServer(httpSrv *http.Server) func() {
	logf(ctx(), "Starting server on http://%s'\n", httpSrv.Addr)
	if isWindows() {
		openBrowser(fmt.Sprintf("http://%s", httpSrv.Addr))
	}

	chServerClosed := make(chan bool, 1)
	go func() {
		err := httpSrv.ListenAndServe()
		// mute error caused by Shutdown()
		if err == http.ErrServerClosed {
			err = nil
		}
		must(err)
		logf(ctx(), "trying to shutdown HTTP server\n")
		chServerClosed <- true
	}()

	return func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt /* SIGINT */, syscall.SIGTERM)

		sig := <-c
		logf(ctx(), "Got signal %s\n", sig)

		if httpSrv != nil {
			go func() {
				// Shutdown() needs a non-nil context
				_ = httpSrv.Shutdown(ctx())
			}()
			select {
			case <-chServerClosed:
				// do nothing
				logf(ctx(), "server shutdown cleanly\n")
			case <-time.After(time.Second * 5):
				// timeout
				logf(ctx(), "server killed due to shutdown timeout\n")
			}
		}
	}
}

func StartServer(srv *server.Server) func() {
	httpSrv := MakeHTTPServer(srv)
	return StartHTTPServer(httpSrv)
}

func RunServerProd(srv *server.Server) {
	httpSrv := MakeHTTPServer(srv)
	logf(ctx(), "Starting server on http://%s'\n", httpSrv.Addr)
	if isWindows() {
		openBrowser(fmt.Sprintf("http://%s", httpSrv.Addr))
	}
	err := httpSrv.ListenAndServe()
	logf(ctx(), "RunServerProd: httpSrv.ListenAndServe() returned '%s'\n", err)
}

// requestGetReferrer returns a referer e.g. "https://presstige.io/dashboard"
func requestGetReferrer(r *http.Request) string {
	return r.Header.Get("referer")
}

func requestGetUserAgent(r *http.Request) string {
	return r.Header.Get("user-agent")
}

// requestGetRemoteAddress returns ip address of the client making the request,
// taking into account http proxies
func requestGetRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("x-real-ip")
	hdrForwardedFor := hdr.Get("x-forwarded-for")
	// Request.RemoteAddress contains port, which we want to remove i.e.:
	// "[::1]:58292" => "[::1]"
	ipAddrFromRemoteAddr := func(s string) string {
		idx := strings.LastIndex(s, ":")
		if idx == -1 {
			return s
		}
		return s[:idx]
	}
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// TODO: should return first non-local address
		return parts[0]
	}
	return hdrRealIP
}
