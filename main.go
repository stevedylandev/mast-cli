package main

import (
	"fmt"
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "message",
						Aliases: []string{"m"},
						Usage:   "Cast message text",
					},
					&cli.StringFlag{
						Name:    "url",
						Aliases: []string{"u"},
						Usage:   "URL to embed in the cast",
					},
					&cli.StringFlag{
						Name:    "url2",
						Aliases: []string{"u2"},
						Usage:   "Second URL to embed in the cast",
					},
					&cli.StringFlag{
						Name:    "channel",
						Aliases: []string{"c"},
						Usage:   "Channel ID for the cast",
					},
				},
				Action: func(ctx *cli.Context) error {
					message := ctx.String("message")
					url1 := ctx.String("url")
					url2 := ctx.String("url2")
					channel := ctx.String("channel")

					if message != "" || url1 != "" || url2 != "" || channel != "" {
						castData := compose.CastData{
							Message: message,
							URL1:    url1,
							URL2:    url2,
							Channel: channel,
						}

						if message == "" && url1 == "" && url2 == "" {
							return fmt.Errorf("at least a message or URL must be provided")
						}

						return compose.SendCast(castData)
					}

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
