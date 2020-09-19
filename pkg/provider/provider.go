package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

// Create redirect token for provided url and insetr it to db
func (m *SQLManager) Create(uri string) (string, error) {
	const CreationError = "creation error"

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
			return t, nil
		}

		time.Sleep(time.Second)
		if i == 10 {
			return "", errors.New(CreationError)
		}
	}
}

func isEmpty(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}
