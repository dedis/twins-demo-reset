package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dedis/twins-demo-reset/server"
	"github.com/urfave/cli/v2"
)

// managed by govvv
var GitSummary string

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s\n", GitSummary)
	}

	app := &cli.App{
		Name:  "reset",
		Usage: "an HTTP server that resets the TWIN agents for the demo",
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "start the organizer server",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Usage:   "port to listen websocket connections on",
						Value:   9999,
					},
				},
				Action: server.Serve,
			},
		},
		Version: GitSummary,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
