package net

import (
	"net"
	"reflect"
	"testing"
	"unsafe"

	// "github.com/golang/glog"

	"github.com/stretchr/testify/require"
)

func TestNetOffset(t *testing.T) {
	nonCanonicalNet := &net.IPNet{
		IP:   net.IPv4(0x0a, 0x00, 0x00, 0x00),
		Mask: net.IPv4Mask(0xff, 0x00, 0xff, 0x00),
	}
	canonicalNet := &net.IPNet{
		IP:   net.IPv4(0x0a, 0x0a, 0x00, 0x00),
		Mask: net.IPv4Mask(0xff, 0xff, 0x00, 0x00),
	}
	canonicalNetOddBits := &net.IPNet{
		IP:   net.IPv4(0x0a, 0x0a, 0x00, 0x00),
		Mask: net.IPv4Mask(0xff, 0xff, 0x80, 0x00),
	}

	t.Run("nonCanonicalNet", func(t *testing.T) {
		var no netOffset = 0x01
		ip, err := no.IP(nonCanonicalNet)
		require.Empty(t, ip)
		require.NotNil(t, err)
		require.Regexp(t, `^non-canonical netmasks are not supported, got`, err.Error())
	})

	t.Run("canonicalNet offset included", func(t *testing.T) {
		var no netOffset = 0x01
		ip, err := no.IP(canonicalNet)
		require.NotEmpty(t, ip)
		require.Nil(t, err)
		require.Equal(t, "10.10.0.1", ip)
	})

	t.Run("canonicalNet offset included boundary", func(t *testing.T) {
		var no netOffset = 0xffff
		ip, err := no.IP(canonicalNet)
		require.NotEmpty(t, ip)
		require.Nil(t, err)
		require.Equal(t, "10.10.255.255", ip)
	})

	t.Run("canonicalNet offset not included boundary", func(t *testing.T) {
		var no netOffset = 0x010000
		ip, err := no.IP(canonicalNet)
		require.Empty(t, ip)
		require.NotNil(t, err)
		require.Regexp(t, `^netmask \d+ does not contain offest \d{1,3}(\.\d{1,3}){3}/\d{2}`, err.Error())
	})

	t.Run("canonicalNetOddBits offset included", func(t *testing.T) {
		var no netOffset = 0x01
		ip, err := no.IP(canonicalNetOddBits)
		require.NotEmpty(t, ip)
		require.Nil(t, err)
		require.Equal(t, "10.10.0.1", ip)
	})

	t.Run("canonicalNetOddBits offset included boundary", func(t *testing.T) {
		var no netOffset = 0x7fff
		ip, err := no.IP(canonicalNetOddBits)
		require.NotEmpty(t, ip)
		require.Nil(t, err)
		require.Equal(t, "10.10.127.255", ip)
	})

	t.Run("canonicalNetOddBits offset not included boundary", func(t *testing.T) {
		var no netOffset = 0x8000
		ip, err := no.IP(canonicalNetOddBits)
		require.Empty(t, ip)
		require.NotNil(t, err)
		require.Regexp(t, `^netmask \d+ does not contain offest \d{1,3}(\.\d{1,3}){3}/\d{2}`, err.Error())
	})
}

