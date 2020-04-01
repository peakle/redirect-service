package main

import (
	"fmt"
	"github.com/urfave/cli"
	"gitlab.com/redirect-service/provider"
	"gitlab.com/redirect-service/server"
	"os"
)

var (
	Version  = "0"
	CommitID = "0"
	commands = []cli.Command{
		{
			Name:        "google-top-charts-parser",
			ShortName:   "parser",
			Description: "parse google play top charts",
			Action:      provider.Create,
			Category:    "parser",
		},
		{
			Name:        "top-charts-server",
			ShortName:   "server",
			Description: "give google play top charts statistic",
			Action:      server.StartServer,
			Category:    "server",
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "atc"
	app.Commands = commands
	app.Version = fmt.Sprintf("%s - %s", Version, CommitID)

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error %s", err.Error())
	}
}
