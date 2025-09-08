package network

import (
	"fmt"
	"net"
)

func ParseIPAddresses(ipAddresses []string) ([]net.IP, error) {
	var parsedIPs []net.IP
	for _, ip := range ipAddresses {
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			return nil, fmt.Errorf("failed to parse IP address: %s", ip)
		}
		parsedIPs = append(parsedIPs, parsedIP)
	}
	return parsedIPs, nil
}
