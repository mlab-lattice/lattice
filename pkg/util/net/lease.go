package net

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"sort"

	"github.com/golang/glog"
)

type netOffset uint32

func (n netOffset) IP(network *net.IPNet) (string, error) {
	ones, bits := network.Mask.Size()
	if ones == 0 && bits == 0 {
		return "", fmt.Errorf("non-canonical netmasks are not supported, got %v", network)
	}
	if bits != 32 {
		return "", fmt.Errorf("invalid netmask %v for netOffset conversion", network)
	}
	if n >= netOffset(math.Pow(2, float64(bits-ones))) {
		return "", fmt.Errorf("netmask %v does not contain offest %v", n, network)
	}
	netIP := network.IP.To4()
	hostBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(hostBytes, uint32(n))
	for i := 0; i < len(hostBytes); i++ {
		hostBytes[i] |= byte(netIP[i])
	}
	return net.IPv4(hostBytes[0], hostBytes[1], hostBytes[2], hostBytes[3]).String(), nil
}

type netOffsets []netOffset

func (ns netOffsets) Len() int {
	return len(ns)
}

func (ns netOffsets) Less(i, j int) bool {
	return ns[i] < ns[j]
}

func (ns netOffsets) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}

func (ns netOffsets) Sort() {
	sort.Sort(ns)
}

func (ns netOffsets) IsSorted() bool {
	return sort.IsSorted(ns)
}

func (ns netOffsets) Contains(no netOffset) (bool, int) {
	i := sort.Search(len(ns), func(i int) bool {
		return ns[i] >= no
	})
	return i < len(ns) && ns[i] == no, i
}

func (ns netOffsets) Add(no netOffset) (netOffsets, bool) {
	present, i := ns.Contains(no)
	if present {
		return ns, false
	}
	_ns := append(ns, 0)
	copy(_ns[i+1:], _ns[i:])
	_ns[i] = no
	return _ns, true
}

func (ns netOffsets) Remove(no netOffset) (netOffsets, bool) {
	present, i := ns.Contains(no)
	if !present {
		return ns, false
	}
	copy(ns[i:], ns[i+1:])
	return ns[:len(ns)-1], true
}

func (ns netOffsets) Pop() (netOffset, netOffsets, bool) {
	if len(ns) == 0 {
		return 0, ns, false
	}
	return ns[len(ns)-1], ns[:len(ns)-1], true
}

type LeaseManager struct {
	// the network to draw leases from
	net *net.IPNet
	// array of leases currently issued. each element is an integer offset from the zero address for the network
	leases netOffsets
	// array of expired leases that can be reused. each element is an integer offset from the zero address for the network
	freelist netOffsets
	// array of leases that should not be granted for this manager
	blacklist netOffsets
}

func NewLeaseManager(cidr string) (*LeaseManager, error) {
	_, net, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	return &LeaseManager{
		net:       net,
		leases:    make(netOffsets, 0, 256),
		freelist:  make(netOffsets, 0, 256),
		blacklist: make(netOffsets, 0, 256),
	}, nil
}

