package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type TrexPcapStat struct {
	timestamp time.Time
	genre     uint8
	latency   int64
}

func parsePcap(filename string) (stats []TrexPcapStat, err error) {
	var handle *pcap.Handle
	var stat TrexPcapStat
	if handle, err = pcap.OpenOffline(filename); err != nil {
		return make([]TrexPcapStat, 0), err
	}
	fmt.Printf("Opening file %s with %s handler.\n", filename, handle.LinkType())
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	notParsedPackets := 0
	for packet := range packetSource.Packets() {
		if stat, err = handlePacket(packet); err != nil {
			notParsedPackets++
		} else {
			stats = append(stats, stat)
		}
	}
	fmt.Printf("Parsing finished, skipped %d packets.\n", notParsedPackets)
	return stats, nil
}

func handlePacket(packet gopacket.Packet) (stat TrexPcapStat, err error) {
	var ipLayer gopacket.Layer
	if ipLayer = packet.Layer(layers.LayerTypeIPv4); ipLayer == nil {
		return stat, errors.New("not an IPv4 packet")
	}
	stat.timestamp = packet.Metadata().Timestamp
	switch stat.genre = ipLayer.LayerPayload()[14]; stat.genre {
	case 0xab:
		ns := int64(binary.LittleEndian.Uint64(ipLayer.LayerPayload()[22:30]))
		stat.latency = stat.timestamp.UnixNano() - ns
	case 0x58:
		stat.latency = 0
	default:
		return stat, errors.New("not interesting packet")
	}
	return stat, nil
}

func main() {
	//filename := "/home/mateusz/Projects/trex-helpers/data/001-br-15397ec68cd0.pcap"
	//filename := "/home/mateusz/Projects/trex-helpers/data/001-br-51d85eb72d54.pcap"
	//filename := "/home/mateusz/Projects/trex-helpers/data/002-br-571f91c727a3.pcap"
	//filename := "/home/mateusz/Projects/trex-helpers/data/002-br-e0e2570b862d.pcap"
	//filename := "/home/mateusz/Projects/trex-helpers/data/003-br-468f95065bbe.pcap"
	filename := "/home/mateusz/Projects/trex-helpers/data/003-br-8f3178dc1ac3.pcap"
	packets, err := parsePcap(filename)
	if err != nil {
		fmt.Printf("File %s could not be read", filename)
	} else {
		fmt.Printf("File %s parsed; %d packets read", filename, len(packets))
	}
}
