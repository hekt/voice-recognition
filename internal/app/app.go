package app

import (
	"github.com/urfave/cli/v2"
)

func New() *cli.App {
	return &cli.App{
		Commands: []*cli.Command{
			NewRecognizeCommand(),
			NewVoskRecognizeCommand(),
			NewRecognizerCreateCommand(),
			NewRecognizerDeleteCommand(),
			NewRecognizerListCommand(),
			NewPhraseSetCreateCommand(),
			NewPhraseSetUpdateCommand(),
			NewPhraseSetListCommand(),
		},
	}
}
