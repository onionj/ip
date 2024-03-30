package geolocation

import (
	"net"
	"testing"
)

func TestQueryIPV4(t *testing.T) {

	want := "us"
	ipRangeMap := make(map[string][]*net.IPNet)
	ipRange := "1.1.1.1/24"
	_, ipNet, _ := net.ParseCIDR(ipRange)

	ipRangeMap[want] = append(ipRangeMap[want], ipNet)

	ipGeolocation := IPGeolocation{
		CIDRListV4: ipRangeMap,
		Ready:      true,
	}

	testCase := net.IPv4(byte(1), byte(1), byte(1), byte(1))

	got, err := ipGeolocation.Query(testCase)
	if got != want {
		t.Errorf("got %q, wanted %q, error: %q", got, want, err.Error())
	}
}
