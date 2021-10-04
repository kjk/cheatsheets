package main

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// ServerConfig represents all files known to the server
type ServerConfig struct {
	Handlers  []Handler
	CleanURLS bool
	Port      int
}

type HandlerFunc = func(w http.ResponseWriter, r *http.Request)
type GetHandlerFunc = func(string) func(w http.ResponseWriter, r *http.Request)

// Handler represents one or more urls and their content
type Handler interface {
	// returns a handler for this url
	// if nil, doesn't handle this url
	Get(url string) HandlerFunc
	// get all urls handled by this Handler
	// useful for e.g. saving a static copy to disk
	URLS() []string
}

func panicIfAbsoluteURL(uri string) {
	panicIf(strings.Contains(uri, "://"), "got absolute url '%s'", uri)
}

// FileWriter implements http.ResponseWriter interface for writing to a file
type FileWriter struct {
	w io.Writer
}

func (w *FileWriter) Header() http.Header {
	return nil
}

func (w *FileWriter) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w *FileWriter) WriteHeader(statusCode int) {
	// no-op
}

func serveFile(w http.ResponseWriter, r *http.Request, path string) {
	if r == nil {
		d := readFileMust(path)
		_, err := w.Write(d)
		must(err)
		return
	}
	http.ServeFile(w, r, path)
}

// d can be nil, in which case no caching
func serve404FileCached(w http.ResponseWriter, r *http.Request, path string, cached *[]byte) {
	var d []byte
	if cached != nil && len(*cached) > 0 {
		d = *cached
	} else {
		var err error
		d, err = os.ReadFile(path)
		must(err)
		if cached != nil {
			*cached = d
		}
	}
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype != "" {
		w.Header().Set("Content-Type", ctype)
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusNotFound)
	w.Write(d)
}

func makeServeFile(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(path, "404.html") {
			serve404FileCached(w, r, path, nil)
			return
		}
		serveFile(w, r, path)
	}
}

// uri is only used to guess content type
func serveContent(w http.ResponseWriter, r *http.Request, uri string, d []byte) {
	if r == nil {
		_, err := w.Write(d)
		must(err)
		return
	}
	content := bytes.NewReader(d)
	http.ServeContent(w, r, uri, time.Now(), content)
}

func makeServeContent(uri string, d []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		serveContent(w, r, uri, d)
	}
}

type FilesHandler struct {
	files map[string]string // maps url to a path on disk
}

func (h *FilesHandler) AddFile(uri, path string) {
	panicIfAbsoluteURL(uri)
	panicIf(!fileExists(path), "file '%s' doesn't exist", path)
	h.files[uri] = path
}

func (h *FilesHandler) AddFilesInDir(dir string, uriPrefix string, files []string) {
	for _, f := range files {
		uri := uriPrefix + f
		path := filepath.Join(dir, f)
		h.AddFile(uri, path)
	}
}

func (h *FilesHandler) Get(url string) func(w http.ResponseWriter, r *http.Request) {
	for uri, path := range h.files {
		// urls are case-insensitive
		// TODO: are they?
		if strings.EqualFold(uri, url) {
			return makeServeFile(path)
		}
	}
	return nil
}

func (h *FilesHandler) URLS() []string {
	urls := []string{}
	for uri := range h.files {
		urls = append(urls, uri)
	}
	return urls
}

// files is: uri1, path1, uri2, path2, ...
func NewFilesHandler(files ...string) *FilesHandler {
	panicIf(len(files)%2 == 1)
	n := len(files)
	h := &FilesHandler{
		files: map[string]string{},
	}
	for i := 0; i < n; i += 2 {
		h.AddFile(files[i], files[i+1])
	}
	return h
}

type DirHandler struct {
	Dir       string
	URLPrefix string

	URL   []string
	paths []string // same order as URL
}

func (h *DirHandler) Get(url string) func(w http.ResponseWriter, r *http.Request) {
	for i, u := range h.URL {
		// urls are case-insensitive
		if strings.EqualFold(u, url) {
			return makeServeFile(h.paths[i])
		}
	}
	return nil
}

func (h *DirHandler) URLS() []string {
	return h.URL
}

func getURLSForFiles(startDir string, urlPrefix string, acceptFile func(string) bool) (urls []string, paths []string) {
	filepath.WalkDir(startDir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if acceptFile != nil && !acceptFile(filePath) {
			return nil
		}
		dir := strings.TrimPrefix(filePath, startDir)
		dir = filepath.ToSlash(dir)
		dir = strings.TrimPrefix(dir, "/")
		uri := path.Join(urlPrefix, dir)
		//logf(ctx(), "getURLSForFiles: dir: '%s'\n", dir)
		urls = append(urls, uri)
		paths = append(paths, filePath)
		return nil
	})
	return
}

func NewDirHandler(dir string, urlPrefix string, acceptFile func(string) bool) *DirHandler {
	urls, paths := getURLSForFiles(dir, urlPrefix, acceptFile)
	return &DirHandler{
		Dir:       dir,
		URLPrefix: urlPrefix,
		URL:       urls,
		paths:     paths,
	}
}

type ContentHandler struct {
	files map[string][]byte
}

func (h *ContentHandler) Get(uri string) func(http.ResponseWriter, *http.Request) {
	for u, body := range h.files {
		if uri == u {
			return makeServeContent(uri, body)
		}
	}
	return nil
}

func (h *ContentHandler) URLS() []string {
	urls := []string{}
	for u := range h.files {
		urls = append(urls, u)
	}
	return urls
}

