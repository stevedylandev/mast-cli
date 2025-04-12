package main

import (
	"log"
	"os"

	auth "mast/auth"
	compose "mast/compose"
	hub "mast/hub"
	login "mast/login"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "mast",
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "steve",
				Email: "hello@stevedylan.dev",
			},
		},
		UsageText: "A simple TUI for sending casts",
		Usage: `
								 +
                +++
               +++++
               +++++
              +++++++
              ++++++++
             +++++++++=
            +++++++++++
            ++++++++++++
           +++++++++++++=
          +++++++++++++++
          +++++++++++++++
         ++++++++++
        +++++=
   +++++++++++++++++++++++++++++
   +++++++++++++++++++++++++++++
    +++++++++++++++++++++++++++
     +++++++++++++++++++++++++
      +++++++++++++++++++++++
		`,
		Commands: []*cli.Command{
			{
				Name:    "auth",
				Aliases: []string{"a"},
				Usage:   "Authorize the CLI with your Signer Private Key and FID",
				Action: func(ctx *cli.Context) error {
					return auth.SetFidAndPrivateKey()
				},
			},
			{
				Name:    "login",
				Aliases: []string{"l"},
				Usage:   "Login with Farcaster mobile app via QR code",
				Action: func(ctx *cli.Context) error {
					return login.Login()
				},
			},
			{
				Name:    "new",
				Aliases: []string{"n"},
				Usage:   "Send a new Cast",
				Action: func(ctx *cli.Context) error {
					castData, err := compose.ComposeCast()
					if err != nil {
						return err
					}
					return compose.SendCast(castData)
				},
			},
			{
				Name:  "hub",
				Usage: "Set a preferred Hub",
				Action: func(ctx *cli.Context) error {
					return hub.SetHub()
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
