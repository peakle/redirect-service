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

	mysqlHostname, isSet := envMap["MYSQL_HOST"]
	if !isSet {
		fmt.Println("MYSQL_HOSTNAME is not setted in .env file")
		return
	}

	mysqlDatabase, isSet := envMap["MYSQL_DATABASE"]
	if !isSet {
		fmt.Println("MYSQL_DATABASE is not setted in .env file")
		return
	}

	APISecret, isSet := envMap["API_SECRET"]
	if !isSet {
		fmt.Println("API_SECRET is not setted in .env file")
		return
	}

	writeUser, isSet := envMap["MYSQL_WRITE_USER"]
	if !isSet {
		fmt.Println("MYSQL_WRITE_USER is not setted in .env file")
		return
	}

	writePass, isSet := envMap["MYSQL_WRITE_PASS"]
	if !isSet {
		fmt.Println("MYSQL_WRITE_PASS is not setted in .env file")
		return
	}

	readUser, isSet := envMap["MYSQL_READ_USER"]
	if !isSet {
		fmt.Println("MYSQL_READ_USER is not setted in .env file")
		return
	}

	readPass, isSet := envMap["MYSQL_READ_PASS"]
	if !isSet {
		fmt.Println("MYSQL_READ_PASS is not setted in .env file")
		return
	}

	projectDir, isSet := envMap["PROJECT_DIR"]
	if !isSet {
		fmt.Println("PROJECT_DIR is not setted in .env file")
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
	os.Setenv("API_SECRET", "{{.APISecret}}")
	os.Setenv("MYSQL_WRITE_USER", "{{.MysqlWriteUser}}:{{.MysqlWritePass}}")
	os.Setenv("MYSQL_READ_USER", "{{.MysqlReadUser}}:{{.MysqlReadPass}}")
	os.Setenv("HOSTNAME", "{{.Hostname}}")
	os.Setenv("PROJECT_DIR", "{{.ProjectDir}}")
}
`))

	_ = tmpl.Execute(f, struct {
		APISecret      string
		MysqlDatabase  string
		MysqlHostname  string
		MysqlWriteUser string
		MysqlWritePass string
		MysqlReadUser  string
		MysqlReadPass  string
		ProjectDir     string
	}{
		APISecret:      APISecret,
		MysqlDatabase:  mysqlDatabase,
		MysqlHostname:  mysqlHostname,
		MysqlWriteUser: writeUser,
		MysqlWritePass: writePass,
		MysqlReadUser:  readUser,
		MysqlReadPass:  readPass,
		ProjectDir:     projectDir,
	})
}
