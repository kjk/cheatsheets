---
title: Go Snippets
category: Go
---

# Basics

## Intro

A collection of Go code snippets that I use often in my programs.

## must

```go
func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
```

## panicIf

```go
func panicIf(cond bool, arg ...interface{}) {
	if !cond {
		return
	}
	s := "condition failed"
	if len(arg) > 0 {
		s = fmt.Sprintf("%s", arg[0])
		if len(arg) > 1 {
			s = fmt.Sprintf(s, arg[1:]...)
		}
	}
	panic(s)
}
```

## ctx
A shortcut for `context.Background()`. Otherwise using [logf](#logf) would be annoying.

```go
func ctx() context.Context {
    return context.Background()
}
```

## logf

```go
func logf(ctx context.Context, s string, arg ...interface{}) {
	if len(arg) > 0 {
		s = fmt.Sprintf(s, arg...)
	}
	fmt.Print(s)
}
```

In this implementation `ctx` is unused but I have implementation that uses it and I prefer to have the same function in all my code to make re-use of code easier.

## logIfErr

```go
func logIfErr(err error) {
	if err != nil {
		logf(err.Error())
	}
}
```

## isWindows

```go
func isWindows() bool {
	return strings.Contains(runtime.GOOS, "windows")
}
```


# Strings

## normalizeNewlines

Convert Windows (CRLF) and Mac (CF) newlines to Unix (LF). Optimized for speed,
modifies d in-place.

```go
func normalizeNewlinesInPlace(d []byte) []byte {
	wi := 0
	n := len(d)
	for i := 0; i < n; i++ {
		c := d[i]
		// 13 is CR
		if c != 13 {
			d[wi] = c
			wi++
			continue
		}
		// replace CR (mac / win) with LF (unix)
		d[wi] = 10
		wi++
		if i < n-1 && d[i+1] == 10 {
			// this was CRLF, so skip the LF
			i++
		}

	}
	return d[:wi]
}

func normalizeNewlinesInPlace(d []byte) []byte {
	d = append([]byte{}, d...)
	return normalizeNewlinesInPlace(d)
}
```

## sliceRmoveDuplicateStrings

```go
// sliceRemoveDuplicateStrings removes duplicate strings from an array of strings.
// It's optimized for the case of no duplicates. It modifes a in place.
func sliceRemoveDuplicateStrings(a []string) []string {
    if len(a) < 2 {
        return a
    }
    sort.Strings(a)
    writeIdx := 1
    for i := 1; i < len(a); i++ {
        if a[i-1] == a[i] {
            continue
        }
        if writeIdx != i {
            a[writeIdx] = a[i]
        }
        writeIdx++
    }
    return a[:writeIdx]
}
```

## stringInSlice

Return true if a string `toCheck` exists in string slice `a`.

```go
func stringInSlice(a []string, toFind string) bool {
	for _, s := range a {
		if s == toFind {
			return true
		}
	}
	return false
}
```

# Files

## expandTildeInPath

Given `~/foo`, will replace `~` with home directory.

```go
func expandTildeInPath(s string) string {
	if strings.HasPrefix(s, "~") {
		dir, err := os.UserHomeDir()
		must(err)
		return dir + s[1:]
	}
	return s
}
```

## pathExists

Returns true if path exists.

```go
func pathExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
```

[os.Lstat](https://pkg.go.dev/os#Lstat)

## fileExists

Returns true if path exists and is a regular file.

```go
func fileExists(path string) bool {
	st, err := os.Lstat(path)
	return err == nil && st.Mode().IsRegular()
}
```

[os.Lstat](https://pkg.go.dev/os#Lstat)

## dirExists

Returns true if path exists and is a directory.

```go
func dirExists(path string) bool {
	st, err := os.Lstat(path)
	return err == nil && st.IsDir()
}
```

[os.Lstat](https://pkg.go.dev/os#Lstat)

## getFileSize

Returns size of the file or `-1` if file doesn't exist. Sometimes checking
for `-1` is easier than checking for error.

```go
func getFileSize(path string) int64 {
	st, err := os.Lstat(path)
	if err == nil {
		return st.Size()
	}
	return -1
}
```

[os.Lstat](https://pkg.go.dev/os#Lstat)

## copyFile

Copy file, ensures to create a destination directory.

```go
func copyFile(dst string, src string) error {
	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}
	fin, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fin.Close()
	fout, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(fout, fin)
	err2 := fout.Close()
	if err != nil || err2 != nil {
		os.Remove(dst)
	}

	return err
}
```

Uses: [os.Remove](https://pkg.go.dev/os#Remove), [os.MkdirAll](https://pkg.go.dev/os#MkdirAll), [os.Open](https://pkg.go.dev/os#Open), [os.Create](https://pkg.go.dev/os#Create)

## createDirForFile

When you create a file in a directory that doesn't exist, it fails. This creates a directory for a file.

```go
func createDirForFile(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}
```

## readGzippedFile

Reads a gzip-compressed file (typically .gz)

```go
func readGzippedFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	gr, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	return ioutil.ReadAll(gr)
}
```

## readFileLines

Reads file content as array of lines.

```go
func readFileLines(filePath string) ([]string, error) {
    file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    res := make([]string, 0)
    for scanner.Scan() {
        line := scanner.Bytes()
        res = append(res, string(line))
    }
    if err = scanner.Err(); err != nil {
        return nil, err
    }
    return res, nil
}
```

## sha1HexOfFile

Returns sha1 of the file content in hex form.

```go
func sha1OfFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := sha1.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func sha1HexOfFile(path string) (string, error) {
	sha1, err := sha1OfFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha1), nil
}
```

## unzipToDir

Unzips a file to a directory.

```go
func recreateDir(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

func createDirForFile(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0755)
}

func unzipFile(f *zip.File, dstPath string) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	err = createDirForFile(dstPath)
	if err != nil {
		return err
	}

	w, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	if err != nil {
		w.Close()
		os.Remove(dstPath)
		return err
	}
	err = w.Close()
	if err != nil {
		os.Remove(dstPath)
		return err
	}
	return nil
}

func unzipToDir(zipPath string, destDir string) error {
	st, err := os.Stat(zipPath)
	if err != nil {
		return err
	}
	fileSize := st.Size()
	f, err := os.Open(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zr, err := zip.NewReader(f, fileSize)
	if err != nil {
		return err
	}
	err = recreateDir(destDir)
	if err != nil {
		return err
	}

	for _, fi := range zr.File {
		if fi.FileInfo().IsDir() {
			continue
		}
		destPath := filepath.Join(destDir, fi.Name)
		err = unzipFile(fi, destPath)
		if err != nil {
			os.RemoveAll(destDir)
			return err
		}
	}
	return nil
}
```

# HTTP

## httpGet

Download the content of URL.

```go
func httpGet(url string) ([]byte, error) {
    // default timeout for http.Get() is really long, so dial it down
    // for both connection and read/write timeouts
    timeoutClient := newTimeoutClient(time.Second*120, time.Second*120)
    resp, err := timeoutClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return nil, errors.New(fmt.Sprintf("'%s': status code not 200 (%d)", url, resp.StatusCode))
    }
    return ioutil.ReadAll(resp.Body)
}

// can be used for http.Get() requests with better timeouts. New one must be created
// for each Get() request
func newTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {
	timeoutDialer := func(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
		return func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, cTimeout)
			if err != nil {
				return nil, err
			}
			conn.SetDeadline(time.Now().Add(rwTimeout))
			return conn, nil
		}
	}

	return &http.Client{
		Transport: &http.Transport{
			Dial:  timeoutDialer(connectTimeout, readWriteTimeout),
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

```

## httpPost

Sends `body` data as POST request. This sends raw data. Use [httpPostMultiPart](#httppostmultipart) if you want to send as multipart encoding.

```go
func httpPost(uri string, body []byte) ([]byte, error) {
	// default timeout for http.Get() is really long, so dial it down
	// for both connection and read/write timeouts
	timeoutClient := newTimeoutClient(time.Second*120, time.Second*120)
	resp, err := timeoutClient.Post(uri, "", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("'%s': status code not 200 (%d)", uri, resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

// can be used for http.Get() requests with better timeouts. New one must be created
// for each Get() request
func newTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {
	timeoutDialer := func(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
		return func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, cTimeout)
			if err != nil {
				return nil, err
			}
			conn.SetDeadline(time.Now().Add(rwTimeout))
			return conn, nil
		}
	}

	return &http.Client{
		Transport: &http.Transport{
			Dial:  timeoutDialer(connectTimeout, readWriteTimeout),
			Proxy: http.ProxyFromEnvironment,
		},
	}
}
```

## httpPostMultiPart

```go
func httpPostMultiPart(uri string, files map[string]string) ([]byte, error) {
	contentType, body, err := createMultiPartForm(files)
	if err != nil {
		return nil, err
	}
	// default timeout for http.Get() is really long, so dial it down
	// for both connection and read/write timeouts
	timeoutClient := newTimeoutClient(time.Second*120, time.Second*120)
	resp, err := timeoutClient.Post(uri, contentType, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("'%s': status code not 200 (%d)", uri, resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func createMultiPartForm(form map[string]string) (string, io.Reader, error) {
	body := new(bytes.Buffer)
	mp := multipart.NewWriter(body)
	defer mp.Close()
	for key, val := range form {
		if strings.HasPrefix(val, "@") {
			val = val[1:]
			file, err := os.Open(val)
			if err != nil {
				return "", nil, err
			}
			defer file.Close()
			part, err := mp.CreateFormFile(key, val)
			if err != nil {
				return "", nil, err
			}
			io.Copy(part, file)
		} else {
			mp.WriteField(key, val)
		}
	}
	return mp.FormDataContentType(), body, nil
}

// can be used for http.Get() requests with better timeouts. New one must be created
// for each Get() request
func newTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {
	timeoutDialer := func(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
		return func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, cTimeout)
			if err != nil {
				return nil, err
			}
			conn.SetDeadline(time.Now().Add(rwTimeout))
			return conn, nil
		}
	}

	return &http.Client{
		Transport: &http.Transport{
			Dial:  timeoutDialer(connectTimeout, readWriteTimeout),
			Proxy: http.ProxyFromEnvironment,
		},
	}
}
```


# Misc

## userHomeDirMust

```go
func userHomeDirMust() string {
	s, err := os.UserHomeDir()
	must(err)
	return s
}
```

## non-blocking channel send

```go
// if ch if full we will not block, thanks to default case
select {
case ch <- value:
default:
}
```

## openBrowser

```go
// from https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
```

## ProgressEstimator

```go
package util

import (
	"sync"
	"time"
)

// ProgressEstimatorData contains fields readable after Next()
type ProgressEstimatorData struct {
	Total             int
	Curr              int
	Left              int
	PercDone          float64 // 0...1
	Skipped           int
	TimeSoFar         time.Duration
	EstimatedTimeLeft time.Duration
}

// ProgressEstimator is for estimating progress
type ProgressEstimator struct {
	timeStart time.Time
	sync.Mutex
	ProgressEstimatorData
}

// NewProgressEstimator creates a ProgressEstimator
func NewProgressEstimator(total int) *ProgressEstimator {
	d := ProgressEstimatorData{
		Total: total,
	}
	return &ProgressEstimator{
		ProgressEstimatorData: d,
		timeStart:             time.Now(),
	}
}

// Next advances estimator
func (pe *ProgressEstimator) next(isSkipped bool) ProgressEstimatorData {
	pe.Lock()
	defer pe.Unlock()
	if isSkipped {
		pe.Skipped++
	}
	pe.Curr++
	pe.Left = pe.Total - pe.Curr
	pe.TimeSoFar = time.Since(pe.timeStart)

	realTotal := pe.Total - pe.Skipped
	realCurr := pe.Curr - pe.Skipped
	if realCurr == 0 || realTotal == 0 {
		pe.EstimatedTimeLeft = pe.TimeSoFar
	} else {
		pe.PercDone = float64(realCurr) / float64(realTotal) // 0..1 range
		realPerc := float64(realTotal) / float64(realCurr)
		estimatedTotalTime := float64(pe.TimeSoFar) * realPerc
		pe.EstimatedTimeLeft = time.Duration(estimatedTotalTime) - pe.TimeSoFar
	}
	cpy := pe.ProgressEstimatorData
	return cpy
}

// Next advances estimator
func (pe *ProgressEstimator) Next() ProgressEstimatorData {
	return pe.next(false)
}

// Skip advances estimator but allows to mark this file as taking no time,
// to allow better estimates
func (pe *ProgressEstimator) Skip() ProgressEstimatorData {
	return pe.next(true)
}
```

## makeDebounced

Debouncing is rate limiting for calling functions. Call `makeDebounced` with a function and a timeout. It'll return a debouncing function. Calling deboucing function will execute the original function after a timeout. If you call debouncing function before timeout expires, it'll extend the timeout.

```go
// returns a function that will de-bounce f for a given interval
func makeDebounced(d time.Duration, f func()) func() {
	var lastTimer *time.Timer
	return func() {
		if lastTimer != nil {
			lastTimer.Stop()
		}
		lastTimer = time.AfterFunc(d, f)
	}
}
```

Here's how to use:
```go
func testDebounce() {
	n := 1
	f := func() {
		fmt.Printf("de-bounced function called, n: %d\n", n)
	}

	df := makeDebounced(time.Millisecond*200, f)
	df() // should not print 1 becase it's debunced for 200ms
			 // and we call it again after 100ms
	time.Sleep(time.Millisecond * 100)
	n++; df() // should not print 2
	time.Sleep(time.Millisecond * 100)
	n++; df() // should print 3
	time.Sleep(time.Millisecond * 250)

	// it can execute multiple times
	n++; df() // should print 4
	time.Sleep(time.Millisecond * 250)
}
```

## mimeTypeFromFileName

```go
func mimeTypeFromFileName(path string) string {
	var mimeTypes = map[string]string{
		// this is a list from go's mime package
		".css":  "text/css; charset=utf-8",
		".gif":  "image/gif",
		".htm":  "text/html; charset=utf-8",
		".html": "text/html; charset=utf-8",
		".jpg":  "image/jpeg",
		".js":   "application/javascript",
		".wasm": "application/wasm",
		".pdf":  "application/pdf",
		".png":  "image/png",
		".svg":  "image/svg+xml",
		".xml":  "text/xml; charset=utf-8",

		// those are my additions
		".txt":  "text/plain",
		".exe":  "application/octet-stream",
		".json": "application/json",
	}

	ext := strings.ToLower(filepath.Ext(path))
	mt := mimeTypes[ext]
	if mt != "" {
		return mt
	}
	// if not given, default to this
	return "application/octet-stream"
}
```

## formatSize

```go
func formatSize(n int64) string {
	sizes := []int64{1024*1024*1024, 1024*1024, 1024}
	suffixes := []string{"GB", "MB", "kB"}

	for i, size := range sizes {
		if n >= size {
			s := fmt.Sprintf("%.2f", float64(n)/float64(size))
			return strings.TrimSuffix(s, ".00") + " " + suffixes[i]
		}
	}
	return fmt.Sprintf("%d bytes", n)
}
```

## formatDuration

```go
// time.Duration with a better string representation
type FormattedDuration time.Duration

func (d FormattedDuration) String() string {
	return formatDuration(time.Duration(d))
}

// formats duration in a more human friendly way
// than time.Duration.String()
func formatDuration(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "µs") {
		// for µs we don't want fractions
		parts := strings.Split(s, ".")
		if len(parts) > 1 {
			return parts[0] + " µs"
		}
		return strings.ReplaceAll(s, "µs", " µs")
	} else if strings.HasSuffix(s, "ms") {
		// for ms we only want 2 digit fractions
		parts := strings.Split(s, ".")
		//fmt.Printf("fmtDur: '%s' => %#v\n", s, parts)
		if len(parts) > 1 {
			s2 := parts[1]
			if len(s2) > 4 {
				// 2 for "ms" and 2+ for fraction
				res := parts[0] + "." + s2[:2] + " ms"
				//fmt.Printf("fmtDur: s2: '%s', res: '%s'\n", s2, res)
				return res
			}
		}
		return strings.ReplaceAll(s, "ms", " ms")
	}
	return s
}
```

## runCmdMust

```go
func fmtCmdShort(cmd exec.Cmd) string {
	cmd.Path = filepath.Base(cmd.Path)
	return cmd.String()
}

func runCmdMust(cmd *exec.Cmd) string {
	logf("> %s\n", fmtCmdShort(*cmd))
	canCapture := (cmd.Stdout == nil) && (cmd.Stderr == nil)
	if canCapture {
		out, err := cmd.CombinedOutput()
		if err == nil {
			if len(out) > 0 {
				logf("Output:\n%s\n", string(out))
			}
			return string(out)
		}
		logf("cmd '%s' failed with '%s'. Output:\n%s\n", cmd, err, string(out))
		must(err)
		return string(out)
	}
	err := cmd.Run()
	if err == nil {
		return ""
	}
	logf("cmd '%s' failed with '%s'\n", cmd, err)
	must(err)
	return ""
}
```

## runCmdLogged

```go
func runCmdLogged(cmd *exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
```

## cdUpDir

```go
func currDirAbsMust() string {
	dir, err := filepath.Abs(".")
	must(err)
	return dir
}

// we are executed for do/ directory so top dir is parent dir
func cdUpDir(dirName string) {
	startDir := currDirAbsMust()
	dir := startDir
	for {
		// we're already in top directory
		if filepath.Base(dir) == dirName && dirExists(dir) {
			err := os.Chdir(dir)
			must(err)
			return
		}
		parentDir := filepath.Dir(dir)
		panicIf(dir == parentDir, "invalid startDir: '%s', dir: '%s'", startDir, dir)
		dir = parentDir
	}
}
```

## encodeBase64

```go
const base64Chars = "0123456789abcdefghijklmnopqrstuvwxyz"

// encodeBase64 encodes n as base64
func encodeBase64(n int) string {
	var buf [16]byte
	size := 0
	for {
		buf[size] = base64Chars[n%36]
		size++
		if n < 36 {
			break
		}
		n /= 36
	}
	end := size - 1
	for i := 0; i < end; i++ {
		b := buf[i]
		buf[i] = buf[end]
		buf[end] = b
		end--
	}
	return string(buf[:size])
}
```
