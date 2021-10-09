package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kjk/common/httplogger"
	"github.com/kjk/minio"
)

var (
	httpLogger *httplogger.Logger
)

// <dir>/httplog-2021-10-06_01.txt.br
// =>
//apps/cheatsheet/httplog/2021/10-06/2021-10-06_01.txt.br
// return "" if <path> is in unexpected format
func remotePathFromFilePath(app, path string) string {
	name := filepath.Base(path)
	parts := strings.Split(name, "_")
	if len(parts) != 2 {
		return ""
	}
	// parts[1]: 01.txt.br
	hr := strings.Split(parts[1], ".")[0]
	if len(hr) != 2 {
		return ""
	}
	// parts[0]: httplog-2021-10-06
	parts = strings.Split(parts[0], "-")
	if len(parts) != 4 {
		return ""
	}
	year := parts[1]
	month := parts[2]
	day := parts[3]
	name = fmt.Sprintf("%s/%s-%s/%s-%s-%s_%s.txt.br", year, month, day, year, month, day, hr)
	return fmt.Sprintf("apps/%s/httplog/%s", app, name)
}

// upload httplog-2021-10-06_01.txt as
// apps/cheatsheet/httplog/2021/10-06/2021-10-06_01.txt.br
func uploadCompressedHTTPLog(app, path string) {
	timeStart := time.Now()
	mc := newMinioSpacesClient()
	remotePath := remotePathFromFilePath(app, path) + ".br"
	if remotePath == "" {
		logf(ctx(), "uploadCompressedHTTPLog: remotePathFromFilePath() failed for '%s'\n", path)
		return
	}
	_, err := mc.UploadFileBrotliCompressedPublic(remotePath, path)
	if err != nil {
		logerrf(ctx(), "uploadCompressedHTTPLog: minioUploadFilePublic() failed with '%s'\n", err)
		return
	}
	logf(ctx(), "uploadCompressedHTTPLog: uploaded '%s' as '%s' in %s\n", path, remotePath, time.Since(timeStart))
}

func OpenHTTPLog(app string) func() {
	panicIf(app == "")
	dir := "logs"
	must(os.MkdirAll(dir, 0755))

	didRotate := func(path string) {
		canUpload := hasSpacesCreds()
		logf(ctx(), "didRotateHTTPLog: '%s', hasSpacesCreds: %v\n", path, canUpload)
		if !canUpload {
			return
		}
		go uploadCompressedHTTPLog(app, path)
	}
	logger, err := httplogger.New(dir, didRotate)
	must(err)
	// TODO: should I change filerotate so that it opens the file immedaitely?
	logf(context.Background(), "opened http log file\n")
	return func() {
		logger.Close()
	}
}

func LogHTTPReq(r *http.Request, code int, size int64, dur time.Duration) {
	uri := r.URL.Path
	if strings.HasPrefix(uri, "/ping") {
		// our internal health monitoring endpoint is called frequently, don't log
		return
	}
	err := httpLogger.LogReq(r, code, size, dur)
	if err != nil {
		logerrf(ctx(), "logHTTPReq: httpLogSiser.WriteRecord() failed with '%s'\n", err)
	}
}

func hasSpacesCreds() bool {
	return os.Getenv("SPACES_KEY") != "" && os.Getenv("SPACES_SECRET") != ""
}

func newMinioSpacesClient() *minio.Client {
	config := &minio.Config{
		Bucket:   "kjklogs",
		Access:   os.Getenv("SPACES_KEY"),
		Secret:   os.Getenv("SPACES_SECRET"),
		Endpoint: "nyc3.digitaloceanspaces.com",
	}
	mc, err := minio.New(config)
	must(err)
	return mc
}
