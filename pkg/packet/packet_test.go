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
	tests := []struct {
		name      string
		args      args
		wantStats []Packet
		wantErr   bool
	}{
		{"Return error for empty file", args{"", false}, nil, true},

		{"Single latency packet", args{"../../tests/001-one-latency-pkt.pcap", false},
			[]Packet{
				LatencyPacket{time.Date(2019, 12, 05, 15, 17, 48, 959001355, time.Local).UTC(),
					243065},
			}, false},

		{"Mixed twenty packets", args{"../../tests/002-mixed-twenty-ptp-latency-other.pcap", false},
			[]Packet{
				OtherPacket{time.Date(2019, 12, 9, 9, 4, 7, 121075768, time.Local).UTC()},
				OtherPacket{time.Date(2019, 12, 9, 9, 4, 7, 121109985, time.Local).UTC()},
				PTPPacket{time.Date(2019, 12, 9, 9, 4, 7, 121610958, time.Local).UTC(), 0x00},
				OtherPacket{time.Date(2019, 12, 9, 9, 4, 7, 122082686, time.Local).UTC()},
				OtherPacket{time.Date(2019, 12, 9, 9, 4, 7, 122125297, time.Local).UTC()},
				PTPPacket{time.Date(2019, 12, 9, 9, 4, 7, 122583224, time.Local).UTC(), 0x08},
				OtherPacket{time.Date(2019, 12, 9, 9, 4, 7, 123075232, time.Local).UTC()},
				OtherPacket{time.Date(2019, 12, 9, 9, 4, 7, 123152834, time.Local).UTC()},
				PTPPacket{time.Date(2019, 12, 9, 9, 4, 7, 123193888, time.Local).UTC(), 0x01},
				PTPPacket{time.Date(2019, 12, 9, 9, 4, 7, 123569061, time.Local).UTC(), 0x09},
				OtherPacket{time.Date(2019, 12, 9, 9, 4, 7, 124150750, time.Local).UTC()},
				LatencyPacket{time.Date(2019, 12, 9, 9, 4, 7, 213015280, time.Local).UTC(),
					-10509},
				OtherPacket{time.Date(2019, 12, 9, 9, 4, 7, 213098093, time.Local).UTC()},
				OtherPacket{time.Date(2019, 12, 9, 9, 4, 7, 213133027, time.Local).UTC()},
				OtherPacket{time.Date(2019, 12, 9, 9, 4, 7, 213987195, time.Local).UTC()},
				LatencyPacket{time.Date(2019, 12, 9, 9, 4, 7, 262985012, time.Local).UTC(),
					-15250},
			}, false},
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
