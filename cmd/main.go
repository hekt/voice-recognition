package main

import (
	"log"
	"os"

	"github.com/hekt/voice-recognition/internal/app"
)

func main() {
	app := app.New()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
