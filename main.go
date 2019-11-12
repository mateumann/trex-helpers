package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/util"

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
	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	totalPackets, notParsedPackets, latencyPackets := 0, 0, 0
	for packet := range packetSource.Packets() {
		totalPackets++
		if stat, err = handlePacket(packet); err != nil {
			notParsedPackets++
		} else {
			stats = append(stats, stat)
			if stat.genre == 0xab {
				latencyPackets++
			}
		}
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
	graph := chart.Chart{
		Background: chart.Style{
			Padding: chart.Box{
				Top:    50,
				Bottom: 50,
				Left:   50,
				Right:  50,
			},
		},
		XAxis: chart.XAxis{
			Name:  "time [ns]",
			Ticks: statsTimestampsTicks(stats),
		},
		YAxis: chart.YAxis{
			Name: "latency [µs]",
		},
		Series: []chart.Series{
			chart.TimeSeries{
				//Style: chart.Style{
				//	StrokeColor: chart.GetDefaultColor(0).WithAlpha(64),
				//	FillColor:   chart.GetDefaultColor(0).WithAlpha(64),
				//},
				//XValues: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
				XValues: statsTimestamps(stats),
				//YValues: []float64{2.2, 2.0, 3.0, 4.0, 3.8},
				YValues: statsLatencies(stats),
			},
		},
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	if err = graph.Render(chart.SVG, f); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if verbose {
		fmt.Printf("done.\n")
	}
	return nil
}

func statsTimestamps(stats []trexPcapStat) (timestamps []time.Time) {
	for _, stat := range stats {
		timestamps = append(timestamps, stat.timestamp)
	}
	return timestamps
}

func statsTimestampsTicks(stats []trexPcapStat) (timestampTicks []chart.Tick) {
	minTimestamp := minStatsTimestamp(stats)
	every := len(stats) / 4
	for i, stat := range stats {
		if i%every == 0 {
			nanoseconds := stat.timestamp.Sub(minTimestamp)
			fmt.Printf("Nanoseconds = %s\n", nanoseconds.String())
			timestampTicks = append(timestampTicks,
				chart.Tick{Value: util.Time.ToFloat64(stat.timestamp), Label: nanoseconds.String()})
		}
	}
	return timestampTicks
}

func minStatsTimestamp(stats []trexPcapStat) (minTimestamp time.Time) {
	minTimestamp = time.Unix(math.MaxInt32, math.MaxInt32)
	for _, stat := range stats {
		if minTimestamp.After(stat.timestamp) {
			minTimestamp = stat.timestamp
		}
	}
	return minTimestamp
}

func statsLatencies(stats []trexPcapStat) (latencies []float64) {
	for _, stat := range stats {
		latencies = append(latencies, float64(stat.latency))
	}
	return latencies
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
