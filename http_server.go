package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// URLContent represents one or more urls and their content
type URLContent interface {
	// returns true if this URLContent matches url
	Matches(url string) bool
	// if Matches returns true, Get() returns the content of the file
	Get([]byte, error)
}

type FileOnDisk struct {
	// Path on disk for this file
	Path string
	// list of urls that this file matches
	URL []string
}

func (f *FileOnDisk) Matches(url string) bool {
	for _, u := range f.URL {
		// urls are case-insensitive
		if strings.EqualFold(u, url) {
			return true
		}
	}
	return false
}

func (f *FileOnDisk) Get() ([]byte, error) {
	return os.ReadFile(f.Path)
}

func NewFileOnDisk(path string, url string, urls ...string) *FileOnDisk {
	// early detection of problems
	panicIf(!fileExists(path), "file '%s' doesn't exist", path)
	res := &FileOnDisk{
		Path: path,
		URL:  []string{url},
	}
	res.URL = append(res.URL, urls...)
	return res
}

// ServerFiles represents all files known to the server
type ServerFiles struct {
	files []URLContent
}

func httpServeText(w http.ResponseWriter, r *http.Request, s string) {
	content := bytes.NewReader([]byte(s))
	http.ServeContent(w, r, "foo.txt", time.Time{}, content)
}

func handleFile(w http.ResponseWriter, r *http.Request, files *ServerFiles) {
	httpServeText(w, r, "this is a simple text reponse")
}

// returns function that will wait for SIGTERM signal (e.g. Ctrl-C) and
// shutdown the server
func StartServer(files *ServerFiles) func() {
	httpPort := 8080
	httpAddr := fmt.Sprintf(":%d", httpPort)
	if isWindows() {
		httpAddr = "localhost" + httpAddr
	}
	mux := &http.ServeMux{}
	handleAll := func(w http.ResponseWriter, r *http.Request) {
		handleFile(w, r, files)
	}
	mux.HandleFunc("/", handleAll)
	var handler http.Handler = mux
	httpSrv := &http.Server{
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second, // introduced in Go 1.8
		Handler:      handler,
	}
	httpSrv.Addr = httpAddr
	ctx := ctx()
	logf(ctx, "Starting server on http://%s'\n", httpAddr)
	if isWindows() {
		openBrowser(fmt.Sprintf("http://%s", httpAddr))
	}

	chServerClosed := make(chan bool, 1)
	go func() {
		err := httpSrv.ListenAndServe()
		// mute error caused by Shutdown()
		if err == http.ErrServerClosed {
			err = nil
		}
		must(err)
		logf(ctx, "trying to shutdown HTTP server\n")
		chServerClosed <- true
	}()

	return func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt /* SIGINT */, syscall.SIGTERM)

		sig := <-c
		logf(ctx, "Got signal %s\n", sig)

		if httpSrv != nil {
			go func() {
				// Shutdown() needs a non-nil context
				_ = httpSrv.Shutdown(ctx)
			}()
			select {
			case <-chServerClosed:
				// do nothing
				logf(ctx, "server shutdown cleanly\n")
			case <-time.After(time.Second * 5):
				// timeout
				logf(ctx, "server killed due to shutdown timeout\n")
			}
		}
	}
}