func TestNetOffsets(t *testing.T) {
	var emptyNetOffsets netOffsets
	var populatedNetOffsets netOffsets
	var unorderedNetOffsets netOffsets

	setup := func() {
		emptyNetOffsets = make(netOffsets, 0)
		populatedNetOffsets = netOffsets{1, 256, 65536}
		unorderedNetOffsets = netOffsets{256, 1, 65536, 2}
	}

	t.Run("empty netOffsets contains", func(t *testing.T) {
		setup()
		present, index := emptyNetOffsets.Contains(1)
		require.False(t, present)
		require.Equal(t, 0, index)
	})

	t.Run("empty netOffsets add", func(t *testing.T) {
		setup()
		emptyNetOffsets, ok := emptyNetOffsets.Add(1)
		require.True(t, ok)
		require.Len(t, emptyNetOffsets, 1)
		emptyNetOffsets, ok = emptyNetOffsets.Add(1)
		require.False(t, ok)
		require.Len(t, emptyNetOffsets, 1)
	})

	t.Run("empty netOffsets remove", func(t *testing.T) {
		setup()
		emptyNetOffsets, ok := emptyNetOffsets.Remove(1)
		require.False(t, ok)
		require.Len(t, emptyNetOffsets, 0)
	})

	t.Run("empty netOffsets pop", func(t *testing.T) {
		setup()
		no, emptyNetOffsets, ok := emptyNetOffsets.Pop()
		require.False(t, ok)
		require.Len(t, emptyNetOffsets, 0)
		require.Equal(t, netOffset(0), no)
	})

	t.Run("populated netOffsets contains", func(t *testing.T) {
		setup()
		present, index := populatedNetOffsets.Contains(1)
		require.True(t, present)
		require.Equal(t, 0, index)
		present, index = populatedNetOffsets.Contains(2)
		require.False(t, present)
		require.Equal(t, 1, index)
	})

	t.Run("populated netOffsets add", func(t *testing.T) {
		setup()
		netOffsetsLength := len(populatedNetOffsets)
		populatedNetOffsets, ok := populatedNetOffsets.Add(1)
		require.False(t, ok)
		require.Len(t, populatedNetOffsets, netOffsetsLength)
		populatedNetOffsets, ok = populatedNetOffsets.Add(2)
		require.True(t, ok)
		require.Len(t, populatedNetOffsets, netOffsetsLength+1)
		require.Equal(t, netOffset(2), populatedNetOffsets[1])
	})

	t.Run("populated netOffsets remove", func(t *testing.T) {
		setup()
		netOffsetsLength := len(populatedNetOffsets)
		populatedNetOffsets, ok := populatedNetOffsets.Remove(2)
		require.False(t, ok)
		require.Len(t, populatedNetOffsets, netOffsetsLength)
		populatedNetOffsets, ok = populatedNetOffsets.Remove(1)
		require.True(t, ok)
		require.Len(t, populatedNetOffsets, netOffsetsLength-1)
	})

	t.Run("populated netOffsets pop", func(t *testing.T) {
		setup()
		netOffsetsLength := len(populatedNetOffsets)
		no, populatedNetOffsets, ok := populatedNetOffsets.Pop()
		require.True(t, ok)
		require.Equal(t, netOffset(65536), no)
		require.Len(t, populatedNetOffsets, netOffsetsLength-1)
	})

	t.Run("unordered netOffsets sort/isSorted", func(t *testing.T) {
		setup()
		require.False(t, unorderedNetOffsets.IsSorted())
		unorderedNetOffsets.Sort()
		require.True(t, unorderedNetOffsets.IsSorted())
		require.Equal(t, netOffsets{1, 2, 256, 65536}, unorderedNetOffsets)
	})
}

