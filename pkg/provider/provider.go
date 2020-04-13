package provider

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"time"
)

const CreationError = "creation error"

func (m *SQLManager) Create(uri string) (string, error) {
	var err error
	_, err = url.ParseRequestURI(uri)
	if err != nil {
		return "", err
	}

	InitManager()

	t := make([]byte, 7)
	var r string

	for i := 0; ; i++ {
		rand.Read(t)
		r = fmt.Sprintf("%x", t)
		if !m.TokenExist(r) {
			break
		}

		time.Sleep(time.Second)
		if i == 10 {
			return "", errors.New(CreationError)
		}
	}

	return r, nil
}
