package app

import "github.com/urfave/cli/v2"

var projectFlag = &cli.StringFlag{
	Name:  "project",
	Usage: "Google Cloud Project ID",
}

var requiredProjectFlag = &cli.StringFlag{
	Name:     "project",
	Usage:    "Google Cloud Project ID",
	Required: true,
}

var recognizerFlag = &cli.StringFlag{
	Name:  "recognizer",
	Usage: "Recognizer name",
}

var requiredRecognizerFlag = &cli.StringFlag{
	Name:     "recognizer",
	Usage:    "Recognizer name",
	Required: true,
}

var debugFlag = &cli.BoolFlag{
	Name:  "debug",
	Usage: "Enable debug log",
	Value: false,
}
