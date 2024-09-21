package main

import (
	"log"
	"os"

	"github.com/hekt/voice-recognition/internal/actions/recognize"
	"github.com/hekt/voice-recognition/internal/actions/speechrecognizer"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "recognize",
				Usage: "recognize voice",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "project",
						Usage:    "Google Cloud Project ID",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "recognizer",
						Usage:    "Recognizer name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "Output file path",
					},
					&cli.DurationFlag{
						Name:  "interval",
						Usage: "Reconnect interval duration",
					},
					&cli.IntFlag{
						Name:  "buffersize",
						Usage: "Buffer size bytes",
					},
				},
				Action: func(cCtx *cli.Context) error {
					return recognize.Run(
						cCtx.Context,
						recognize.Arg{
							ProjectID:         cCtx.String("project"),
							RecognizerName:    cCtx.String("recognizer"),
							OutputFilePath:    cCtx.String("output"),
							BufferSize:        cCtx.Int("buffersize"),
							ReconnectInterval: cCtx.Duration("interval"),
						},
					)
				},
			},
			{
				Category: "manage",
				Name:     "recognizer-create",
				Usage:    "create recognizer for Speech-to-Text API",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "project",
						Usage:    "Google Cloud Project ID",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "recognizer",
						Usage:    "Recognizer name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "model",
						Usage: "Model for the recognizer",
						Value: "long",
					},
					&cli.StringFlag{
						Name:  "language",
						Usage: "Language Code for the recognizer",
						Value: "ja-jp",
					},
				},
				Action: func(cCtx *cli.Context) error {
					return speechrecognizer.Create(cCtx.Context, speechrecognizer.CreateArgs{
						ProjectID:      cCtx.String("project"),
						RecognizerName: cCtx.String("recognizer"),
						Model:          cCtx.String("model"),
						LanguageCode:   cCtx.String("language"),
					})
				},
			},
			{
				Category: "manage",
				Name:     "recognizer-delete",
				Usage:    "delete recognizer for Speech-to-Text API",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "project",
						Usage:    "Google Cloud Project ID",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "recognizer",
						Usage:    "Recognizer name",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					return speechrecognizer.Delete(cCtx.Context, speechrecognizer.DeleteArgs{
						ProjectID:      cCtx.String("project"),
						RecognizerName: cCtx.String("recognizer"),
					})
				},
			},
			{
				Category: "manage",
				Name:     "recognizer-list",
				Usage:    "list recognizers for Speech-to-Text API",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "project",
						Usage:    "Google Cloud Project ID",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					return speechrecognizer.List(cCtx.Context, speechrecognizer.ListArgs{
						ProjectID: cCtx.String("project"),
					})
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
