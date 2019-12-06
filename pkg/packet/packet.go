package packet

import (
	"errors"
	"time"

	"github.com/google/gopacket"

	"github.com/google/gopacket/pcap"
)

type TRexLatencyStat struct {
	sentTimestamp     time.Time
	receivedTimestamp time.Time
	seq               uint32
	latencyNs         int64
}

type trexLatencyPkt struct {
	magic     uint8
	flowSeq   uint8
	hwId      uint16
	seq       uint32
	timestamp uint64
}

func ParsePcap(filename string, verbose bool) (stats []TRexLatencyStat, err error) {
	var handle *pcap.Handle
	if handle, err = pcap.OpenOffline(filename); err != nil {
		return stats, err
	}
	defer handle.Close()

	totalPackets, erroneousPackets := 0, 0
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		totalPackets++
		packet := packet // pin!
		pkt, err = handlePacket(packet)

		if err != nil {
			erroneousPackets++
			erroneousPackets++
		}
	}
	return nil, errors.New("not yet implemented")
}

func handlePacket(packet gopacket.Packet) (trexLatency trexLatencyPkt, err error) {
	return trexLatency, errors.New("not yet implemented")
}
