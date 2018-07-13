package net

import (
	"net"
)

// TODO<GEB>: implement me

type LeaseManager struct {
	// the CIDR to draw leases from
	CIDR *net.IPNet
	// array of leases currently issued. each element is an integer offset from the zero address for the CIDR.
	leases []uint32
	// array of leases that should not be granted for this manager.
	blacklist []uint32
}

func NewLeaseManager() *LeaseManager {
	return nil
}

func (l *LeaseManager) addBlacklistedIP(ip string) error {
	return nil
}

func (l *LeaseManager) removeBlacklistedIP(ip string) error {
	return nil
}

func (l *LeaseManager) getUnusedIP() (string, error) {
	return "", nil
}

func (l *LeaseManager) releaseIP(ip string) error {
	return nil
}
