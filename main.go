package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "mast",
		Usage: "A simple TUI for sending casts",
		Commands: []*cli.Command{
			{
				Name:    "auth",
				Aliases: []string{"a"},
				Usage:   "Authorize the CLI with your Signer Private Key and FID",
				Action: func(ctx *cli.Context) error {
					return SetFidAndPrivateKey()
				},
			},
			{
				Name:      "new",
				Aliases:   []string{"n"},
				Usage:     "Send a new Cast",
				ArgsUsage: "[message]",
				Action: func(ctx *cli.Context) error {
					castData, err := ComposeCast()
					if err != nil {
						return err
					}
					return SendCast(castData)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
