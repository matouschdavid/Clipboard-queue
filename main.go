package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/matouschdavid/Clipboard-queue/pkg/monitor"
)

var version = "v0.1.0" // overridden by -ldflags "-X main.version=..."

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	log.Printf("CBQ %s", version)
	monitor.Start()
}