func NewLeaseManagerWithState(cidr string, leases, blacklist []string) (*LeaseManager, error) {
	l, err := NewLeaseManager(cidr)
	if err != nil {
		return nil, err
	}
	err = l.Blacklist(blacklist...)
	if err != nil {
		return nil, err
	}
	_, err = l.Lease(leases...)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (l *LeaseManager) NetContains(ip string) bool {
	_ip := net.ParseIP(ip).To4()
	return l.net.Contains(_ip)
}

func (l *LeaseManager) GetNetIP() string {
	return l.net.IP.String()
}

func (l *LeaseManager) getNetOffset(ip string) (netOffset, error) {
	if !l.NetContains(ip) {
		return 0, fmt.Errorf("invalid IP %s for network %v", ip, l.net)
	}

	netBytes := []byte(l.net.IP.To4())
	hostBytes := []byte(net.ParseIP(ip).To4())

	for i := 0; i < len(hostBytes); i++ {
		hostBytes[i] ^= netBytes[i]
	}

	return netOffset(binary.BigEndian.Uint32(hostBytes)), nil
}

func (l *LeaseManager) Blacklist(ips ...string) error {
	var ok bool

	nos := make(netOffsets, 0, len(ips))

	for _, ip := range ips {
		no, err := l.getNetOffset(ip)
		if err != nil {
			return err
		}
		if present, _ := l.blacklist.Contains(no); present {
			return fmt.Errorf("cannot blacklist existing blacklisted IP %s", ip)
		}
		if present, _ := l.leases.Contains(no); present {
			return fmt.Errorf("cannot blacklist currently leased IP %s", ip)
		}
		if nos, ok = nos.Add(no); !ok {
			return fmt.Errorf("cannot blacklist duplicate IP %s", ip)
		}
	}

	for i, no := range nos {
		if l.freelist, ok = l.freelist.Remove(no); ok {
			glog.Warningf("blacklisted IP %s while on free list", ips[i])
		}
		if l.blacklist, ok = l.blacklist.Add(no); !ok {
			// unreachable
			panic(fmt.Sprintf("cannot blacklist existing blacklisted offset %v", no))
		}
	}
	return nil
}

func (l *LeaseManager) RemoveBlacklisted(ips ...string) error {
	var ok bool

	nos := make(netOffsets, 0, len(ips))

	for _, ip := range ips {
		no, err := l.getNetOffset(ip)
		if err != nil {
			return err
		}
		if present, _ := l.blacklist.Contains(no); !present {
			return fmt.Errorf("cannot remove nonexistent blacklisted IP %s", ip)
		}
		if nos, ok = nos.Add(no); !ok {
			return fmt.Errorf("cannot remove duplicate blacklisted IP %s", ip)
		}
	}

	for _, no := range nos {
		if l.blacklist, ok = l.blacklist.Remove(no); !ok {
			// unreachable
			panic(fmt.Sprintf("cannot remove nonexistent blacklisted offset %v", no))
		}
		if l.freelist, ok = l.freelist.Add(no); !ok {
			// unreachable
			panic(fmt.Sprintf("detected double free for blacklisted offset %v", no))
		}
	}

	return nil
}

func (l *LeaseManager) newLease() (string, netOffset, error) {
	var no netOffset = 0
	if len(l.leases) > 0 {
		no = l.leases[len(l.leases)-1] + 1
	}
	for {
		if present, _ := l.blacklist.Contains(no); !present {
			break
		}
		no += 1
	}
	ip, err := no.IP(l.net)
	if err != nil {
		return "", 0, err
	}
	return ip, no, nil
}

func (l *LeaseManager) Lease(ips ...string) ([]string, error) {
	var ip string
	var ok bool
	var err error
	var no netOffset
	var nos netOffsets

	fixFreelist := true

	if len(ips) == 0 {
		no, l.freelist, ok = l.freelist.Pop()
		if ok {
			ip, err = no.IP(l.net)
			if err != nil {
				// unreachable
				panic(err.Error())
			}
		} else {
			ip, no, err = l.newLease()
			if err != nil {
				return nil, err
			}
		}
		ips = append(ips, ip)
		fixFreelist = false
	}

	nos = make(netOffsets, 0, len(ips))

	for _, ip := range ips {
		no, err = l.getNetOffset(ip)
		if err != nil {
			return nil, err
		}
		if present, _ := l.leases.Contains(no); present {
			return nil, fmt.Errorf("cannot lease currently leased IP %s", ip)
		}
		if present, _ := l.blacklist.Contains(no); present {
			return nil, fmt.Errorf("cannot lease currently blacklisted IP %s", ip)
		}
		nos, ok = nos.Add(no)
		if !ok {
			return nil, fmt.Errorf("cannot lease duplicate IP %s", ip)
		}
	}

	for i, no := range nos {
		l.leases, ok = l.leases.Add(no)
		l.freelist, _ = l.freelist.Remove(no)
		if !ok {
			// unreachable
			panic(fmt.Sprintf("cannot lease currently leased IP %s", ips[i]))
		}
	}

	if fixFreelist {
		freeNO := netOffset(0)
		for _, no := range l.leases {
			for freeNO < no {
				if present, _ := l.blacklist.Contains(freeNO); !present {
					// we may try to re-add free list entries here, so ignore "ok" return val
					l.freelist, _ = l.freelist.Add(freeNO)
				}
				freeNO++
			}
			freeNO = no + 1
		}
	}

	return ips, nil
}

func (l *LeaseManager) RemoveLeased(ips ...string) error {
	var ok bool

	nos := make(netOffsets, 0, len(ips))

	for _, ip := range ips {
		no, err := l.getNetOffset(ip)
		if err != nil {
			return err
		}
		if present, _ := l.leases.Contains(no); !present {
			return fmt.Errorf("cannot remove nonexistent leased IP %s", ip)
		}
		nos, ok = nos.Add(no)
		if !ok {
			return fmt.Errorf("cannot remove duplicate leased IP %s", ip)
		}
	}
	for _, no := range nos {
		if l.leases, ok = l.leases.Remove(no); !ok {
			// unreachable
			panic(fmt.Sprintf("cannot remove nonexistent leased offset %v", no))
		}
		if l.freelist, ok = l.freelist.Add(no); !ok {
			// unreachable
			panic(fmt.Sprintf("detected double free for leased offset %v", no))
		}
	}
	return nil
}
