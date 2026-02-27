// Copyright (c) 2025 Justin Cranford
//
//

package network

import (
	"fmt"
	"net"
)

// ParseIPAddresses parses a slice of IP address strings into net.IP objects.
func ParseIPAddresses(ipAddresses []string) ([]net.IP, error) {
	parsedIPs := make([]net.IP, 0, len(ipAddresses))

	for _, ip := range ipAddresses {
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			return nil, fmt.Errorf("failed to parse IP address: %s", ip)
		}

		parsedIPs = append(parsedIPs, parsedIP)
	}

	return parsedIPs, nil
}

// NormalizeIPv4Addresses converts IPv4-mapped IPv6 addresses to IPv4.
func NormalizeIPv4Addresses(ips []net.IP) []net.IP {
	normalizedIPv4Addresses := make([]net.IP, len(ips))

	for i, ip := range ips {
		normalizedIPv4Address := ip.To4() // Attempt to convert IPv4-mapped IPv6 address to IPv4
		if normalizedIPv4Address == nil {
			normalizedIPv4Addresses[i] = ip // not an IPv4-mapped IPv6, keep original IPv6
		} else {
			normalizedIPv4Addresses[i] = normalizedIPv4Address
		}
	}

	return normalizedIPv4Addresses
}
