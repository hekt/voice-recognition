package app

import (
	"github.com/urfave/cli/v2"
)

func New() *cli.App {
	return &cli.App{
		Commands: []*cli.Command{
			recognizeCommand,
			voskRecognizeCommand,
			recognizerCreateCommand,
			recognizerDeleteCommand,
			recognizerListCommand,
			phraseSetCreateCommand,
			phraseSetUpdateCommand,
			phraseSetListCommand,
		},
	}
}