func TestLeaseManager(t *testing.T) {
	var l LeaseManager

	setup := func() {
		l, _ = NewLeaseManagerWithState(
			"10.10.0.0/16",
			[]string{
				"10.10.0.1",
				"10.10.0.2",
				"10.10.0.3",
				"10.10.0.4",
				"10.10.0.5",
			},
			[]string{
				"10.10.0.0",
			})
	}

	getMember := func(s interface{}, name string) interface{} {
		// XXX: inspired by https://stackoverflow.com/questions/42664837/access-unexported-fields-in-golang-reflect
		// XXX: the following does not work
		// https://stackoverflow.com/questions/17981651/in-go-is-there-any-way-to-access-private-fields-of-a-struct-from-another-packag/17982725
		// get the underlying concrete value that s points to
		v := reflect.ValueOf(s).Elem()
		// get the field by name (note, we cannot call `Interface` here because it "cannot
		// return value obtained from unexported field or method")
		f := v.FieldByName(name)
		// work around the "unexported field" problem by getting an unsafe pointer to the underlying
		// field and then converting back to a Value before returning the underlying value as an
		// interface
		return reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Interface()
	}

	t.Run("NewLeaseManager", func(t *testing.T) {
		l, err := NewLeaseManager("10.10.0.0/16")
		require.Nil(t, err)
		require.NotNil(t, l)
	})

	t.Run("NewLeaseManager non-CIDR notation", func(t *testing.T) {
		l, err := NewLeaseManager("10.10.0.0")
		require.Nil(t, l)
		require.NotNil(t, err)
		require.Regexp(t, `^invalid CIDR address`, err.Error())
	})

	t.Run("NewLeaseManagerWithState", func(t *testing.T) {
		l, err := NewLeaseManagerWithState(
			"10.10.0.0/16",
			[]string{
				"10.10.0.1",
				"10.10.0.2",
				"10.10.0.9",
				"10.10.0.10",
			},
			[]string{
				"10.10.0.4",
				"10.10.0.7",
			})
		require.Nil(t, err)
		leases, ok := getMember(l, "leases").(netOffsets)
		require.True(t, ok)
		require.Len(t, leases, 4)
		require.True(t, leases.IsSorted())
		require.Equal(t, netOffsets{0x01, 0x02, 0x09, 0x0a}, leases)
		blacklist, ok := getMember(l, "blacklist").(netOffsets)
		require.True(t, ok)
		require.Len(t, blacklist, 2)
		require.True(t, blacklist.IsSorted())
		require.Equal(t, netOffsets{0x04, 0x07}, blacklist)
		freelist, ok := getMember(l, "freelist").(netOffsets)
		require.True(t, ok)
		require.Len(t, freelist, 5)
		require.True(t, freelist.IsSorted())
		require.Equal(t, netOffsets{0x00, 0x03, 0x05, 0x06, 0x08}, freelist)
	})

	t.Run("NewLeaseManagerWithState IP leased and blacklisted", func(t *testing.T) {
		l, err := NewLeaseManagerWithState(
			"10.10.0.0/16",
			[]string{
				"10.10.0.9",
				"10.10.0.10",
			},
			[]string{
				"10.10.0.7",
				"10.10.0.10",
			})
		require.Nil(t, l)
		require.NotNil(t, err)
		require.Regexp(t, "cannot lease currently blacklisted IP 10.10.0.10", err.Error())
	})

	t.Run("NewLeaseManagerWithState duplicate leases", func(t *testing.T) {
		l, err := NewLeaseManagerWithState(
			"10.10.0.0/16",
			[]string{
				"10.10.0.9",
				"10.10.0.9",
			},
			[]string{
				"10.10.0.7",
				"10.10.0.10",
			})
		require.Nil(t, l)
		require.NotNil(t, err)
		require.Regexp(t, "cannot lease duplicate IP 10.10.0.9", err.Error())
	})

	t.Run("NewLeaseManagerWithState duplicate blacklisted IPs", func(t *testing.T) {
		l, err := NewLeaseManagerWithState(
			"10.10.0.0/16",
			[]string{
				"10.10.0.9",
				"10.10.0.10",
			},
			[]string{
				"10.10.0.7",
				"10.10.0.7",
			})
		require.Nil(t, l)
		require.NotNil(t, err)
		require.Equal(t, "cannot blacklist duplicate IP 10.10.0.7", err.Error())
	})

	t.Run("NewLeaseManagerWithState with bad lease IP", func(t *testing.T) {
		l, err := NewLeaseManagerWithState(
			"10.10.0.0/16",
			[]string{
				"192.168.1.1",
				"10.10.0.10",
			},
			[]string{
				"10.10.0.7",
				"10.10.0.8",
			})
		require.Nil(t, l)
		require.NotNil(t, err)
		require.Equal(t, "invalid IP 192.168.1.1 for network 10.10.0.0/16", err.Error())
	})

	t.Run("NewLeaseManagerWithState with bad blacklist IP", func(t *testing.T) {
		l, err := NewLeaseManagerWithState(
			"10.10.0.0/16",
			[]string{
				"10.10.0.9",
				"10.10.0.10",
			},
			[]string{
				"192.168.1.1",
				"10.10.0.8",
			})
		require.Nil(t, l)
		require.NotNil(t, err)
		require.Equal(t, "invalid IP 192.168.1.1 for network 10.10.0.0/16", err.Error())
	})

	t.Run("LeaseManager new unspecified lease", func(t *testing.T) {
		setup()
		leases, err := l.Lease()
		require.Nil(t, err)
		require.NotNil(t, leases)
		require.Len(t, leases, 1)
		require.Equal(t, "10.10.0.6", leases[0])
		lmLeases := getMember(l, "leases")
		require.Len(t, lmLeases, 6)
		require.Equal(t, netOffsets{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, lmLeases)
		freelist := getMember(l, "freelist")
		require.Len(t, freelist, 0)
	})

	t.Run("LeaseManager new specified lease", func(t *testing.T) {
		setup()
		leases, err := l.Lease("10.10.0.10")
		require.Nil(t, err)
		require.NotNil(t, leases)
		require.Len(t, leases, 1)
		require.Equal(t, "10.10.0.10", leases[0])
		lmLeases := getMember(l, "leases")
		require.Len(t, lmLeases, 6)
		require.Equal(t, netOffsets{0x01, 0x02, 0x03, 0x04, 0x05, 0x0a}, lmLeases)
		freelist := getMember(l, "freelist")
		require.Len(t, freelist, 4)
		require.Equal(t, netOffsets{0x06, 0x07, 0x08, 0x09}, freelist)
	})

	t.Run("LeaseManager new specified leases", func(t *testing.T) {
		setup()
		leases, err := l.Lease("10.10.0.7", "10.10.0.9")
		require.Nil(t, err)
		require.NotNil(t, leases)
		require.Len(t, leases, 2)
		require.Equal(t, []string{"10.10.0.7", "10.10.0.9"}, leases)
		lmLeases := getMember(l, "leases")
		require.Len(t, lmLeases, 7)
		require.Equal(t, netOffsets{0x01, 0x02, 0x03, 0x04, 0x05, 0x07, 0x09}, lmLeases)
		freelist := getMember(l, "freelist")
		require.Len(t, freelist, 2)
		require.Equal(t, netOffsets{0x06, 0x08}, freelist)
	})

	t.Run("LeaseManager new specified unordered lease", func(t *testing.T) {
		setup()
		leases, err := l.Lease("10.10.0.6", "10.10.0.1")
		require.Nil(t, leases)
		require.NotNil(t, err)
		require.Equal(t, "cannot lease currently leased IP 10.10.0.1", err.Error())
		lmLeases := getMember(l, "leases")
		require.Equal(t, netOffsets{0x01, 0x02, 0x03, 0x04, 0x05}, lmLeases)
		freelist := getMember(l, "freelist")
		require.Len(t, freelist, 0)
		blacklist := getMember(l, "blacklist")
		require.Equal(t, netOffsets{0x00}, blacklist)
	})

	t.Run("LeaseManager new specified unordered blacklisted lease", func(t *testing.T) {
		setup()
		leases, err := l.Lease("10.10.0.6", "10.10.0.0")
		require.Nil(t, leases)
		require.NotNil(t, err)
		require.Equal(t, "cannot lease currently blacklisted IP 10.10.0.0", err.Error())
		lmLeases := getMember(l, "leases")
		require.Equal(t, netOffsets{0x01, 0x02, 0x03, 0x04, 0x05}, lmLeases)
		freelist := getMember(l, "freelist")
		require.Len(t, freelist, 0)
		blacklist := getMember(l, "blacklist")
		require.Equal(t, netOffsets{0x00}, blacklist)
	})

	t.Run("LeaseManager reuse previously freed leases", func(t *testing.T) {
		setup()
		err := l.RemoveLeased("10.10.0.4", "10.10.0.2")
		require.Nil(t, err)
		freelist := getMember(l, "freelist")
		require.Len(t, freelist, 2)
		require.Equal(t, netOffsets{0x02, 0x04}, freelist)
		leases, err := l.Lease()
		require.Nil(t, err)
		require.NotNil(t, leases)
		require.Equal(t, []string{"10.10.0.4"}, leases)
		freelist = getMember(l, "freelist")
		require.Len(t, freelist, 1)
		require.Equal(t, netOffsets{0x02}, freelist)
		lmLeases := getMember(l, "leases")
		require.Equal(t, netOffsets{0x01, 0x03, 0x04, 0x05}, lmLeases)
	})

	t.Run("LeaseManager blacklist current lease", func(t *testing.T) {
		setup()
		err := l.Blacklist("10.10.0.1")
		require.NotNil(t, err)
		require.Equal(t, "cannot blacklist currently leased IP 10.10.0.1", err.Error())
		lmLeases := getMember(l, "leases")
		require.Equal(t, netOffsets{0x01, 0x02, 0x03, 0x04, 0x05}, lmLeases)
		freelist := getMember(l, "freelist")
		require.Len(t, freelist, 0)
		blacklist := getMember(l, "blacklist")
		require.Equal(t, netOffsets{0x00}, blacklist)
	})

	t.Run("LeaseManager blacklist blacklisted IP", func(t *testing.T) {
		setup()
		err := l.Blacklist("10.10.0.0")
		require.NotNil(t, err)
		require.Equal(t, "cannot blacklist existing blacklisted IP 10.10.0.0", err.Error())
		lmLeases := getMember(l, "leases")
		require.Equal(t, netOffsets{0x01, 0x02, 0x03, 0x04, 0x05}, lmLeases)
		freelist := getMember(l, "freelist")
		require.Len(t, freelist, 0)
		blacklist := getMember(l, "blacklist")
		require.Equal(t, netOffsets{0x00}, blacklist)
	})

	t.Run("LeaseManager duplicate blacklisted IPs", func(t *testing.T) {
		setup()
		err := l.Blacklist("10.10.0.6", "10.10.0.6", "10.10.0.7")
		require.NotNil(t, err)
		require.Equal(t, "cannot blacklist duplicate IP 10.10.0.6", err.Error())
		lmLeases := getMember(l, "leases")
		require.Equal(t, netOffsets{0x01, 0x02, 0x03, 0x04, 0x05}, lmLeases)
		freelist := getMember(l, "freelist")
		require.Len(t, freelist, 0)
		blacklist := getMember(l, "blacklist")
		require.Equal(t, netOffsets{0x00}, blacklist)
	})

	t.Run("LeaseManager blacklist IP on freelist", func(t *testing.T) {
		setup()
		err := l.RemoveLeased("10.10.0.4")
		require.Nil(t, err)
		err = l.Blacklist("10.10.0.4")
		require.Nil(t, err)
		lmLeases := getMember(l, "leases")
		require.Equal(t, netOffsets{0x01, 0x02, 0x03, 0x05}, lmLeases)
		freelist := getMember(l, "freelist")
		require.Len(t, freelist, 0)
		blacklist := getMember(l, "blacklist")
		require.Equal(t, netOffsets{0x00, 0x04}, blacklist)
	})

	t.Run("LeaseManager remove nonexistent blacklisted IP", func(t *testing.T) {
		setup()
		err := l.RemoveBlacklisted("10.10.0.1", "10.10.0.0")
		require.NotNil(t, err)
		require.Equal(t, "cannot remove nonexistent blacklisted IP 10.10.0.1", err.Error())
		blacklist := getMember(l, "blacklist")
		require.Equal(t, netOffsets{0x00}, blacklist)
	})

	t.Run("LeaseManager remove duplicate blacklisted IP", func(t *testing.T) {
		setup()
		err := l.RemoveBlacklisted("10.10.0.0", "10.10.0.0")
		require.NotNil(t, err)
		require.Equal(t, "cannot remove duplicate blacklisted IP 10.10.0.0", err.Error())
		blacklist := getMember(l, "blacklist")
		require.Equal(t, netOffsets{0x00}, blacklist)
	})

	t.Run("LeaseManager get net ip", func(t *testing.T) {
		setup()
		ip := l.GetNetIP()
		require.Equal(t, "10.10.0.0", ip)
	})

	t.Run("LeaseManager net contains", func(t *testing.T) {
		setup()
		contained := l.NetContains("192.168.1.1")
		require.False(t, contained)
		contained = l.NetContains("10.10.0.0")
		require.True(t, contained)
	})
}
