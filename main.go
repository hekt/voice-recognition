package main

import (
	"log"
	"os"
	"time"

	"github.com/hekt/voice-recognition/actions/initialize"
	"github.com/hekt/voice-recognition/actions/recognize"
	"github.com/urfave/cli/v2"
)

func main() {
	// command flag destinations
	var (
		// general
		project    string
		recognizer string

		// initialize
		model string

		// recognize
		output   string
		interval time.Duration
		buffer   int
	)

	generalFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "project",
			Usage:       "Google Cloud Project ID",
			Required:    true,
			Destination: &project,
		},
		&cli.StringFlag{
			Name:        "recognizer",
			Usage:       "Recognizer name",
			Required:    true,
			Destination: &recognizer,
		},
	}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "initialize",
				Usage: "create recognizer for Speech-to-Text API",
				Flags: append(
					generalFlags,
					&cli.StringFlag{
						Name:        "model",
						Aliases:     []string{"m"},
						Usage:       "Model for the recognizer",
						Value:       "long",
						Destination: &model,
					},
				),
				Action: func(c *cli.Context) error {
					initialize.Run(c.Context, initialize.Args{
						ProjectID:      project,
						RecognizerName: recognizer,
						Model:          model,
					})
					return nil
				},
			},
			{
				Name:  "recognize",
				Usage: "recognize voice",
				Flags: append(
					generalFlags,
					&cli.StringFlag{
						Name:        "output",
						Usage:       "Output file path",
						Destination: &output,
					},
					&cli.DurationFlag{
						Name:        "interval",
						Usage:       "Reconnect interval duration",
						Destination: &interval,
					},
					&cli.IntFlag{
						Name:        "buffersize",
						Usage:       "Buffer size bytes",
						Destination: &buffer,
					},
				),
				Action: func(c *cli.Context) error {
					return recognize.Run(
						c.Context,
						recognize.Arg{
							ProjectID:         project,
							RecognizerName:    recognizer,
							OutputFilePath:    output,
							BufferSize:        buffer,
							ReconnectInterval: interval,
						},
					)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
