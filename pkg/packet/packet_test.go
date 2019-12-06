package packet

import (
	"reflect"
	"testing"
	"time"
)

func TestParsePcap(t *testing.T) {
	type args struct {
		filename string
		verbose  bool
	}
	ts := TRexLatencyStat{time.Date(2019, 12, 05, 14, 17, 48, 959001355, time.UTC),
		time.Date(2019, 12, 05, 14, 17, 48, 958758290, time.UTC),
		0, 243065}

	tests := []struct {
		name      string
		args      args
		wantStats []TRexLatencyStat
		wantErr   bool
	}{
		{"Return error for empty file", args{"", false}, nil, true},
		{"Single latency packet", args{"../../tests/001-one-latency-pkt.pcap", false}, []TRexLatencyStat{ts}, false},
	}

	for _, tt := range tests {
		tt := tt // pin!
		t.Run(tt.name, func(t *testing.T) {
			gotStats, err := ParsePcap(tt.args.filename, tt.args.verbose)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePcap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotStats, tt.wantStats) {
				t.Errorf("ParsePcap() gotStats = %v, want %v", gotStats, tt.wantStats)
			}
		})
	}
}

//func Test_handlePacket(t *testing.T) {
//	type args struct {
//		packet gopacket.Packet
//	}
//	tests := []struct {
//		name            string
//		args            args
//		wantTrexLatency trexLatencyPkt
//		wantErr         bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			gotTrexLatency, err := handlePacket(tt.args.packet)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("handlePacket() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(gotTrexLatency, tt.wantTrexLatency) {
//				t.Errorf("handlePacket() gotTrexLatency = %v, want %v", gotTrexLatency, tt.wantTrexLatency)
//			}
//		})
//	}
//}
