package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"github.com/akamensky/argparse"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type trexPcapStat struct {
	timestamp time.Time
	seconds   float64
	genre     uint8
	latency   int64
}

func parsePcap(filename string, verbose bool) (stats []trexPcapStat, err error) {
	var handle *pcap.Handle
	if handle, err = pcap.OpenOffline(filename); err != nil {
		return make([]trexPcapStat, 0), err
	}
	if verbose {
		fmt.Printf("Opening file %s with %s handler... ", filepath.Base(filename),
			handle.LinkType())
	}
	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	totalPackets, notParsedPackets, latencyPackets := 0, 0, 0
	var stat trexPcapStat
	minTimestamp := time.Unix(1<<32, 1<<32)
	for packet := range packetSource.Packets() {
		totalPackets++
		if stat, err = handlePacket(packet); err != nil {
			notParsedPackets++
		} else {
			stats = append(stats, stat)
			if stat.genre == 0xab {
				latencyPackets++
			}
			if minTimestamp.After(stat.timestamp) {
				minTimestamp = stat.timestamp
			}
		}
	}

	for i, stat := range stats {
		stats[i].seconds = stat.timestamp.Sub(minTimestamp).Seconds()
	}

	if verbose {
		fmt.Printf("parsing is finished: %d total pkts, skipped %d pkts, %d latency pkts.\n",
			totalPackets, notParsedPackets, latencyPackets)
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

func plotChart(filename string, verbose bool, pcapFilename string, stats []trexPcapStat) (err error) {
	if verbose {
		fmt.Printf("Generating chart in %s... ", filepath.Base(filename))
	}

	p, err := plot.New()
	if err != nil {
		return err
	}
	lMin, lMax := statsLatencyMinMax(stats)
	p.Title.Text = fmt.Sprintf("TRex latency plot for %s (min = %d ns, max = %d ns)",
		filepath.Base(pcapFilename), lMin, lMax)
	p.X.Label.Text = "time [s]"
	p.Y.Label.Text = "latency [ns]"
	p.Add(plotter.NewGrid())
	b, err := plotter.NewBarChart(statsLatencyValues(stats), vg.Points(1))
	if err != nil {
		return nil
	}
	b.LineStyle.Width = vg.Length(0)
	b.Color = color.RGBA{R: 64, G: 128, B: 255, A: 255}
	p.Add(b)
	p.NominalX(statsSeconds(stats)...)
	if err = p.Save(18*vg.Inch, 4*vg.Inch, filename); err != nil {
		return nil
	}

	if verbose {
		fmt.Printf("done.\n")
	}
	return nil
}

func statsLatencyValues(stats []trexPcapStat) (latencies plotter.Values) {
	for _, stat := range stats {
		if stat.genre == 0xab {
			latencies = append(latencies, float64(stat.latency))
		}
	}
	return latencies
}

func statsSeconds(stats []trexPcapStat) (seconds []string) {
	var prev trexPcapStat
	for i, stat := range stats {
		if stat.genre == 0xab {
			if i == 0 || (math.Round(2.0*stat.seconds) != math.Round(2.0*prev.seconds)) {
				seconds = append(seconds, fmt.Sprintf("%.3f", stat.seconds))
			} else {
				seconds = append(seconds, "")
			}
			prev = stat
		}
	}
	return seconds
}

func statsLatencyMinMax(stats []trexPcapStat) (min int64, max int64) {
	min = 1 << 32
	max = -min
	for _, stat := range stats {
		if min > stat.latency && stat.latency > 0 {
			min = stat.latency
		}
		if max < stat.latency {
			max = stat.latency
		}
	}
	return
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
	if err = plotChart(outputFilename, verbose, pcapFilename, packets); err != nil {
		fmt.Printf("Unable to generate chart %s\n", outputFilename)
		os.Exit(1)
	}
}
