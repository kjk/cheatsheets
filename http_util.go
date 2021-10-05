package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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

var (
	httpLogRec siser.Record
	httpLogMu  sync.Mutex
)

func recWriteNonEmpty(rec *siser.Record, k, v string) {
	if v != "" {
		rec.Write(k, v)
	}
}

func logHTTPReq(r *http.Request, code int, size int64, dur time.Duration) {
	logf(ctx(), "%s %s %d in %s\n", r.Method, r.RequestURI, code, dur)
	httpLogMu.Lock()
	defer httpLogMu.Unlock()
	rec := &httpLogRec
	rec.Reset()
	rec.Write("req", fmt.Sprintf("%s %s %d", r.Method, r.RequestURI, code))
	recWriteNonEmpty(rec, "referer", requestGetReferrer(r))
	recWriteNonEmpty(rec, "host", r.Host)
	rec.Write("ipaddr", requestGetRemoteAddress(r))
	rec.Write("size", strconv.FormatInt(size, 10))
	durMs := int64(dur / time.Millisecond)
	rec.Write("duration", strconv.FormatInt(durMs, 10))
	rec.Write("ua", requestGetUserAgent(r))
	// TODO: write to log file
}

func MakeHTTPServer(srv *server.Server) *http.Server {
	panicIf(srv == nil, "must provide files")
	httpPort := 9210
	if srv.Port != 0 {
		httpPort = srv.Port
	}
	httpAddr := fmt.Sprintf(":%d", httpPort)
	if isWindows() {
		httpAddr = "localhost" + httpAddr
	}

	mainHandler := func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()
		defer func() {
			if p := recover(); p != nil {
				logf(ctx(), "mainHandler: panicked with with %v\n", p)
				http.Error(w, fmt.Sprintf("Error: %v", r), http.StatusInternalServerError)
				logHTTPReq(r, http.StatusInternalServerError, 0, time.Since(timeStart))
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
	httpServer := MakeHTTPServer(srv)
	return StartHTTPServer(httpServer)
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
