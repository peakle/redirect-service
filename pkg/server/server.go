package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oschwald/geoip2-golang"
	"github.com/urfave/cli"
	"github.com/valyala/fasthttp"
	"gitlab.com/Peakle/redirect-service/pkg/provider"
)

var (
	db          *geoip2.Reader
	mRead       *provider.SQLManager
	mWrite      *provider.SQLManager
	uriReplacer = strings.NewReplacer(
		"*", "",
		"<", "",
		">", "",
		"(", "",
		")", "",
		"'", "",
		`"`, "",
		`;`, "",
		" ", "",
	)
)

const ErrorMessage = "please reload page and try again"

func StartServer(c *cli.Context) {
	var err error

	fmt.Println("Start server...")

	db, err = geoip2.Open("GeoIP2.mmdb")
	defer db.Close()

	mRead = provider.InitManager()
	mWrite = provider.InitManager()
	defer mRead.Close()
	defer mWrite.Close()

	var hostname, path string
	hostname = c.App.Metadata["hostname"].(string)

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path = string(ctx.Path())
		switch path {
		case "/":
			handleFront(ctx)
		case "/creation":
			handleCreateToken(ctx)
		case "/stats":
			handleStats(ctx)
		default:
			if strings.Contains(path, "favicon.ico") {
				return
			}

			handleRedirect(ctx, path)
		}
	}

	go fasthttp.ListenAndServe(":80", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Location", hostname+string(ctx.Path()))
		ctx.Response.Header.Set("Content-Type", "text/html; charset=utf-8")

		ctx.Response.SetStatusCode(302)

		fmt.Println(ctx)
	})

	err = fasthttp.ListenAndServe(":443", requestHandler)
	if err != nil {
		fmt.Println(err)
	}
}

func handleFront(ctx *fasthttp.RequestCtx) {
	ctx.SendFile("./index.html")
}

func handleStats(ctx *fasthttp.RequestCtx) {
	userId := string(ctx.Request.PostArgs().Peek("user_id"))
	token := string(ctx.Request.PostArgs().Peek("token"))

	if userId == "" {
		userId = "1"
	}

	stats, err := mRead.FindUrlByTokenAndUserId(userId, token)
	if err != nil {
		fmt.Printf("error occurred on handle stats: %v, token: %s", err, token)
		_, _ = fmt.Fprintf(ctx, "[]")

		return
	}

	var r []byte
	r, err = json.Marshal(stats)
	if err != nil {
		fmt.Printf("error occurred on handle stats format json: %v, token: %s", err, token)
		_, _ = fmt.Fprintf(ctx, "[]")

		return
	}

	ctx.Response.Header.SetContentType("application/json")
	_, _ = fmt.Fprintf(ctx, string(r))
}

func handleCreateToken(ctx *fasthttp.RequestCtx) {
	var userId, uri, token string
	var err error

	uri = string(ctx.Request.PostArgs().Peek("external_url"))
	userId = string(ctx.Request.PostArgs().Peek("user_id"))

	if userId == "" {
		userId = "1"
	}

	uri = uriReplacer.Replace(uri)
	token, err = mRead.Create(uri)
	if err != nil {
		fmt.Printf("error occurred on create token: %v, provided url: %s", err, uri)
		_, _ = fmt.Fprintf(ctx, ErrorMessage)

		return
	}

	err = mWrite.InsertToken(userId, uri, token)
	if err != nil {
		fmt.Printf("error occurred on insert token: %v, provided data: userId: %s, uri: %s, token: %s", err, userId, uri, token)
		_, _ = fmt.Fprintf(ctx, ErrorMessage)

		return
	}

	_, _ = fmt.Fprintf(ctx, uri)
}

func handleRedirect(ctx *fasthttp.RequestCtx, path string) {
	var err error
	var redirectUri string

	redirectUri, err = mRead.FindUrl(uriReplacer.Replace(path))
	if err != nil {
		fmt.Printf("error occurred on find url: %v", err)
		redirectUri = fmt.Sprintf("https://yandex.ru/search/?text=%s", strings.TrimLeft(path, "/"))
	}

	ctx.Response.Header.Set("Location", redirectUri)
	ctx.Response.Header.Set("Content-Type", "text/html; charset=utf-8")

	ctx.Response.SetStatusCode(302)

	fmt.Println(ctx)

	ctx.SetConnectionClose()

	var city *geoip2.City

	city, err = db.City(ctx.RemoteIP())
	if err != nil {
		fmt.Printf("error occurred on parse remote ip: %v", err)
		return
	}

	var v string
	var isSet bool
	if v, isSet = city.City.Names["ru"]; isSet {
		mWrite.RecordStats(ctx.RemoteIP().String(), string(ctx.UserAgent()), city.Country.Names["ru"]+" "+v)
	} else if v, isSet = city.Country.Names["ru"]; isSet {
		mWrite.RecordStats(ctx.RemoteIP().String(), string(ctx.UserAgent()), v)
	}
}
