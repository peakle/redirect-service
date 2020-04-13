package main

import (
	"fmt"
	"github.com/urfave/cli"
	"gitlab.com/Peakle/redirect-service/pkg/server"
	"os"
)

var (
	Hostname = "http://localhost:443"
	commands = []cli.Command{
		{
			Name:        "server",
			Description: "starts redirect server",
			Action:      server.StartServer,
			Category:    "server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "hostname",
					Value:    Hostname,
				},
			},
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "redi"
	app.Commands = commands

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error %s", err.Error())
	}
}
