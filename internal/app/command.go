package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	speech "cloud.google.com/go/speech/apiv2"
	"github.com/hekt/voice-recognition/internal/file"
	"github.com/hekt/voice-recognition/internal/logger"
	"github.com/hekt/voice-recognition/internal/recognizer"
	"github.com/hekt/voice-recognition/internal/resource"
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
				Value: fmt.Sprintf("output/%d.json", time.Now().Unix()),
			},
			&cli.DurationFlag{
				Name:  "interval",
				Usage: "Reconnect interval duration",
				Value: time.Minute,
			},
			&cli.IntFlag{
				Name:  "buffersize",
				Usage: "Buffer size bytes",
				Value: 1024,
			},
			&cli.DurationFlag{
				Name:  "timeout",
				Usage: "Inactive timeout duration",
				Value: 5 * time.Minute,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.Bool(debugFlag.Name) {
				if err := setLogger(slog.LevelDebug); err != nil {
					return fmt.Errorf("failed to set logger: %w", err)
				}
			}

			// This behavior ensures the output file is created early,
			// making it easier to use with tools like `tail -f`.
			if err := prepareOutputFile(cCtx.String("output")); err != nil {
				return fmt.Errorf("failed to prepare output file: %w", err)
			}

			audioReader := os.Stdin
			resultWriter := file.NewOpenCloseFileWriter(
				cCtx.String("output"),
				os.O_APPEND|os.O_CREATE|os.O_WRONLY,
				os.FileMode(0o644),
			)
			interimWriter := os.Stdout
			client, err := speech.NewClient(cCtx.Context)
			if err != nil {
				return fmt.Errorf("failed to create speech client: %w", err)
			}

			recognizer, err := recognizer.New(
				cCtx.String(projectFlag.Name),
				cCtx.String(recognizerFlag.Name),
				cCtx.Duration("interval"),
				cCtx.Int("buffersize"),
				cCtx.Duration("timeout"),
				client,
				audioReader,
				resultWriter,
				interimWriter,
			)
			if err != nil {
				return fmt.Errorf("failed to create recognizer: %w", err)
			}

			if err := recognizer.Start(cCtx.Context); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return fmt.Errorf("failed to start recognizer: %w", err)
			}

			return nil
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
			manager, err := buildRecognizerManager(cCtx.Context)
			if err != nil {
				return fmt.Errorf("failed to build recognizer manager: %w", err)
			}

			// TODO support multiple language code
			languageCode := ""
			if fs := cCtx.StringSlice(languageCodeFlag.Name); len(fs) > 0 {
				languageCode = fs[0]
			}

			args := resource.CreateRecognizerArgs{
				ProjectID:      cCtx.String(projectFlag.Name),
				RecognizerName: cCtx.String(recognizerFlag.Name),
				Model:          cCtx.String(modelFlag.Name),
				LanguageCode:   languageCode,
				PhraseSet:      cCtx.String(phraseSetFlag.Name),
			}
			if err := manager.Create(cCtx.Context, args); err != nil {
				return fmt.Errorf("failed to create recognizer: %w", err)
			}

			fmt.Println("Recognizer created")

			return nil
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
			manager, err := buildRecognizerManager(cCtx.Context)
			if err != nil {
				return fmt.Errorf("failed to build recognizer manager: %w", err)
			}

			args := resource.DeleteRecognizerArgs{
				ProjectID:      cCtx.String(projectFlag.Name),
				RecognizerName: cCtx.String(recognizerFlag.Name),
			}
			if err := manager.Delete(cCtx.Context, args); err != nil {
				return fmt.Errorf("failed to delete recognizer: %w", err)
			}

			fmt.Println("Recognizer deleted")

			return nil
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
			manager, err := buildRecognizerManager(cCtx.Context)
			if err != nil {
				return fmt.Errorf("failed to build recognizer manager: %w", err)
			}

			args := resource.ListRecognizerArgs{
				ProjectID: cCtx.String(projectFlag.Name),
			}
			recognizers, err := manager.List(cCtx.Context, args)
			if err != nil {
				return fmt.Errorf("failed to list recognizers: %w", err)
			}

			for _, recognizer := range recognizers {
				fmt.Println(recognizer.Value)
			}

			return nil
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
			manager, err := buildPhraseSetManager(cCtx.Context)
			if err != nil {
				return fmt.Errorf("failed to build phrase set manager: %w", err)
			}

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

			args := resource.CreatePhraseSetArgs{
				ProjectID:     cCtx.String(projectFlag.Name),
				PhraseSetName: cCtx.String(phraseSetFlag.Name),
				Phrases:       phrases,
				Boost:         float32(cCtx.Float64(boostFlag.Name)),
			}
			if err := manager.Create(cCtx.Context, args); err != nil {
				return fmt.Errorf("failed to create phrase set: %w", err)
			}

			fmt.Println("Phrase set created")

			return nil
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
			manager, err := buildPhraseSetManager(cCtx.Context)
			if err != nil {
				return fmt.Errorf("failed to build phrase set manager: %w", err)
			}

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

			args := resource.UpdatePhraseSetArgs{
				ProjectID:     cCtx.String(projectFlag.Name),
				PhraseSetName: cCtx.String(phraseSetFlag.Name),
				Phrases:       phrases,
				Boost:         float32(cCtx.Float64(boostFlag.Name)),
			}
			if err := manager.Update(cCtx.Context, args); err != nil {
				return fmt.Errorf("failed to update phrase set: %w", err)
			}

			fmt.Println("Phrase set updated")

			return nil
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
			manager, err := buildPhraseSetManager(cCtx.Context)
			if err != nil {
				return fmt.Errorf("failed to build phrase set manager: %w", err)
			}

			args := resource.ListPhraseSetArgs{
				ProjectID: cCtx.String(projectFlag.Name),
			}
			phraseSets, err := manager.List(cCtx.Context, args)
			if err != nil {
				return fmt.Errorf("failed to list phrase sets: %w", err)
			}

			for _, phraseSet := range phraseSets {
				fmt.Println(phraseSet.Value)
			}

			return nil
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

func buildRecognizerManager(ctx context.Context) (resource.RecognizerManager, error) {
	client, err := speech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create speech client: %w", err)
	}

	manager := resource.NewRecognizerManager(client)

	return manager, nil
}

func buildPhraseSetManager(ctx context.Context) (resource.PhraseSetManager, error) {
	client, err := speech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create speech client: %w", err)
	}

	manager := resource.NewPhraseSetManager(client)

	return manager, nil
}

func prepareOutputFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE, os.FileMode(0o644))
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close output file: %w", err)
	}
	return nil
}
