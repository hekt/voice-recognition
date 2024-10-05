package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/hekt/voice-recognition/internal/actions/manage/phraseset"
	"github.com/hekt/voice-recognition/internal/actions/manage/recognizer"
	"github.com/hekt/voice-recognition/internal/actions/recognize"
	"github.com/hekt/voice-recognition/internal/logger"
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
					&cli.BoolFlag{
						Name:  "debug",
						Usage: "Enable debug log",
						Value: false,
					},
				},
				Action: func(cCtx *cli.Context) error {
					if cCtx.Bool("debug") {
						if err := setLogger(slog.LevelDebug); err != nil {
							return fmt.Errorf("failed to set logger: %w", err)
						}
					}

					options := make([]recognize.Option, 0, 3)
					if cCtx.IsSet("output") {
						options = append(options, recognize.WithOutputFilePath(cCtx.String("output")))
					}
					if cCtx.IsSet("interval") {
						options = append(options, recognize.WithReconnectInterval(cCtx.Duration("interval")))
					}
					if cCtx.IsSet("buffersize") {
						options = append(options, recognize.WithBufferSize(cCtx.Int("buffersize")))
					}

					return recognize.Run(
						cCtx.Context,
						recognize.Args{
							ProjectID:      cCtx.String("project"),
							RecognizerName: cCtx.String("recognizer"),
						},
						options...,
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
					&cli.StringFlag{
						Name:  "phraseset",
						Usage: "Phrase set to use for the recognizer",
					},
				},
				Action: func(cCtx *cli.Context) error {
					return recognizer.Create(cCtx.Context, recognizer.CreateArgs{
						ProjectID:      cCtx.String("project"),
						RecognizerName: cCtx.String("recognizer"),
						Model:          cCtx.String("model"),
						LanguageCode:   cCtx.String("language"),
						PhraseSet:      cCtx.String("phraseset"),
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
					return recognizer.Delete(cCtx.Context, recognizer.DeleteArgs{
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
					return recognizer.List(cCtx.Context, recognizer.ListArgs{
						ProjectID: cCtx.String("project"),
					})
				},
			},
			{
				Category: "manage",
				Name:     "phrase-set-create",
				Usage:    "create phrase set for Speech-to-Text API",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "project",
						Usage:    "Google Cloud Project ID",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Phrase set name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "phrases",
						Usage: "Commma separated phrases to add to the phrase set",
					},
					&cli.StringSliceFlag{
						Name:    "phrase",
						Aliases: []string{"p"},
						Usage:   "Phrase to add to the phrase set possibly multiple",
					},
					&cli.Float64Flag{
						Name:  "boost",
						Usage: "Boost value for the phrase set",
						Value: 0,
					},
				},
				Action: func(cCtx *cli.Context) error {
					rawPhrases := append(
						strings.Split(cCtx.String("phrases"), ","),
						cCtx.StringSlice("phrase")...,
					)
					phrases := make([]string, 0, len(rawPhrases))
					for _, phrase := range rawPhrases {
						trimed := strings.TrimSpace(phrase)
						if trimed != "" {
							phrases = append(phrases, trimed)
						}
					}
					if len(phrases) == 0 {
						return fmt.Errorf("no valid phrases provided")
					}

					return phraseset.Create(cCtx.Context, phraseset.CreateArgs{
						ProjectID:     cCtx.String("project"),
						PhraseSetName: cCtx.String("name"),
						Phrases:       phrases,
						Boost:         float32(cCtx.Float64("boost")),
					})
				},
			},
			{
				Category: "manage",
				Name:     "phrase-set-update",
				Usage:    "update phrase set for Speech-to-Text API",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "project",
						Usage:    "Google Cloud Project ID",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Phrase set name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "phrases",
						Usage: "Commma separated phrases to add to the phrase set",
					},
					&cli.StringSliceFlag{
						Name:    "phrase",
						Aliases: []string{"p"},
						Usage:   "Phrase to add to the phrase set possibly multiple",
					},
					&cli.Float64Flag{
						Name:  "boost",
						Usage: "Boost value for the phrase set",
						Value: 0,
					},
				},
				Action: func(cCtx *cli.Context) error {
					rawPhrases := append(
						strings.Split(cCtx.String("phrases"), ","),
						cCtx.StringSlice("phrase")...,
					)
					phrases := make([]string, 0, len(rawPhrases))
					for _, phrase := range rawPhrases {
						trimed := strings.TrimSpace(phrase)
						if trimed != "" {
							phrases = append(phrases, trimed)
						}
					}
					if len(phrases) == 0 {
						return fmt.Errorf("no valid phrases provided")
					}

					return phraseset.Update(cCtx.Context, phraseset.UpdateArgs{
						ProjectID:     cCtx.String("project"),
						PhraseSetName: cCtx.String("name"),
						Phrases:       phrases,
						Boost:         float32(cCtx.Float64("boost")),
					})
				},
			},
			{
				Category: "manage",
				Name:     "phrase-set-list",
				Usage:    "list phrase sets for Speech-to-Text API",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "project",
						Usage:    "Google Cloud Project ID",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					return phraseset.List(cCtx.Context, phraseset.ListArgs{
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

func setLogger(level slog.Level) error {
	logger, err := logger.NewFileLogger(
		fmt.Sprintf("output/log-%d.log", time.Now().Unix()),
		level,
	)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	slog.SetDefault(logger)
	return nil
}
