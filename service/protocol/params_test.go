package protocol_test

import (
	"reflect"
	"testing"

	"github.com/fsuhrau/automationhub/service/protocol"
)

func TestExtractParams(t *testing.T) {
	proxy := "http://10.35.111.217:8888"
	cases := map[string]map[string]string{
		// the run dialog
		"MarketId=xu;Proxy=" + proxy: {"MarketId": "xu", "Proxy": proxy},
		// a stored run sent back by "Run again" (runs 160 and 161)
		"MarketId=xu\nProxy=" + proxy:   {"MarketId": "xu", "Proxy": proxy},
		"MarketId=xu\r\nProxy=" + proxy: {"MarketId": "xu", "Proxy": proxy},
		// values keep their "=", empty and keyless entries are skipped
		"Url=https://x.test/?a=1&b=2;;=x;Flag": {"Url": "https://x.test/?a=1&b=2"},
		" Key = value ":                        {"Key": "value"},
		"Empty=":                               {"Empty": ""},
		"":                                     {},
	}
	for in, want := range cases {
		if got := protocol.ExtractParams(in); !reflect.DeepEqual(got, want) {
			t.Errorf("ExtractParams(%q) = %v, want %v", in, got, want)
		}
	}
}
