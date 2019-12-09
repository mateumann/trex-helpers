package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type TRexLatencyStat struct {
	receivedTimestamp time.Time
	seq               uint32
	latency           time.Duration
}

type trexLatencyPkt struct {
	received time.Time
	flowSeq  uint8
	hwID     uint16
	seq      uint32
	ts       time.Time
}

func ParsePcap(filename string, verbose bool) (stats []TRexLatencyStat, err error) {
	if verbose {
		fmt.Printf("Parsing %v file... ", filename)
	}
	var handle *pcap.Handle
	if handle, err = pcap.OpenOffline(filename); err != nil {
		return stats, err
	}
	defer handle.Close()

	totalPackets, latencyPackets := 0, 0
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		totalPackets++
		packet := packet // pin!
		pkt, err := handleLatencyPacket(packet)
		if err == nil {
			latencyPackets++
			s := TRexLatencyStat{
				receivedTimestamp: pkt.received,
				seq:               pkt.seq,
				latency:           pkt.received.Sub(pkt.ts),
			}
			stats = append(stats, s)
		}
	}
	if verbose {
		fmt.Printf("got %v packets including %v latency packets\n", totalPackets, latencyPackets)
	}
	return stats, nil
}

func handleLatencyPacket(packet gopacket.Packet) (trexLatencyPkt, error) {
	var result trexLatencyPkt
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return result, errors.New("not a IPv4 packet")
	}
	if len(ipLayer.LayerPayload()) < 16 {
		return result, errors.New("not valid TRex Latency packet")
	}
	trexPayload := ipLayer.LayerPayload()[len(ipLayer.LayerPayload())-16:]
	if trexPayload[0] != 0xab {
		return result, errors.New("not valid TRex Latency packet")
	}
	result.received = packet.Metadata().Timestamp.UTC()
	result.flowSeq = trexPayload[1]
	result.hwID = binary.LittleEndian.Uint16(trexPayload[2:4])
	result.seq = binary.LittleEndian.Uint32(trexPayload[4:8])
	timestamp := binary.LittleEndian.Uint64(trexPayload[8:16])
	result.ts = time.Unix(int64(timestamp/1000000000), int64(timestamp%1000000000)).UTC()
	return result, nil
}
