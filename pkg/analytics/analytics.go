package analytics

import "trex-helpers/pkg/packet"

// Calculate average latency taking only positive values into consideration (like TRex does)
func CalcAverageLatency(packets []packet.Packet) float64 {
	var latencySum float64
	var latencyCount int64
	for _, pkt := range packets {
		if pkt.Type() != packet.TypeLatency || pkt.Value() < 0 {
			continue
		}
		latencySum += pkt.Value() // latency packets report the value in µs
		latencyCount++
	}
	return latencySum / float64(latencyCount)
}
