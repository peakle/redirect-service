package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"strings"

	"github.com/oschwald/geoip2-golang"
	"github.com/urfave/cli"
	"github.com/valyala/fasthttp"
	"gitlab.com/Peakle/redirect-service/pkg/provider"
)

// apiDto for ajax requests from front
type apiDto struct {
	URL    string `json:"external_url"`
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

// VkDto url params from vk sended on init session
type VkDto struct {
	AuthKey  string `schema:"auth_key"`
	APIID    string `schema:"api_id"`
	ViewerID string `schema:"viewer_id"`
}

var (
	dbs       = make(map[string]*geoip2.Reader, 2)
	mRead     *provider.SQLManager
	mWrite    *provider.SQLManager
	apiSecret = os.Getenv("API_SECRET")
	validator = strings.NewReplacer(
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
	errorMessage  = `{"code":1,"text":"Возникла ошибка обновите страницу"}`
	jsonResponse  = `{"code":"%d","text":"%s"}`
	frontPage     = "index.html"
	undefinedCity = "неизвестно"
)

// StartServer - starts redirect server
func StartServer(c *cli.Context) {
	var err error

	fmt.Println("Start server...")

	db, err := geoip2.Open("GeoIP2City.mmdb")
	if err != nil {
		fmt.Fprintln(os.Stderr, "on StartServer: on open GeoIP2City: "+err.Error())
	}
	defer db.Close()
	dbs["city"] = db

	db, err = geoip2.Open("GeoIP2Country.mmdb")
	if err != nil {
		fmt.Fprintln(os.Stderr, "on StartServer: on open GeoIP2Country: "+err.Error())
	}
	defer db.Close()
	dbs["country"] = db

	projectDir := os.Getenv("PROJECT_DIR")

	writeUserPass := os.Getenv("MYSQL_WRITE_USER")
	readUserPass := os.Getenv("MYSQL_READ_USER")

	mRead = provider.InitManager(readUserPass)
	mWrite = provider.InitManager(writeUserPass)
	defer mRead.Close()
	defer mWrite.Close()

	hostname := os.Getenv("HOSTNAME")

	front := projectDir + "/" + frontPage
	favicon := projectDir + "/favicon.ico"
	frontTemplate, _ := template.ParseFiles(front)

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		switch path {
		case "/":
			handleFront(ctx, frontTemplate)
		case "/create":
			handleCreateToken(ctx, hostname)
		case "/stats":
			handleStats(ctx)
		case "/redirect":
			handleRedirect(ctx, path)
		case "/favicon.ico":
			ctx.SendFile(favicon)
		default:
			ctx.Response.SetConnectionClose()
			return
		}
	}

	go fasthttp.ListenAndServe(":8080", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Location", hostname+string(ctx.Path()))
		ctx.Response.Header.Set("Content-Type", "text/html; charset=utf-8")

		ctx.Response.SetStatusCode(302)

		fmt.Println(ctx)
	})

	certFile := projectDir + "/certificate.crt"
	keyFile := projectDir + "/key.pem"

	err = fasthttp.ListenAndServeTLS(":8443", certFile, keyFile, requestHandler)
	if err != nil {
		fmt.Fprintln(os.Stderr, "on startServer: "+err.Error())
	}
}

func handleFront(ctx *fasthttp.RequestCtx, frontTemplate *template.Template) {
	vkDto, err := parseVKDTO(ctx.Request.URI().String())
	if err != nil {
		fmt.Fprintln(os.Stderr, "on handleFront: "+err.Error())

		ctx.Request.SetConnectionClose()

		return
	}

	if !verifyAPIRequest(vkDto) {
		ctx.Request.SetConnectionClose()

		return
	}

	ctx.Response.Header.SetContentType("text/html")
	_ = frontTemplate.Execute(ctx, VkDto{
		AuthKey:  vkDto.AuthKey,
		ViewerID: vkDto.ViewerID,
		APIID:    vkDto.APIID,
	})
}

