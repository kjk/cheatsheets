package main

import (
	"path/filepath"
	"testing"
)

func TestRemotePathFromFilePath(t *testing.T) {
	tests := []string{
		filepath.Join("foo", "httplog-2021-10-06_01.txt.br"),
		"apps/cheatsheet/httplog/2021/10-06/2021-10-06_01.txt.br",
	}
	n := len(tests)
	for i := 0; i < n; i += 2 {
		s := tests[i]
		exp := tests[i+1]
		got := remotePathFromFilePath(s)
		if exp != got {
			t.Errorf("s:'%s', got: '%s', exp: '%s'", s, got, exp)
		}
	}
}
