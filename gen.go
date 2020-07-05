// +build ignore

package main

import (
	"fmt"
	"os"
	"text/template"

	"github.com/joho/godotenv"
)

func main() {
	envFile, err := os.Open("../.env")
	if err != nil {
		fmt.Println("error on read .env file")
		return
	}

	envMap, _ := godotenv.Parse(envFile)

	var isSet bool
	var MysqlHostname, MysqlDatabase, APISecret string

	MysqlHostname, isSet = envMap["MYSQL_HOST"]
	if !isSet {
		fmt.Println("MYSQL_HOSTNAME is not setted in .env file")
		return
	}

	MysqlDatabase, isSet = envMap["MYSQL_DATABASE"]
	if !isSet {
		fmt.Println("MYSQL_DATABASE is not setted in .env file")
		return
	}

	APISecret, isSet = envMap["API_SECRET"]
	if !isSet {
		fmt.Println("API_SECRET is not setted in .env file")
		return
	}

	f, err := os.Create("env.go")
	if err != nil {
		fmt.Println("error occurred on create env.go file")
		return
	}

	tmpl := template.Must(template.New("").Parse(`
package main

import "os"

func init() {
	os.Setenv("MYSQL_HOSTNAME", "{{.MysqlHostname}}")
	os.Setenv("MYSQL_DATABASE", "{{.MysqlDatabase}}")
	os.Setenv("APISecret", "{{.APISecret}}")
}
`))

	_ = tmpl.Execute(f, struct {
		APISecret     string
		MysqlDatabase string
		MysqlHostname string
	}{
		APISecret:     APISecret,
		MysqlDatabase: MysqlDatabase,
		MysqlHostname: MysqlHostname,
	})
}
