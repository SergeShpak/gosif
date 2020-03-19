package main

import (
	"log"
	"os"

	"github.com/SergeyShpak/gosif/generator"
)

func main() {
	if len(os.Args) < 2 {
		return
	}
	scriptsDir := os.Args[1]
	if err := generator.GenerateScriptsForDir(scriptsDir); err != nil {
		log.Println("[ERR]: ", err)
		os.Exit(1)
	}
}
