package main

import (
	"flag"
	"fmt"
	"os"
	"trex-helpers/pkg/plot"

	"path/filepath"

	"trex-helpers/pkg/packet"
)

func main() {
	debug := false

	if os.Getenv("DEBUG") != "" {
		debug = true
	}

	inputFilenamePtr := flag.String("input", "", "Input pcap file. (Required)")
	outputFilenamePtr := flag.String("output", "plot.pdf", "Output a chart.")
	flag.Parse()

	if *inputFilenamePtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if debug {
		fmt.Printf("Running %v, analyzing %s → %s\n", filepath.Base(os.Args[0]), *inputFilenamePtr, *outputFilenamePtr)
	}

	packets, err := packet.ParsePcap(*inputFilenamePtr, debug)
	if err != nil {
		fmt.Printf("Could not parse %v file: %v\n", *inputFilenamePtr, err)
	}

	if err = plot.SavePDF(packets, filepath.Base(*inputFilenamePtr), *outputFilenamePtr, debug); err != nil {
		fmt.Printf("Could not plot %v file: %v\n", *outputFilenamePtr, err)
	}
}
