package analytics

import (
	"time"
	"trex-helpers/pkg/packet"
)

type PeriodicAvgLatency struct {
	StartTimestamp time.Time
	EndTimestamp   time.Time
	Value          float64
}

// Calculate average latency taking only positive values into consideration (like TRex does)
func CalcPositiveAverageLatency(packets []packet.Packet) float64 {
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

// Calculate average latency between time synchronizations
func CalcPeriodicAverageLatency(packets []packet.Packet) (latencies []PeriodicAvgLatency) {
	var latencySum float64
	var latencyCount int64
	var periodStart, periodEnd time.Time
	for n, pkt := range packets {
		if (pkt.Type() == packet.TypePTP && pkt.Value() == 0x49) || // use DELAY_RESPONSE PTP message as a marker
			(n == len(packets)-1) {
			periodEnd = pkt.ReceivedAt()
			if latencyCount > 0 {
				latencies = append(latencies, PeriodicAvgLatency{
					StartTimestamp: periodStart,
					EndTimestamp:   periodEnd,
					Value:          latencySum / float64(latencyCount),
				})
			}
			pkt.ReceivedAt()
			latencyCount = 0
			latencySum = 0.0
			periodStart = pkt.ReceivedAt()
			continue
		}
		if pkt.Type() != packet.TypeLatency {
			continue
		}
		latencySum += pkt.Value() // latency packets report the value in µs
		latencyCount++
	}
	return
}
