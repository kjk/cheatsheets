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
)

type Content struct {
	URL     string
	Content []byte
}

// URLContent represents one or more urls and their content
type URLContent interface {
	// returns true if this URLContent matches url
	Matches(url string) bool
	// if Matches returns true call Send() to send the output
	// this allows doing things like redirects
	Send(w http.ResponseWriter, r *http.Request) error
	// get all content. useful for e.g. saving a static copy to disk
	Content() []*Content
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

func (f *FileOnDisk) Send(w http.ResponseWriter, r *http.Request) error {
	panicIf(!fileExists(f.Path), "file '%s' doesn't exist")
	http.ServeFile(w, r, f.Path)
	return nil
}

func (f *FileOnDisk) Content() []*Content {
	d := readFileMust(f.Path)
	var res []*Content
	for _, uri := range f.URL {
		res = append(res, &Content{
			URL:     uri,
			Content: d,
		})
	}
	return res
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

type DynamicContent struct {
	matches func(string) bool
	send    func(http.ResponseWriter, *http.Request) error
	content func() []*Content
}

func (f *DynamicContent) Matches(uri string) bool {
	return f.matches(uri)
}

func (f *DynamicContent) Send(w http.ResponseWriter, r *http.Request) error {
	return f.send(w, r)
}

func (f *DynamicContent) Content() []*Content {
	return f.content()
}

func NewDynamicContent(matches func(string) bool, send func(w http.ResponseWriter, r *http.Request) error, content func() []*Content) *DynamicContent {
	return &DynamicContent{
		matches: matches,
		send:    send,
		content: content,
	}
}

func WriteServerFilesToDir(dir string, files []URLContent) {
	for _, f := range files {
		all := f.Content()
		for _, c := range all {
			path := filepath.Join(dir, c.URL)
			d := c.Content
			must(createDirForFile(path))
			must(os.WriteFile(path, d, 0644))
			sizeStr := formatSize(int64(len(d)))
			logf(ctx(), "WriteServerFilesToDir: '%s' of size %s\n", path, sizeStr)
		}
	}
}

// ServerFiles represents all files known to the server
type ServerFiles struct {
	Files []URLContent
}

// returns function that will wait for SIGTERM signal (e.g. Ctrl-C) and
// shutdown the server
func StartServer(files *ServerFiles) func() {
	panicIf(files == nil, "must provide files")
	httpPort := 8080
	httpAddr := fmt.Sprintf(":%d", httpPort)
	if isWindows() {
		httpAddr = "localhost" + httpAddr
	}
	mux := &http.ServeMux{}
	handleAll := func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.Path
		if uri == "/" {
			uri = "/index.html"
		}
		for _, f := range files.Files {
			if f.Matches(uri) {
				logf(ctx(), "handleFile: found match for '%s'\n", r.URL)
				f.Send(w, r)
				return
			}
		}
		logf(ctx(), "handleFile: no match for '%s'\n", r.URL)
		http.NotFound(w, r)
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
