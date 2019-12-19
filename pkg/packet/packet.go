package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Type uint8

const (
	TypeLatency Type = 0xab
	TypePTP     Type = 0xf7
	TypeOther   Type = 0xff
)

const InvalidLatency time.Duration = -1

type Packet interface {
	ReceivedAt() time.Time
	Type() Type
	Value() float64
}

type LatencyPacket struct {
	received time.Time
	latency  time.Duration
}

func (pkt LatencyPacket) ReceivedAt() time.Time {
	return pkt.received
}

func (pkt LatencyPacket) Type() Type {
	return TypeLatency
}

func (pkt LatencyPacket) Value() float64 {
	// return latency in µs
	return float64(pkt.latency) / 1000
}

type PTPPacket struct {
	received  time.Time
	messageID uint8
}

func (pkt PTPPacket) ReceivedAt() time.Time {
	return pkt.received
}

func (pkt PTPPacket) Type() Type {
	return TypePTP
}

func (pkt PTPPacket) Value() float64 {
	// kind of 40 µs high peak
	return float64(0x40 + pkt.messageID)
}

type OtherPacket struct {
	received time.Time
}

func (pkt OtherPacket) ReceivedAt() time.Time {
	return pkt.received
}

func (pkt OtherPacket) Type() Type {
	return TypeOther
}

func (pkt OtherPacket) Value() float64 {
	// kind of 5 µs high peak
	return 5.0
}

func ParsePcap(filename string, verbose bool) (packets []Packet, err error) {
	if verbose {
		fmt.Printf("Parsing %v file... ", filename)
	}
	var handle *pcap.Handle
	if handle, err = pcap.OpenOffline(filename); err != nil {
		return packets, err
	}
	defer handle.Close()

	totalPackets, latencyPackets, invalidLatencyPackets, ptpPackets := 0, 0, 0, 0
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		totalPackets++
		packet := packet // pin!
		received := packet.Metadata().Timestamp.UTC()
		latencyPkt, err := handleLatencyPacket(packet, received)
		if err == nil {
			latencyPackets++
			packets = append(packets, latencyPkt)
			if latencyPkt.latency == InvalidLatency {
				invalidLatencyPackets++
			}
			continue
		}
		ptpPkt, err := handlePTPPacket(packet, received)
		if err == nil {
			ptpPackets++
			packets = append(packets, ptpPkt)
			continue
		}
		ptpInIPPkt, err := handlePTPinIPPacket(packet, received)
		if err == nil {
			ptpPackets++
			packets = append(packets, ptpInIPPkt)
			continue
		}
		packets = append(packets, OtherPacket{received})
	}
	if verbose {
		fmt.Printf("got %v packets including %v latency and %v PTP packets; %v latency packets are somewhat invalid\n",
			totalPackets, latencyPackets, ptpPackets, invalidLatencyPackets)
	}
	return packets, nil
}

func handleLatencyPacket(packet gopacket.Packet, received time.Time) (LatencyPacket, error) {
	var result LatencyPacket
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return result, errors.New("not a IPv4 packet")
	}
	if len(ipLayer.LayerPayload()) < 16 {
		return result, errors.New("not valid TRex Latency packet")
	}
	latencyData := ipLayer.LayerPayload()[len(ipLayer.LayerPayload())-16:]
	if latencyData[0] != 0xab {
		return result, errors.New("not valid TRex Latency packet")
	}

	result.received = received
	timestamp := binary.LittleEndian.Uint64(latencyData[8:16])
	sent := time.Unix(int64(timestamp/1000000000), int64(timestamp%1000000000)).UTC()
	result.latency = received.Sub(sent)
	if math.Abs(float64(result.latency)) > float64(time.Second)*10 {
		result.latency = InvalidLatency
	}
	return result, nil
}

func handlePTPPacket(packet gopacket.Packet, received time.Time) (PTPPacket, error) {
	var result PTPPacket
	ethLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		return result, errors.New("not an Ethernet packet")
	}

	return parsePTPPayload(ethLayer.LayerPayload(), received)
}

func handlePTPinIPPacket(packet gopacket.Packet, received time.Time) (PTPPacket, error) {
	var result PTPPacket
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return result, errors.New("not a IPv4 packet")
	}
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return result, errors.New("not a UDP packet")
	}

	return parsePTPPayload(udpLayer.LayerPayload(), received)
}

func parsePTPPayload(data []byte, received time.Time) (result PTPPacket, err error) {
	if ((data[0] == 0x00 || data[0] == 0x01 || data[0] == 0x08) && len(data) == 44) ||
		(data[0] == 0x09 && len(data) == 54) {
		result.received = received
		result.messageID = data[0]
		return result, nil
	}

	return result, errors.New("not a PTP packet")
}
