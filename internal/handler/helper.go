package handler

import (
	"net"
	"net/http"
)

func isIPTrusted(r *http.Request, subnet string) bool {
	if subnet == "" {
		return false
	}

	ipStr := r.Header.Get("X-Real-IP")
	if ipStr == "" {
		return false
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false
	}

	return ipNet.Contains(ip)
}
