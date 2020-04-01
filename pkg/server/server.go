package server

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"github.com/valyala/fasthttp"
	"gitlab.com/Peakle/redirect-service/pkg/provider"
	"os"
	"strconv"
)

const (
	VkGroupId    = 177387678
	Confirmation = "confirmation"
	MessageNew   = "message_new"
)

var secret string
var vkConfirmationToken string

type vkMessage struct {
	Secret  string          `json:"secret"`
	GroupId int             `json:"group_id"`
	Type    string          `json:"type"`
	Object  vkMessageObject `json:"object"`
}

type vkMessageObject struct {
	FromId int    `json:"from_id"`
	Text   string `json:"text"`
}

func StartServer(c *cli.Context) {
	fmt.Println("Start server...")
	secret = os.Getenv("VK_API_SECRET")
	vkConfirmationToken = os.Getenv("VK_CONFIRMATION_TOKEN")

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/vk":
			handleVK(ctx)
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
	var err error
	ctx.Response.Header.Set("Content-Type", "text/html; charset=utf-8")

	redirectUri, err := provider.FindUrl(string(ctx.Request.RequestURI()))
	if err != nil {
		_, _ = fmt.Fprint(ctx, "")
	}

	ctx.Response.Header.Set("Location", redirectUri)
	ctx.Response.SetStatusCode(301)
	ctx.SetConnectionClose()

	provider.RecordStats(ctx)
}

func handleVK(ctx *fasthttp.RequestCtx) {
	var err error

	message := vkMessage{}
	err = json.Unmarshal(ctx.Request.Body(), &message)
	if err != nil {
		fmt.Printf("error occurred on parse vk message: %v", err)
		_, _ = fmt.Fprint(ctx, "")
		return
	}

	if message.Secret != secret || len(message.Secret) == 0 {
		_, _ = fmt.Fprint(ctx, "")
		return
	}

	switch message.Type {
	case Confirmation:
		if message.GroupId != VkGroupId {
			_, _ = fmt.Fprint(ctx, "")
			return
		}

		_, _ = fmt.Fprint(ctx, vkConfirmationToken)
	case MessageNew:
		_, _ = fmt.Fprint(ctx, "ok")

		var r string
		r, err = provider.Create(message.Object.Text)
		if err != nil {
			r = "невалидный url"
		}

		provider.SendMessageVK(strconv.Itoa(message.Object.FromId), r)
	default:
		_, _ = fmt.Fprint(ctx, "")
	}
	ctx.SetConnectionClose()
}
