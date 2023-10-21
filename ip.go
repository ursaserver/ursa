package ursa

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

var errorIP = errors.New("invalid IP")

// Get the IP address of the downstream client from the request
// Note that in this function, we assume that the values provided
// in the Header fields are safe and any spoofing attempts have been
// taken care of.
// See https://github.com/ursaserver/ursa/issues/4  for details.
func clientIpAddr(r *http.Request) (string, error) {
	// By HTTP standards, the value of X-Forwarded-For is a list of comma+space
	// separated IP addresses (ip:port or ip). Where the leftmost is the
	// address of the the client, then first proxy, second proxy, so on
	f := r.Header.Get("X-Forwarded-For")
	if f != "" {
		// Here client means the first client. The initiator of the request, not proxy.
		clientIP := strings.Split(f, ", ")[0]
		return strings.Split(clientIP, ":")[0], nil // Split ip from ip:port format
	}
	// If no proxies between upstream and downstream, we read IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", errorIP
	}
	return ip, nil
}
