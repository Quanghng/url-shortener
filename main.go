package main

import (
	"github.com/Quanghng/url-shortener/cmd"
	_ "github.com/Quanghng/url-shortener/cmd/cli"
	_ "github.com/Quanghng/url-shortener/cmd/server"
)

func main() {
	// Exécute la commande racine Cobra
	cmd.Execute()
}
