package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
To enable logging to logdna, set LOGDNA_API_KEY env variable

To enable logging to logtail, set LOGTAIL_API_KEY env variable
It should be the Authorization header: "Bearer XXX"
*/

const (
	logdnaApp  = "cheatsheets"
	logdnaHost = "main"
)

func printLoggingStats() {
	{
		apiKey := os.Getenv("LOGDNA_API_KEY")
		if len(apiKey) < 32 {
			fmt.Printf("Not logging to logdna because LOGDNA_API_KEY env var not set or invalid\n")
		} else {
			fmt.Printf("Logging to logdna because LOGDNA_API_KEY env var set\n")
		}
	}
	{
		apiKey := os.Getenv("LOGTAIL_API_KEY")
		if !strings.HasPrefix(apiKey, "Bearer ") {
			fmt.Printf("Not logging to logdna because LOGTAIL_API_KEY env var not set or invalid\n")
		} else {
			fmt.Printf("Logging to logdna because LOGTAIL_API_KEY env var set\n")
		}
	}
}

// https://docs.logdna.com/reference#logsingest
func logdna(s string, now time.Time, isError bool) {
	// sending logs async because we don't care if fails
	apiKey := os.Getenv("LOGDNA_API_KEY")
	if len(apiKey) < 32 {
		return
	}
	go func() {
		line := map[string]interface{}{
			"line":      s,
			"app":       "",
			"timestamp": now.UnixNano() / 1000000,
		}
		if isError {
			line["level"] = "ERROR"
		}
		lines := []map[string]interface{}{line}
		v := map[string]interface{}{
			"lines": lines,
		}
		d, _ := json.Marshal(v)
		body := strings.NewReader(string(d))
		c := http.DefaultClient
		uri := fmt.Sprintf("https://logs.logdna.com/logs/ingest?hostname=%s&apikey=%s", logdnaHost, apiKey)
		req, err := http.NewRequest("POST", uri, body)
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		rsp, err := c.Do(req)
		if err != nil {
			return
		}
		defer rsp.Body.Close()
	}()
}

// https://docs.logtail.com/integrations/rest-api
func logtail(s string, isError bool) {
	// sending logs async because we don't care if fails
	apiKey := os.Getenv("LOGTAIL_API_KEY")
	if !strings.HasPrefix(apiKey, "Bearer ") {
		return
	}

	go func() {
		v := map[string]interface{}{}
		if isError {
			v["error"] = s
		} else {
			v["message"] = s
		}
		d, _ := json.Marshal(v)
		body := strings.NewReader(string(d))
		c := http.DefaultClient
		uri := "https://in.logtail.com/"
		req, err := http.NewRequest("POST", uri, body)
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", apiKey)
		rsp, err := c.Do(req)
		if err != nil {
			return
		}
		defer rsp.Body.Close()
	}()
}

func logf(ctx context.Context, s string, args ...interface{}) {
	now := time.Now()
	if len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}

	fmt.Print(s)
	logtail(s, false)
	logdna(s, now, false)
}

// TODO: write to a file logs/${day}.txt
func logvf(ctx context.Context, s string, args ...interface{}) {
	if len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}
	fmt.Print(s)
}

func logerrf(ctx context.Context, format string, args ...interface{}) {
	now := time.Now()
	s := format
	if len(args) > 0 {
		s = fmt.Sprintf(format, args...)
	}
	fmt.Printf("Error: %s", s)
	logtail(s, true)
	logdna(s, now, true)
}

var (
	hdrsToNotLog = []string{
		"Connection",
		"Sec-Ch-Ua-Mobile",
		"Sec-Fetch-Dest",
		"Sec-Ch-Ua-Platform",
		"Dnt",
		"Upgrade-Insecure-Requests",
		"Sec-Fetch-Site",
		"Sec-Fetch-Mode",
		"Sec-Fetch-User",
		"If-Modified-Since",
		"Accept-Language",
		"Cf-Ray",
		"X-Request-Start",
		"Cdn-Loop",
		"X-Forwarded-Proto",
	}
	hdrsToNotLogMap map[string]bool
)

func shouldLogHeader(s string) bool {
	if hdrsToNotLogMap == nil {
		hdrsToNotLogMap = map[string]bool{}
		for _, h := range hdrsToNotLog {
			h = strings.ToLower(h)
			hdrsToNotLogMap[h] = true
		}
	}
	s = strings.ToLower(s)
	return !hdrsToNotLogMap[s]
}

func logHTTPReq(r *http.Request, code int, size int64, dur time.Duration) {
	logf(ctx(), "%s %s %d in %s\n", r.Method, r.RequestURI, code, dur)
	httpLogMu.Lock()
	defer httpLogMu.Unlock()
	rec := &httpLogRec
	rec.Reset()
	rec.Write("req", fmt.Sprintf("%s %s %d", r.Method, r.RequestURI, code))
	recWriteNonEmpty(rec, "host", r.Host)
	rec.Write("ipaddr", requestGetRemoteAddress(r))
	rec.Write("size", strconv.FormatInt(size, 10))
	durMs := int64(dur / time.Millisecond)
	rec.Write("duration", strconv.FormatInt(durMs, 10))

	for k, v := range r.Header {
		if !shouldLogHeader(k) {
			continue
		}
		logf(ctx(), "%s: %s\n", k, v[0])
		if len(v) > 0 && len(v[0]) > 0 {
			rec.Write(k, v[0])
		}
	}

	// TODO: write to log file
}
