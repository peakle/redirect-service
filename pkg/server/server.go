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

type EntryDto struct {
	Url    string `json:"external_url"`
	UserId string `json:"user_id"`
	Token  string `json:"token"`
}

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
		",", "",
	)
)

const (
	ErrorMessage = `{"code":1,"text":"please reload page and try again"}`
	JsonResponse = `{"code":"%d","text":"%s"}`
	FrontPage    = "./index.html"
)

func StartServer(c *cli.Context) {
	var err error

	fmt.Println("Start server...")

	db, err = geoip2.Open("GeoIP2.mmdb")
	defer db.Close()

	mRead = provider.InitManager(c.App.Metadata["ReadUser"].(string))
	mWrite = provider.InitManager(c.App.Metadata["WriteUser"].(string))
	defer mRead.Close()
	defer mWrite.Close()

	var hostname, path string
	hostname = c.App.Metadata["Hostname"].(string)

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
	ctx.SendFile(FrontPage)
}

func handleStats(ctx *fasthttp.RequestCtx) {
	var (
		err      error
		entryDto EntryDto
	)

	err = json.Unmarshal(ctx.Request.Body(), &entryDto)
	if err != nil {
		fmt.Printf("error occurred on unmarshal json: %v, provided body: %s \r\n", err, string(ctx.Request.Body()))
		_, _ = fmt.Fprintf(ctx, ErrorMessage)

		return
	}

	if entryDto.UserId == "" {
		entryDto.UserId = "1"
	} else {
		entryDto.UserId = uriReplacer.Replace(entryDto.UserId)
	}

	stats, err := mRead.FindUrlByTokenAndUserId(entryDto.UserId, uriReplacer.Replace(entryDto.Token))
	if err != nil {
		fmt.Printf("error occurred on handle stats: %v, entryDto: %v \r\n", err, entryDto)
		_, _ = fmt.Fprintf(ctx, ErrorMessage)

		return
	}

	var r []byte
	r, err = json.Marshal(stats)
	if err != nil {
		fmt.Printf("error occurred on handle stats format json: %v, entryDto: %v \r\n", err, entryDto)
		_, _ = fmt.Fprintf(ctx, ErrorMessage)

		return
	}

	ctx.Response.Header.SetContentType("application/json")
	_, _ = fmt.Fprintf(ctx, string(r))
}

func handleCreateToken(ctx *fasthttp.RequestCtx) {
	var (
		token    string
		err      error
		entryDto EntryDto
	)

	err = json.Unmarshal(ctx.Request.Body(), &entryDto)
	if err != nil {
		fmt.Printf("error occurred on unmarshal json: %v, provided body: %s \r\n", err, string(ctx.Request.Body()))
		_, _ = fmt.Fprintf(ctx, ErrorMessage)

		return
	}

	if entryDto.UserId == "" {
		entryDto.UserId = "1"
	} else {
		entryDto.UserId = uriReplacer.Replace(entryDto.UserId)
	}

	entryDto.Url = validateAndFixUrl(entryDto.Url)
	token, err = mRead.Create(entryDto.Url)
	if err != nil {
		fmt.Printf("error occurred on create token: %v, entryDto: %v \r\n", err, entryDto)
		_, _ = fmt.Fprintf(ctx, ErrorMessage)

		return
	}

	err = mWrite.InsertToken(entryDto.UserId, entryDto.Url, token)
	if err != nil {
		fmt.Printf("error occurred on insert token: %v, provided data: entryDto: %v, token: %s \r\n", err, entryDto, token)
		_, _ = fmt.Fprintf(ctx, ErrorMessage)

		return
	}

	_, _ = fmt.Fprintf(ctx, fmt.Sprintf(JsonResponse, 0, token))
}

func handleRedirect(ctx *fasthttp.RequestCtx, path string) {
	var err error
	var redirectUri string

	redirectUri, err = mRead.FindUrl(uriReplacer.Replace(path))
	if err != nil {
		fmt.Printf("error occurred on find url: %v \r\n", err)
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
		fmt.Printf("error occurred on parse remote ip: %v \r\n", err)
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

func validateAndFixUrl(u string) string {
	u = uriReplacer.Replace(u)

	if strings.Contains(u, "http://") || strings.Contains(u, "https://") {
		return u
	} else {
		return fmt.Sprintf("http://%s", u)
	}
}
