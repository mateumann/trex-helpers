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
		wantStats []TRexLatencyStat
		wantErr   bool
	}{
		{"Return error for empty file", args{"", false}, nil, true},

		{"Single latency packet", args{"../../tests/001-one-latency-pkt.pcap", false},
			[]TRexLatencyStat{
				{time.Date(2019, 12, 05, 15, 17, 48, 959001355, time.Local).UTC(),
					50, 243065},
			}, false},

		{"Mixed twenty packets", args{"../../tests/002-mixed-twenty-ptp-latency-other.pcap", false},
			[]TRexLatencyStat{
				{time.Date(2019, 12, 9, 9, 4, 7, 213015280, time.Local).UTC(),
					0x50, -10509},
				{time.Date(2019, 12, 9, 9, 4, 7, 262985012, time.Local).UTC(),
					0x51, -15250},
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