func handleStats(ctx *fasthttp.RequestCtx) {
	var (
		err      error
		entryDto apiDto
	)

	vkDTO, err := parseVKDTO(ctx.Request.URI().String())
	if err != nil {
		fmt.Fprintln(os.Stderr, "on handleStats: "+err.Error())
		ctx.Request.SetConnectionClose()

		return
	}

	if !verifyAPIRequest(vkDTO) {
		ctx.Request.SetConnectionClose()

		return
	}

	err = json.Unmarshal(ctx.Request.Body(), &entryDto)
	if err != nil {
		err = fmt.Errorf("on handleStats: on unmarshal json: %s, provided body: %s", err.Error(), string(ctx.Request.Body()))
		fmt.Fprintln(os.Stderr, err.Error())

		ctx.Request.SetConnectionClose()
		_, _ = fmt.Fprintf(ctx, errorMessage)

		return
	}

	entryDto.UserID = validator.Replace(entryDto.UserID)

	var (
		uri   *url.URL
		token string
	)

	uri, err = url.Parse(entryDto.Token)
	if err == nil && uri.Path != "" {
		token = validator.Replace(strings.TrimLeft(uri.Path, "/"))
	} else {
		entryDto.Token = validateRedirectURL(entryDto.Token)
	}

	stats, err := mRead.FindURLByTokenAndUserID(entryDto.UserID, token)
	if err != nil {
		err = fmt.Errorf("on handleStats: %s, entryDto: %v", err.Error(), entryDto)
		fmt.Fprintln(os.Stderr, err.Error())

		ctx.Request.SetConnectionClose()
		_, _ = fmt.Fprintf(ctx, errorMessage)

		return
	}

	var r []byte
	r, err = json.Marshal(stats)
	if err != nil {
		err = fmt.Errorf("on handleStats: on format json: %s, entryDto: %v", err.Error(), entryDto)
		fmt.Fprintln(os.Stderr, err.Error())

		ctx.Request.SetConnectionClose()
		_, _ = fmt.Fprintf(ctx, errorMessage)

		return
	}

	ctx.Response.Header.SetContentType("application/json")
	_, _ = fmt.Fprintf(ctx, string(r))
}

func handleCreateToken(ctx *fasthttp.RequestCtx, hostname string) {
	var (
		token    string
		err      error
		entryDto apiDto
	)

	vkDTO, err := parseVKDTO(ctx.Request.URI().String())
	if err != nil {
		fmt.Fprintln(os.Stderr, "on handleCreateToken: "+err.Error())
		ctx.Request.SetConnectionClose()

		return
	}

	if !verifyAPIRequest(vkDTO) {
		ctx.Request.SetConnectionClose()

		return
	}

	err = json.Unmarshal(ctx.Request.Body(), &entryDto)
	if err != nil {
		err = fmt.Errorf("on handleCreateToken: on unmarshal json: %s, provided body: %s", err.Error(), string(ctx.Request.Body()))
		fmt.Fprintln(os.Stderr, err.Error())

		ctx.Request.SetConnectionClose()
		_, _ = fmt.Fprintf(ctx, errorMessage)

		return
	}

	entryDto.UserID = validator.Replace(entryDto.UserID)

	entryDto.URL = validateRedirectURL(entryDto.URL)
	token, err = mRead.Create(entryDto.URL)
	if err != nil {
		err = fmt.Errorf("on handleCreateToken: on Create: %s, entryDto: %v", err.Error(), entryDto)
		fmt.Fprintln(os.Stderr, err.Error())

		ctx.Request.SetConnectionClose()
		_, _ = fmt.Fprintf(ctx, errorMessage)

		return
	}

	err = mWrite.InsertToken(entryDto.UserID, entryDto.URL, token)
	if err != nil {
		err = fmt.Errorf("on handleCreateToken: on insert token: %s, provided data: entryDto: %v, token: %s", err.Error(), entryDto, token)
		fmt.Fprintln(os.Stderr, err.Error())

		ctx.Request.SetConnectionClose()
		_, _ = fmt.Fprintf(ctx, errorMessage)

		return
	}

	redirectURI := fmt.Sprintf(`%s/redirect/%s`, hostname, token)

	_, _ = fmt.Fprintf(ctx, fmt.Sprintf(jsonResponse, 0, redirectURI))
}

func handleRedirect(ctx *fasthttp.RequestCtx, path string) {
	var err error
	var token, redirectURI string

	token = strings.TrimLeft(validator.Replace(path), "/")
	redirectURI, err = mRead.FindURLByToken(token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "on handleRedirect: on find url: "+err.Error())

		redirectURI = fmt.Sprintf("https://yandex.ru/search/?text=%s", strings.TrimLeft(path, "/"))
	}

	ctx.Response.Header.Set("Location", redirectURI)
	ctx.Response.Header.Set("Content-Type", "text/html; charset=utf-8")

	ctx.Response.SetStatusCode(302)

	fmt.Println(ctx)

	ctx.SetConnectionClose()

	recordReq(ctx.RemoteIP(), ctx.UserAgent(), token)
}
