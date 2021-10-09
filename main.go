package main

import (
	"flag"
	"os"

	"github.com/kjk/common/httputil"
)

func deployToRender() {
	deployURL := os.Getenv("CHEATSHEETS_DEPLOY_HOOK")
	panicIf(deployURL == "", "need env variable CHEATSHEETS_DEPLOY_HOOK")
	d, err := httputil.Get(deployURL)
	must(err)
	logf(ctx(), "deployed to render.com:\n%s\n", string(d))
}

func main() {
	var (
		flgRunServer     bool
		flgRunServerProd bool
		flgGen           bool
		flgDeploy        bool
	)
	{
		flag.BoolVar(&flgRunServer, "run", false, "run dev server")
		flag.BoolVar(&flgRunServerProd, "run-prod", false, "run prod server serving www_generated")
		flag.BoolVar(&flgGen, "gen", false, "generate static files in www_generated dir")
		flag.BoolVar(&flgDeploy, "deploy", false, "deploy to render.com")
		flag.Parse()
	}

	if false {
		compareCompr()
		return
	}

	if flgRunServer {
		runServerDynamic()
		return
	}

	if flgRunServerProd {
		runServerProd()
		return
	}

	if flgGen {
		generateStatic()
		return
	}

	if flgDeploy {
		deployToRender()
		return
	}

	flag.Usage()
}
