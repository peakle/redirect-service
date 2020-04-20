package main

import (
	"fmt"
	"github.com/urfave/cli"
	"gitlab.com/Peakle/redirect-service/pkg/server"
	"os"
)

var (
	Hostname   = "http://localhost:443"
	WriteUser  = "root:root"
	ReadUser   = "root:root"
	commands   = []cli.Command{
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
	app.Name = "rds"
	app.Commands = commands
	app.Metadata = map[string]interface{}{
		"Hostname": Hostname,
		"WriteUser": WriteUser,
		"ReadUser": ReadUser,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error %s", err.Error())
	}
}