func (h *ContentHandler) Add(uri string, body []byte) {
	panicIfAbsoluteURL(uri)
	h.files[uri] = body
}

func NewContentHandler(uri string, d []byte) *ContentHandler {
	h := &ContentHandler{
		files: map[string][]byte{},
	}
	h.Add(uri, d)
	return h
}

type DynamicHandler struct {
	get  GetHandlerFunc
	urls func() []string
}

func (h *DynamicHandler) Get(uri string) func(http.ResponseWriter, *http.Request) {
	return h.get(uri)
}

func (h *DynamicHandler) URLS() []string {
	return h.urls()
}

func NewDynamicHandler(get GetHandlerFunc, urls func() []string) *DynamicHandler {
	return &DynamicHandler{
		get:  get,
		urls: urls,
	}
}

type InMemoryFilesHandler struct {
	files map[string][]byte
}

func (h *InMemoryFilesHandler) Get(uri string) func(http.ResponseWriter, *http.Request) {
	for path, d := range h.files {
		if strings.EqualFold(path, uri) {
			return makeServeContent(uri, d)
		}
	}
	return nil
}

func (h *InMemoryFilesHandler) URLS() []string {
	var urls []string
	for path := range h.files {
		urls = append(urls, path)
	}
	return urls
}

func NewInMemoryFilesHandler(files map[string][]byte) *InMemoryFilesHandler {
	for path, d := range files {
		newp := strings.Replace(path, "\\", "/", -1)
		if !strings.HasPrefix(newp, "/") {
			newp = "/" + newp
		}
		if path == newp {
			continue
		}
		files[newp] = d
		delete(files, path)
	}

	return &InMemoryFilesHandler{
		files: files,
	}
}

func iterHandlers(handlers []Handler, fn func(uri string, d []byte)) {
	var buf bytes.Buffer
	for _, h := range handlers {
		urls := h.URLS()
		for _, uri := range urls {
			buf.Reset()
			fw := &FileWriter{
				w: &buf,
			}
			serve := h.Get(uri)
			panicIf(serve == nil, "must have a handler for '%s'", uri)
			serve(fw, nil)
		}
	}
}

func WriteServerFilesToDir(dir string, handlers []Handler) (int, int64) {
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
	iterHandlers(handlers, writeFile)
	return nFiles, totalSize
}

func zipWriteFile(zw *zip.Writer, name string, data []byte) error {
	fw, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = fw.Write(data)
	return err
}

func WriteServerFilesToZip(handlers []Handler) int {
	nFiles := 0

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	zw.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})
	writeFile := func(uri string, d []byte) {
		name := strings.TrimPrefix(uri, "/")
		err := zipWriteFile(zw, name, d)
		must(err)
		sizeStr := formatSize(int64(len(d)))
		if nFiles%128 == 0 {
			logf(ctx(), "WriteServerFilesToZip: %d file '%s' of size %s\n", nFiles+1, name, sizeStr)
		}
		nFiles++
	}
	iterHandlers(handlers, writeFile)
	return nFiles
}

type CodeCaptureWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *CodeCaptureWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (s *ServerConfig) FindHandlerExact(uri string) HandlerFunc {
	for _, h := range s.Handlers {
		if send := h.Get(uri); send != nil {
			return send
		}
	}
	return nil
}

func commonExt(uri string) bool {
	ext := strings.ToLower(filepath.Ext(uri))
	switch ext {
	case ".html", ".js", ".css", ".txt", ".xml":
		return true
	}
	return false
}

func gen404Candidates(uri string) []string {
	parts := strings.Split(uri, "/")
	n := len(parts)
	for n > 0 {
		n = len(parts) - 1
		if parts[n] != "" {
			break
		}
		parts = parts[:n]
	}
	var res []string
	for len(parts) > 0 {
		s := strings.Join(parts, "/") + "/404.html"
		res = append(res, s)
		parts = parts[:len(parts)-1]
	}
	res = append(res, "/404.html")
	return res
}

func (s *ServerConfig) FindHandler(uri string) HandlerFunc {
	if h := s.FindHandlerExact(uri); h != nil {
		return h
	}
	// if we support clean urls, try find "/foo.html" for "/foo"
	if s.CleanURLS && !commonExt(uri) {
		if h := s.FindHandlerExact(uri + ".html"); h != nil {
			return h
		}
	}
	// try 404.html
	a := gen404Candidates(uri)
	for _, uri404 := range a {
		if h := s.FindHandlerExact(uri404); h != nil {
			return h
		}
	}
	return nil
}

func (s *ServerConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path
	if strings.HasSuffix(uri, "/") {
		uri = "/index.html"
	}
	serve := s.FindHandler(uri)
	if serve != nil {
		cw := CodeCaptureWriter{ResponseWriter: w}
		serve(&cw, r)
		logreq(r, cw.statusCode)
		return
	}
	http.NotFound(w, r)
	logreq(r, http.StatusNotFound)
}

func MakeHTTPServer(server *ServerConfig) *http.Server {
	panicIf(server == nil, "must provide files")
	httpPort := 8080
	if server.Port != 0 {
		httpPort = server.Port
	}
	httpAddr := fmt.Sprintf(":%d", httpPort)
	if isWindows() {
		httpAddr = "localhost" + httpAddr
	}
	httpSrv := &http.Server{
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second, // introduced in Go 1.8
		Handler:      server,
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

func StartServer(server *ServerConfig) func() {
	httpServer := MakeHTTPServer(server)
	return StartHTTPServer(httpServer)
}
