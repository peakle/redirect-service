package server

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/valyala/fasthttp"
	"net/url"
)

func StartServer(c *cli.Context) {
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/create":
			handleCreate(ctx)
		default:
			handleRedirect(ctx)
		}
	}

	err := fasthttp.ListenAndServe(":80", requestHandler)
	if err != nil {
		fmt.Print(err)
	}
}

func handleRedirect(ctx *fasthttp.RequestCtx) {
	//var err error

	uri, _ := url.ParseRequestURI(string(ctx.Request.RequestURI()))

	fmt.Print(uri.RawQuery)
}

func handleCreate(ctx *fasthttp.RequestCtx) {

}
