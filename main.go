package main

import (
	"github.com/Quanghng/url-shortener/cmd"
	_ "github.com/Quanghng/url-shortener/cmd/cli"    // Importe le package 'cli' pour que ses init() soient exécutés
	// _ "github.com/Quanghng/url-shortener/cmd/server" // Temporairement commenté - handlers.go a build ignore
)

func main() {
	// Exécute la commande racine Cobra
	cmd.Execute()
}
