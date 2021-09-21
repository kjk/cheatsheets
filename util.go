package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}
func ctx() context.Context {
	return context.Background()
}

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

func logf(ctx context.Context, s string, arg ...interface{}) {
	if len(arg) > 0 {
		s = fmt.Sprintf(s, arg...)
	}
	fmt.Print(s)
}

func isWindows() bool {
	return strings.Contains(runtime.GOOS, "windows")
}

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

func fileExists(path string) bool {
	st, err := os.Lstat(path)
	return err == nil && st.Mode().IsRegular()
}
