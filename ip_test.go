package ursa

import (
	"bytes"
	"net/http"
	"testing"
)

func TestClientIPAddress(t *testing.T) {
	rawUrl := "https://example.com"
	b := new(bytes.Buffer)
	r, err := http.NewRequest("GET", rawUrl, b)
	if err != nil {
		t.Fatalf("Error creating request object")
	}
	t.Errorf("Got ip", r.RemoteAddr)
}
