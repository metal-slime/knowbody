package main

import (
	"flag"
	"log"
	"os"

	knowbody "github.com/jeefy/knowbody/pkg"
)

func main() {
	lint := flag.Bool("lint", false, "Lint a config then exit")
	flag.Parse()

	if *lint {
		log.Print("Linting config files...")
		knowbody.Lint()
		log.Print("Success!")
		os.Exit(0)
	}

	knowbody.Start()
}
