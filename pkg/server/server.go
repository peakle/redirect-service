package server

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/schema"
	"github.com/oschwald/geoip2-golang"
	"github.com/urfave/cli"
	"github.com/valyala/fasthttp"
	"gitlab.com/Peakle/redirect-service/pkg/provider"
)

// EntryDto for ajax requests from front
type EntryDto struct {
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
	db        *geoip2.Reader
	mRead     *provider.SQLManager
	mWrite    *provider.SQLManager
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
	errorMessage  = `{"code":1,"text":"please reload page and try again"}`
	jsonResponse  = `{"code":"%d","text":"%s"}`
	frontPage     = "index.html"
	undefinedCity = "неизвестно"
)

// StartServer - starts redirect server
func StartServer(c *cli.Context) {
	var err error

	fmt.Println("Start server...")

	db, err = geoip2.Open("GeoIP2.mmdb")
	defer db.Close()

	projectDir := c.App.Metadata["ProjectDir"].(string)

	// TODO replace manager init from boot
	mRead = provider.InitManager(c.App.Metadata["ReadUser"].(string))
	mWrite = provider.InitManager(c.App.Metadata["WriteUser"].(string))
	defer mRead.Close()
	defer mWrite.Close()

	var hostname, path string
	hostname = c.App.Metadata["Hostname"].(string)

	front := projectDir + "/" + frontPage
	favicon := projectDir + "/favicon.ico"
	frontTemplate, _ := template.ParseFiles(front)

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path = string(ctx.Path())
		switch path {
		case "/":
			handleFront(ctx, frontTemplate)
		case "/create":
			handleCreateToken(ctx, hostname)
		case "/stats":
			handleStats(ctx)
		case "/redirect":
			handleRedirect(ctx, path)
		default:
			if strings.Contains(path, "favicon.ico") {
				ctx.SendFile(favicon)
			}
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
		fmt.Println(err)
	}
}

func handleFront(ctx *fasthttp.RequestCtx, frontTemplate *template.Template) {
	uri, err := url.Parse(ctx.Request.URI().String())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var entryDto VkDto

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	err = decoder.Decode(&entryDto, uri.Query())
	if err != nil {
		fmt.Println("error occurred on decode url params" + err.Error())
		return
	}

	if !verifyRequest(entryDto) {
		return
	}

	ctx.Response.Header.SetContentType("text/html")
	_ = frontTemplate.Execute(ctx, VkDto{
		AuthKey:  entryDto.AuthKey,
		ViewerID: entryDto.ViewerID,
		APIID:    entryDto.APIID,
	})
}

func handleStats(ctx *fasthttp.RequestCtx) {
	var (
		err      error
		entryDto EntryDto
	)

	err = json.Unmarshal(ctx.Request.Body(), &entryDto)
	if err != nil {
		fmt.Printf("error occurred on unmarshal json: %v, provided body: %s \r\n", err, string(ctx.Request.Body()))
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
		fmt.Printf("error occurred on handle stats: %v, entryDto: %v \r\n", err, entryDto)
		_, _ = fmt.Fprintf(ctx, errorMessage)

		return
	}

	var r []byte
	r, err = json.Marshal(stats)
	if err != nil {
		fmt.Printf("error occurred on handle stats format json: %v, entryDto: %v \r\n", err, entryDto)
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
		entryDto EntryDto
	)

	if !verifyAPIRequest(ctx.Request.URI().String()) {
		return
	}

	err = json.Unmarshal(ctx.Request.Body(), &entryDto)
	if err != nil {
		fmt.Printf("error occurred on unmarshal json: %v, provided body: %s \r\n", err, string(ctx.Request.Body()))
		_, _ = fmt.Fprintf(ctx, errorMessage)

		return
	}

	entryDto.UserID = validator.Replace(entryDto.UserID)

	entryDto.URL = validateRedirectURL(entryDto.URL)
	token, err = mRead.Create(entryDto.URL)
	if err != nil {
		fmt.Printf("error occurred on create token: %v, entryDto: %v \r\n", err, entryDto)
		_, _ = fmt.Fprintf(ctx, errorMessage)

		return
	}

	err = mWrite.InsertToken(entryDto.UserID, entryDto.URL, token)
	if err != nil {
		fmt.Printf("error occurred on insert token: %v, provided data: entryDto: %v, token: %s \r\n", err, entryDto, token)
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
		fmt.Printf("error occurred on find url: %v \r\n", err)
		redirectURI = fmt.Sprintf("https://yandex.ru/search/?text=%s", strings.TrimLeft(path, "/"))
	}

	ctx.Response.Header.Set("Location", redirectURI)
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
		mWrite.RecordStats(ctx.RemoteIP().String(), string(ctx.UserAgent()), city.Country.Names["ru"]+" "+v, token)
	} else if v, isSet = city.Country.Names["ru"]; isSet {
		mWrite.RecordStats(ctx.RemoteIP().String(), string(ctx.UserAgent()), v, token)
	} else {
		mWrite.RecordStats(ctx.RemoteIP().String(), string(ctx.UserAgent()), undefinedCity, token)
	}
}

func validateRedirectURL(uri string) string {
	uri = validator.Replace(uri)

	if strings.Contains(uri, "http://") || strings.Contains(uri, "https://") {
		return uri
	}

	return fmt.Sprintf("http://%s", uri)
}

func verifyRequest(reqDto VkDto) bool {
	apiSecret := os.Getenv("API_SECRET")
	serverAuthKey := fmt.Sprintf("%x", md5.Sum([]byte(reqDto.APIID+"_"+reqDto.ViewerID+"_"+apiSecret)))

	return reqDto.AuthKey == serverAuthKey
}

func verifyAPIRequest(uriPath string) bool {
	uri, err := url.Parse(uriPath)
	if err != nil {
		fmt.Println("error occurred on verify api reques: " + err.Error())
		return false
	}

	var entryDto VkDto

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	err = decoder.Decode(&entryDto, uri.Query())
	if err != nil {
		fmt.Println("error occurred on decode url params " + err.Error())
		return false
	}

	if !verifyRequest(entryDto) {
		return false
	}

	return true
}
