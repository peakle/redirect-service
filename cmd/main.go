package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/cli"
	"gitlab.com/Peakle/redirect-service/pkg/server"
)

// TODO delete global vars interpritate to env vars and add to gen.go to add on generate
var (
	Hostname   = "http://localhost:443"
	WriteUser  = "root:root_pass"
	ReadUser   = "root:root_pass"
	ProjectDir = "/"
	commands   = []cli.Command{
		{
			Name:        "server",
			Description: "starts redirect server",
			Action:      server.StartServer,
			Category:    "server",
		},
	}
)

//go:generate go run get.go
func main() {
	app := cli.NewApp()
	app.Name = "rds"
	app.Commands = commands
	app.Metadata = map[string]interface{}{
		"Hostname":   Hostname,
		"WriteUser":  WriteUser,
		"ReadUser":   ReadUser,
		"ProjectDir": ProjectDir,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error %s", err.Error())
	}
}
