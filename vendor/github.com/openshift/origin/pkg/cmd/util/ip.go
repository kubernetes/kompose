package util

import (
	"errors"
	"net"
)

// ErrorNoDefaultIP is returned when no suitable non-loopback address can be found.
var ErrorNoDefaultIP = errors.New("no suitable IP address")

// DefaultLocalIP4 returns an IPv4 address that this host can be reached
// on. Will return NoDefaultIP if no suitable address can be found.
func DefaultLocalIP4() (net.IP, error) {
	devices, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, dev := range devices {
		if (dev.Flags&net.FlagUp != 0) && (dev.Flags&net.FlagLoopback == 0) {
			addrs, err := dev.Addrs()
			if err != nil {
				continue
			}
			for i := range addrs {
				if ip, ok := addrs[i].(*net.IPNet); ok {
					if ip.IP.To4() != nil {
						return ip.IP, nil
					}
				}
			}
		}
	}
	return nil, ErrorNoDefaultIP
}

// AllLocalIP4 returns all the IPv4 addresses that this host can be reached
// on.
func AllLocalIP4() ([]net.IP, error) {
	devices, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	ret := []net.IP{}
	for _, dev := range devices {
		if dev.Flags&net.FlagUp != 0 {
			addrs, err := dev.Addrs()
			if err != nil {
				continue
			}
			for i := range addrs {
				if ip, ok := addrs[i].(*net.IPNet); ok {
					if ip.IP.To4() != nil {
						ret = append(ret, ip.IP)
					}
				}
			}
		}
	}
	return ret, nil
}
