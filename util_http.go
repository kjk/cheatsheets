package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kjk/common/server"
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
	httpSrv := makeHTTPServer(srv)
	return StartHTTPServer(httpSrv)
}
