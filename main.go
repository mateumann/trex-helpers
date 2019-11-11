package main

import "fmt"

import "github.com/google/gopacket"
import "github.com/google/gopacket/pcap"

func parsePcap(filename string) (count uint32, err error) {
    if handle, err := pcap.OpenOffline(filename); err != nil {
        return 0, err
    } else {
        fmt.Printf("Opening file with %s timestamp resolution.\n", handle.Resolution())
        packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
        count = 0
        for packet := range packetSource.Packets() {
            handlePacket(packet) // Do something with a packet here.
            count++
        }
    }
    return count, nil
}

func handlePacket(packet gopacket.Packet) {
    //packet.
}

func main() {
    filename := "/home/mateusz/Projects/trex-helpers/data/001-br-15397ec68cd0.pcap"
    //filename := "/home/mateusz/Projects/trex-helpers/data/001-br-51d85eb72d54.pcap"
    //filename := "/home/mateusz/Projects/trex-helpers/data/001-enp3s0.pcap"
    fmt.Println("Hello world")
    packets, err := parsePcap(filename)
    if err != nil {
        fmt.Printf("File %s could not be read", filename)
    } else {
        fmt.Printf("File %s parsed; %d packets read", filename, packets)
    }
}
