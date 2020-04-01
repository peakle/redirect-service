package provider

import (
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const ApiMessage = "https://api.vk.com/method/messages.send?"
const ApiVersion = "5.101"

func Create(uri string) (string, error) {
	var err error
	_, err = url.ParseRequestURI(uri)
	if err != nil {
		return "", err
	}

	t := make([]byte, 7)
	var r string

	for i := 0; ; i++ {
		rand.Read(t)
		r = fmt.Sprintf("%x", t)
		if !tokenExist(r) {
			break
		}

		time.Sleep(time.Second)
		if i == 10 {
			return "", errors.New("ошибка создания")
		}
	}

	return r, nil
}

func FindUrl(url string) (string, error) {
	if url == "" {
		return "", nil
	}

	return "", nil
}

func RecordStats(ctx *fasthttp.RequestCtx) {
	time.Sleep(time.Second * 100)
	fmt.Println(123)
}

func SendMessageVK(userId, message string) {
	rand.Seed(time.Now().Unix())
	randomId := strconv.Itoa(rand.Intn(10000))

	urlArgs := url.Values{}
	urlArgs.Add("user_id", userId)
	urlArgs.Add("random_id", randomId)
	urlArgs.Add("access_token", os.Getenv("VK_TOKEN"))
	urlArgs.Add("message", message)
	urlArgs.Add("v", ApiVersion)
	urlInfo := urlArgs.Encode()

	_, err := http.Get(ApiMessage + urlInfo)
	if err != nil {
		fmt.Printf("error in send message: %v", err)
	}
}

func tokenExist(token string) bool {
	return false
}
