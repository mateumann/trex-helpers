package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/akamensky/argparse"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type trexPcapStat struct {
	timestamp time.Time
	genre     uint8
	latency   int64
}

func parsePcap(filename string, verbose bool) (stats []trexPcapStat, err error) {
	var handle *pcap.Handle
	var stat trexPcapStat
	if handle, err = pcap.OpenOffline(filename); err != nil {
		return make([]trexPcapStat, 0), err
	}
	if verbose {
		fmt.Printf("Opening file %s with %s handler... ", filepath.Base(filename), handle.LinkType())
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	notParsedPackets := 0
	for packet := range packetSource.Packets() {
		if stat, err = handlePacket(packet); err != nil {
			notParsedPackets++
		} else {
			stats = append(stats, stat)
		}
	}
	if verbose {
		fmt.Printf("parsing is finished, skipped %d packets.\n", notParsedPackets)
	}
	return stats, nil
}

func handlePacket(packet gopacket.Packet) (stat trexPcapStat, err error) {
	var ipLayer gopacket.Layer
	if ipLayer = packet.Layer(layers.LayerTypeIPv4); ipLayer == nil {
		return stat, errors.New("not an IPv4 packet")
	}
	stat.timestamp = packet.Metadata().Timestamp
	stat.latency = 0
	switch stat.genre = ipLayer.LayerPayload()[14]; stat.genre {
	case 0xab:
		ns := int64(binary.LittleEndian.Uint64(ipLayer.LayerPayload()[22:30]))
		stat.latency = stat.timestamp.UnixNano() - ns
	case 0x58:
		break
	default:
		return stat, errors.New("not interesting packet")
	}
	return stat, nil
}

func plotChart(filename string, verbose bool, stats []trexPcapStat) (err error) {
	if verbose {
		fmt.Printf("Generating chart in %s...", filepath.Base(filename))
	}
	// TODO
	return nil
}

func parseCmdArgs() (string, string, bool) {
	parser := argparse.NewParser("trex-helper", "")
	parser.HelpFunc = func(c *argparse.Command, msg interface{}) string {
		helpString := fmt.Sprintf("Name; %s\n", c.GetName())
		for _, com := range c.GetCommands() {
			// Calls parser.HelpFunc, because command.HelpFuncs are nil
			helpString += com.Help(nil)
		}
		return helpString
	}
	pcapFn := parser.String("f", "filename", &argparse.Options{Required: true, Help: "Filename to be analyzed"})
	outputFn := parser.String("o", "output",
		&argparse.Options{Required: false, Help: "Filename for the generated chart", Default: "plot.svg"})
	verbose := parser.Flag("v", "verbose", &argparse.Options{Required: false, Help: "Be verbose", Default: true})
	if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}
	return *pcapFn, *outputFn, *verbose
}

func main() {
	pcapFilename, outputFilename, verbose := parseCmdArgs()
	var packets []trexPcapStat
	var err error
	if packets, err = parsePcap(pcapFilename, verbose); err != nil {
		fmt.Printf("File %s could not be read\n", pcapFilename)
		os.Exit(1)
	}
	if verbose {
		fmt.Printf("File parsed; %d packets read\n", len(packets))
	}
	if err = plotChart(outputFilename, verbose, packets); err != nil {
		fmt.Printf("Unable to generate chart %s\n", outputFilename)
		os.Exit(1)
	}
}
