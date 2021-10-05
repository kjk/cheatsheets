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
	server.IterHandlers(handlers, writeFile)
	return nFiles, totalSize
}

func MakeHTTPServer(srv *server.Server) *http.Server {
	panicIf(srv == nil, "must provide files")
	httpPort := 8080
	if srv.Port != 0 {
		httpPort = srv.Port
	}
	httpAddr := fmt.Sprintf(":%d", httpPort)
	if isWindows() {
		httpAddr = "localhost" + httpAddr
	}
	httpSrv := &http.Server{
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second, // introduced in Go 1.8
		Handler:      srv,
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
				//logf(ctx(), "server shutdown cleanly\n")
			case <-time.After(time.Second * 5):
				// timeout
				//logf(ctx(), "server killed due to shutdown timeout\n")
			}
		}
	}
}

func StartServer(srv *server.Server) func() {
	httpServer := MakeHTTPServer(srv)
	return StartHTTPServer(httpServer)
}
