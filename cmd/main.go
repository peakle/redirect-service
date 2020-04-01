package main

import (
	"fmt"
	"github.com/urfave/cli"
	"gitlab.com/redirect-service/pkg/server"
	"os"
)

var (
	Version  = "0"
	CommitID = "0"
	commands = []cli.Command{
		{
			Name:        "server",
			Description: "starts redirect server",
			Action:      server.StartServer,
			Category:    "server",
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "redirect-service"
	app.Commands = commands
	app.Version = fmt.Sprintf("%s - %s", Version, CommitID)

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error %s", err.Error())
	}
}
