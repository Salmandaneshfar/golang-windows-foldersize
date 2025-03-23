package main

import (
	"flag"
)

func main() {
	// Check for special flag to force CLI mode
	cliMode := flag.Bool("cli", false, "Run in command-line mode")
	flag.Parse()

	// Check if we're running in CLI mode
	if *cliMode {
		// Run the CLI version
		runCLI()
		return
	}

	// Otherwise run the UI version
	runUI()
}
