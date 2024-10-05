package app

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hekt/voice-recognition/internal/actions/manage/phraseset"
	"github.com/hekt/voice-recognition/internal/actions/manage/recognizer"
	"github.com/hekt/voice-recognition/internal/actions/recognize"
	"github.com/hekt/voice-recognition/internal/logger"
	"github.com/urfave/cli/v2"
)

func NewRecognizeCommand() *cli.Command {
	return &cli.Command{
		Name:  "recognize",
		Usage: "recognize voice",
		Flags: []cli.Flag{
			requiredProjectFlag,
			requiredRecognizerFlag,
			debugFlag,
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
			if cCtx.Bool(debugFlag.Name) {
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
					ProjectID:      cCtx.String(projectFlag.Name),
					RecognizerName: cCtx.String(recognizerFlag.Name),
				},
				options...,
			)
		},
	}
}

func NewRecognizerCreateCommand() *cli.Command {
	return &cli.Command{
		Category: "manage",
		Name:     "recognizer-create",
		Usage:    "create recognizer for Speech-to-Text API",
		Flags: []cli.Flag{
			requiredProjectFlag,
			requiredRecognizerFlag,
			modelFlag,
			languageCodeFlag,
			phraseSetFlag,
		},
		Action: func(cCtx *cli.Context) error {
			// TODO support multiple language code
			languageCode := ""
			if fs := cCtx.StringSlice(languageCodeFlag.Name); len(fs) > 0 {
				languageCode = fs[0]
			}

			return recognizer.Create(cCtx.Context, recognizer.CreateArgs{
				ProjectID:      cCtx.String(projectFlag.Name),
				RecognizerName: cCtx.String(recognizerFlag.Name),
				Model:          cCtx.String(modelFlag.Name),
				LanguageCode:   languageCode,
				PhraseSet:      cCtx.String(phraseSetFlag.Name),
			})
		},
	}
}

func NewRecognizerDeleteCommand() *cli.Command {
	return &cli.Command{
		Category: "manage",
		Name:     "recognizer-delete",
		Usage:    "delete recognizer for Speech-to-Text API",
		Flags: []cli.Flag{
			requiredProjectFlag,
			requiredRecognizerFlag,
		},
		Action: func(cCtx *cli.Context) error {
			return recognizer.Delete(cCtx.Context, recognizer.DeleteArgs{
				ProjectID:      cCtx.String(projectFlag.Name),
				RecognizerName: cCtx.String(recognizerFlag.Name),
			})
		},
	}
}

func NewRecognizerListCommand() *cli.Command {
	return &cli.Command{
		Category: "manage",
		Name:     "recognizer-list",
		Usage:    "list recognizers for Speech-to-Text API",
		Flags: []cli.Flag{
			requiredProjectFlag,
		},
		Action: func(cCtx *cli.Context) error {
			return recognizer.List(cCtx.Context, recognizer.ListArgs{
				ProjectID: cCtx.String(projectFlag.Name),
			})
		},
	}
}

func NewPhraseSetCreateCommand() *cli.Command {
	return &cli.Command{
		Category: "manage",
		Name:     "phrase-set-create",
		Usage:    "create phrase set for Speech-to-Text API",
		Flags: []cli.Flag{
			requiredProjectFlag,
			requiredPhraseSetFlag,
			phraseFlag,
			phrasesFlag,
			boostFlag,
		},
		Action: func(cCtx *cli.Context) error {
			rawPhrases := append(
				strings.Split(cCtx.String(phrasesFlag.Name), ","),
				cCtx.StringSlice(phraseFlag.Name)...,
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
				ProjectID:     cCtx.String(projectFlag.Name),
				PhraseSetName: cCtx.String(phraseSetFlag.Name),
				Phrases:       phrases,
				Boost:         float32(cCtx.Float64(boostFlag.Name)),
			})
		},
	}
}

func NewPhraseSetUpdateCommand() *cli.Command {
	return &cli.Command{
		Category: "manage",
		Name:     "phrase-set-update",
		Usage:    "update phrase set for Speech-to-Text API",
		Flags: []cli.Flag{
			requiredProjectFlag,
			requiredPhraseSetFlag,
			phraseFlag,
			phrasesFlag,
			boostFlag,
		},
		Action: func(cCtx *cli.Context) error {
			rawPhrases := append(
				strings.Split(cCtx.String(phrasesFlag.Name), ","),
				cCtx.StringSlice(phraseFlag.Name)...,
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
				ProjectID:     cCtx.String(projectFlag.Name),
				PhraseSetName: cCtx.String(phraseSetFlag.Name),
				Phrases:       phrases,
				Boost:         float32(cCtx.Float64(boostFlag.Name)),
			})
		},
	}
}

func NewPhraseSetListCommand() *cli.Command {
	return &cli.Command{
		Category: "manage",
		Name:     "phrase-set-list",
		Usage:    "list phrase sets for Speech-to-Text API",
		Flags: []cli.Flag{
			requiredProjectFlag,
		},
		Action: func(cCtx *cli.Context) error {
			return phraseset.List(cCtx.Context, phraseset.ListArgs{
				ProjectID: cCtx.String(projectFlag.Name),
			})
		},
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
