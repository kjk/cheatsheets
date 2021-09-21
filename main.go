package main

import (
	"flag"
)

func doRun() {
	logf(ctx(), "doRun:\n")
}

func runServer() {

	waitFn := StartServer(nil)
	waitFn()
}

func main() {
	var (
		flgRunServer bool
	)
	{
		flag.BoolVar(&flgRunServer, "run-server", false, "run me")
		flag.Parse()
	}
	if flgRunServer {
		runServer()
		return
	}
	flag.Usage()
}
