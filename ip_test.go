package ursa

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
)

func TestClientIPAddress(t *testing.T) {
	// Setup
	rawUrl := "https://example.com"
	type test struct {
		ips []string
	}
	// Test cases for when X-Forwarded-For header is present
	tests := []test{
		{ips: []string{"192.168.1.1", "10.24.3.54"}},
		{ips: []string{"192.168.1.1"}},
		{ips: []string{"0.0.12.13:3000"}},
		{ips: []string{"0.0.12.13:3000", "34.39.34.34"}},
		{ips: []string{"0.0.12.13:3000", "34.39.34.34", "10.33.53.34"}},
	}
	for _, test := range tests {
		expectedIp := strings.Split(test.ips[0], ":")[0]
		b := new(bytes.Buffer)
		r, err := http.NewRequest("GET", rawUrl, b)
		// Add X-Forwarded-For requests
		for _, ip := range test.ips {
			r.Header.Add("X-Forwarded-For", ip)
		}
		if err != nil {
			t.Fatalf("Error creating request object for test case %v", test)
		}
		gotIp, _ := clientIpAddr(r)
		if gotIp != expectedIp {
			t.Errorf("Expected ip %v got %v", expectedIp, gotIp)
		}
	}
	// TODO
	// Test case for when no X-Forwarded-For header is present
}
