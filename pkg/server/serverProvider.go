package server

import (
	"crypto/md5"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/schema"
	"github.com/oschwald/geoip2-golang"
)

func validateRedirectURL(uri string) string {
	uri = validator.Replace(uri)

	if strings.Contains(uri, "http://") || strings.Contains(uri, "https://") {
		return uri
	}

	return fmt.Sprintf("http://%s", uri)
}

func recordReq(ip net.IP, useragent []byte, token string) {
	var db *geoip2.Reader
	var v string
	var ok bool

	db = dbs["city"]
	city, err := db.City(ip)
	if err != nil {
		fmt.Fprintln(os.Stderr, "on recordReq: on parse ip by City: "+err.Error())

		return
	}

	if v, ok = city.City.Names["ru"]; ok {
		mWrite.RecordStats(ip.String(), string(useragent), city.Country.Names["ru"]+" "+v, token)

		return
	} else if v, ok = city.Country.Names["ru"]; ok {
		mWrite.RecordStats(ip.String(), string(useragent), v, token)

		return
	}

	db = dbs["country"]
	country, err := db.Country(ip)
	if err != nil {
		fmt.Fprintln(os.Stderr, "on recordReq: on parse ip by Country: "+err.Error())

		return
	}

	if v, ok = country.Country.Names["ru"]; ok {
		mWrite.RecordStats(ip.String(), string(useragent), v, token)

		return
	}

	mWrite.RecordStats(ip.String(), string(useragent), undefinedCity, token)
}

func verifyRequest(reqDto *VkDto) bool {
	serverAuthKey := fmt.Sprintf("%x", md5.Sum([]byte(reqDto.APIID+"_"+reqDto.ViewerID+"_"+apiSecret)))

	return reqDto.AuthKey == serverAuthKey
}

func verifyAPIRequest(vkDto *VkDto) bool {
	if vkDto.ViewerID == "" {
		vkDto.ViewerID = "0"
		return true
	}

	return verifyRequest(vkDto)
}

func parseVKDTO(uriPath string) (*VkDto, error) {
	uri, err := url.Parse(uriPath)
	if err != nil {
		err = fmt.Errorf("on verifyApiRequest: on url.Parse: " + err.Error())
		return nil, err
	}

	var entryDto *VkDto

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	err = decoder.Decode(entryDto, uri.Query())
	if err != nil {
		err = fmt.Errorf("on verifyApiRequest: on decode url params: " + err.Error())
		return nil, err
	}

	return entryDto, nil
}
