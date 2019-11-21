package main

import (
	"fmt"
	"os"

	"path/filepath"
)

func main() {
	debug := false

	if os.Getenv("DEBUG") != "" {
		debug = true
	}

	if debug {
		fmt.Printf("Running %v\n", filepath.Base(os.Args[0]))
	}
}
