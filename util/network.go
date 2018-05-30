package util

import (
	"net"
)

// GetLocalIP returns local node's IP address
func GetLocalIP() string {
	addrSlice, err := net.InterfaceAddrs()
	if nil != err {
		return ""
	}

	for _, addr := range addrSlice {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if nil != ipnet.IP.To4() {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
