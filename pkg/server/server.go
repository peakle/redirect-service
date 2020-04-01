package server

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/valyala/fasthttp"
)

func StartServer(c *cli.Context) {
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/create":
			handleCreate(ctx)
		case "/redirect":
			handleRedirect(ctx)
		}
	}
	err := fasthttp.ListenAndServe(":80", requestHandler)
	if err != nil {
		fmt.Print(err)
	}
}

func handleRedirect(ctx *fasthttp.RequestCtx) {

}

func handleCreate(ctx *fasthttp.RequestCtx) {

}
