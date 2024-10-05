package app

import "github.com/urfave/cli/v2"

var projectFlag = &cli.StringFlag{
	Name:  "project",
	Usage: "Google Cloud Project ID",
}

var requiredProjectFlag = &cli.StringFlag{
	Name:     projectFlag.Name,
	Usage:    projectFlag.Usage,
	Required: true,
}

var debugFlag = &cli.BoolFlag{
	Name:  "debug",
	Usage: "Enable debug log",
	Value: false,
}

//
// Recognizer flags
//

var recognizerFlag = &cli.StringFlag{
	Name:  "recognizer",
	Usage: "Recognizer name",
}

var requiredRecognizerFlag = &cli.StringFlag{
	Name:     recognizerFlag.Name,
	Usage:    recognizerFlag.Usage,
	Required: true,
}

var modelFlag = &cli.StringFlag{
	Name:  "model",
	Usage: "Model name",
}

var languageCodeFlag = &cli.StringSliceFlag{
	Name:    "language-code",
	Aliases: []string{"l"},
	Usage:   "Language code possibly multiple",
}

//
// Phrase set flags
//

var phraseSetFlag = &cli.StringFlag{
	Name:  "name",
	Usage: "Phrase set name",
}

var requiredPhraseSetFlag = &cli.StringFlag{
	Name:     phraseSetFlag.Name,
	Usage:    phraseSetFlag.Usage,
	Required: true,
}

var phraseFlag = &cli.StringSliceFlag{
	Name:    "phrase",
	Aliases: []string{"p"},
	Usage:   "Phrase to add to the phrase set possibly multiple",
}

var phrasesFlag = &cli.StringFlag{
	Name:  "phrases",
	Usage: "Commma separated phrases to add to the phrase set",
}

var boostFlag = &cli.Float64Flag{
	Name:  "boost",
	Usage: "Boost value for the phrase set",
	Value: 0,
}
