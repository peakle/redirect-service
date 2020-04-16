package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"time"
)

const CreationError = "creation error"

func (m *SQLManager) Create(uri string) (string, error) {
	var err error
	_, err = url.ParseRequestURI(uri)
	if err != nil {
		return "", err
	}

	r := make([]byte, 7)
	var t string

	for i := 0; ; i++ {
		rand.Read(r)
		t = fmt.Sprintf("%x", r)
		if !m.TokenExist(t) {
			break
		}

		time.Sleep(time.Second)
		if i == 10 {
			return "", errors.New(CreationError)
		}
	}

	return t, nil
}

func isEmpty(str string) bool {
	regex := regexp.MustCompile(`(?m)\s+`)

	return regex.ReplaceAllString(str, "") == ""
}
