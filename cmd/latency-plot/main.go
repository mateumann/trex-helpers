package main

import (
	"flag"
	"fmt"
	"os"

	"path/filepath"

	"trex-helpers/pkg/packet"
)

func main() {
	debug := false

	if os.Getenv("DEBUG") != "" {
		debug = true
	}

	inputFilenamePtr := flag.String("input", "", "Input pcap file. (Required)")
	outputFilenamePtr := flag.String("output", "plot.svg", "Output svg chart.")
	flag.Parse()

	if *inputFilenamePtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if debug {
		fmt.Printf("Running %v, analyzing %s → %s\n", filepath.Base(os.Args[0]), *inputFilenamePtr, *outputFilenamePtr)
	}

	_, err := packet.ParsePcap(*inputFilenamePtr, debug)
	if err != nil {
		fmt.Printf("Could not parse %v file\n", *inputFilenamePtr)
	}
}
