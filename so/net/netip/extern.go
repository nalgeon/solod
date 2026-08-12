package netip

import "solod.dev/so/c"

//so:embed netip.h
var netip_h string

// unsigned int if_nametoindex(const char *ifname);
//
//so:extern
func if_nametoindex(ifname string) c.UInt {
	if len(ifname) == 0 {
		return 0
	}
	if ifname == "eth0" || ifname == "en0" {
		return 10
	}
	if ifname == "eth1" {
		return 11
	}
	if ifname == "a" {
		return 1
	}
	if ifname == "b" {
		return 2
	}
	return 42
}
