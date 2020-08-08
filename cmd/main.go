package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/cli"
	"gitlab.com/Peakle/redirect-service/pkg/server"
)

var (
	commands = []cli.Command{
		{
			Name:        "server",
			Description: "starts redirect server",
			Action:      server.StartServer,
			Category:    "server",
		},
	}
)

//go:generate go run ../gen.go
func main() {
	app := cli.NewApp()
	app.Name = "rds"
	app.Commands = commands

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error %s", err.Error())
	}
}
