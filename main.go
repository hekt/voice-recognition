package main

import (
	"log"
	"os"

	"github.com/hekt/voice-recognition/actions/initialize"
	"github.com/hekt/voice-recognition/actions/recognize"
	"github.com/urfave/cli/v2"
)

func main() {
	// command flag values
	var (
		// general
		project    string
		recognizer string

		// initialize
		model string

		// recognize
		output string
	)

	generalFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "project",
			Aliases:     []string{"p"},
			Usage:       "Google Cloud Project ID",
			Required:    true,
			Destination: &project,
		},
		&cli.StringFlag{
			Name:        "recognizer",
			Aliases:     []string{"r"},
			Usage:       "Recognizer name",
			Required:    true,
			Destination: &recognizer,
		},
	}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "initialize",
				Aliases: []string{"init", "i"},
				Usage:   "create recognizer for Speech-to-Text API",
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
				Name:    "recognize",
				Aliases: []string{"r"},
				Usage:   "recognize voice",
				Flags: append(
					generalFlags,
					&cli.StringFlag{
						Name:        "output",
						Aliases:     []string{"o"},
						Usage:       "Output file path",
						Destination: &output,
					},
				),
				Action: func(c *cli.Context) error {
					recognizer, err := recognize.NewRecognizer(c.Context, project, recognizer, output, 0)
					if err != nil {
						return err
					}
					return recognizer.Run(c.Context)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
