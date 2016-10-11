package flagtypes

import (
	"fmt"
	"net"
	"strings"
)

// lifted from kubernetes/pkg/util/net.go.   same flags vs pflags problem as we had with StringList

// IP adapts net.IP for use as a flag.
type IP net.IP

func (ip IP) String() string {
	return net.IP(ip).String()
}

func (ip *IP) Set(value string) error {
	*ip = IP(net.ParseIP(strings.TrimSpace(value)))
	if *ip == nil {
		return fmt.Errorf("invalid IP address: '%s'", value)
	}
	return nil
}

// Type returns a string representation of what kind of argument this is
func (ip *IP) Type() string {
	return "cmd.flagtypes.IP"
}

// IPNet adapts net.IPNet for use as a flag.
type IPNet net.IPNet

func DefaultIPNet(value string) IPNet {
	ret := IPNet{}
	if err := ret.Set(value); err != nil {
		panic(err)
	}
	return ret
}

func (ipnet IPNet) String() string {
	n := net.IPNet(ipnet)
	return n.String()
}

func (ipnet *IPNet) Set(value string) error {
	_, n, err := net.ParseCIDR(strings.TrimSpace(value))
	if err != nil {
		return err
	}
	*ipnet = IPNet(*n)
	return nil
}

// Type returns a string representation of what kind of argument this is
func (ipnet *IPNet) Type() string {
	return "cmd.flagtypes.IPNet"
}
