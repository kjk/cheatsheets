package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

func logreq(r *http.Request, code int) {
	logf(r.Context(), "%s %s %d\n", r.Method, r.RequestURI, code)
}
