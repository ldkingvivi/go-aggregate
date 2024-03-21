package Agg

import (
	"math/big"
	"net"
	"net/netip"
	"sort"
)

type cidr struct {
	netIP       netip.Addr
	startIP     *big.Int
	nextStartIP *big.Int // this is end IP + 1
	ones        int
	bits        int

	prev *cidr
	next *cidr

	entry CidrEntry
}

type CidrEntry interface {
	GetNetwork() netip.Prefix
	SetNetwork(netip.Prefix)
}

type basicCidrEntry struct {
	ipNet netip.Prefix
}

func (b *basicCidrEntry) GetNetwork() netip.Prefix {
	return b.ipNet
}

func (b *basicCidrEntry) SetNetwork(ipNet netip.Prefix) {
	b.ipNet = ipNet
}

func NewBasicCidrEntry(ipNet netip.Prefix) CidrEntry {
	return &basicCidrEntry{
		ipNet: ipNet,
	}
}

type Merge func(keep, delete CidrEntry)

func Aggregate(cidrEntries []CidrEntry, mergeFn Merge) []CidrEntry {
	if len(cidrEntries) < 2 {
		return cidrEntries
	}
	cidrs := convertToCidr(cidrEntries)
	// sort it
	sortIt(cidrs)
	// add pointer
	addPointer(cidrs)
	// unlink the smaller ones that already in bigger ones
	unlinkCovered(cidrs, mergeFn)
	// do the aggregate
	aggregateAdj(cidrs, mergeFn)

	return getEntries(cidrs)
}

func convertToCidr(cidrEntries []CidrEntry) []cidr {
	var cidrs []cidr
	bigOne := big.NewInt(1)
	var ipnet netip.Prefix
	// convert
	for _, cidrEntry := range cidrEntries {
		ipnet = cidrEntry.GetNetwork().Masked()

		// cover IPv6
		startIP := big.NewInt(0)
		startIP.SetBytes(ipnet.Addr().AsSlice())

		nextStartIP := big.NewInt(0)

		ones := ipnet.Bits()
		bits := ipnet.Addr().BitLen()
		diff := uint(bits) - uint(ones)

		nextStartIP.Lsh(bigOne, diff)
		nextStartIP.Add(nextStartIP, startIP)

		cidrs = append(cidrs, cidr{
			netIP:       ipnet.Addr(),
			startIP:     startIP,
			nextStartIP: nextStartIP,
			ones:        ones,
			bits:        bits,
			entry:       cidrEntry,
		})
	}

	return cidrs
}

func sortIt(cidrs []cidr) {
	sort.Slice(cidrs, func(i, j int) bool {
		startIPCmp := cidrs[i].startIP.Cmp(cidrs[j].startIP)
		if startIPCmp < 0 {
			return true
		} else if startIPCmp == 0 && cidrs[i].ones < cidrs[j].ones {
			return true
		}
		return false
	})
}

func addPointer(cidrs []cidr) {
	s := 0
	e := 1
	for e < len(cidrs) {
		cidrs[s].next = &cidrs[e]
		cidrs[e].prev = &cidrs[s]
		s++
		e++
	}
}

func unlinkCovered(cidrs []cidr, mergeFn Merge) {
	// check already done from Aggregate()
	currentP := &cidrs[0]
	nextP := currentP.next

	for nextP != nil {
		if currentP.nextStartIP.Cmp(nextP.nextStartIP) >= 0 {
			// run the merge func
			mergeFn(currentP.entry, nextP.entry)
			// skip the next
			currentP.next = nextP.next
			if nextP.next != nil {
				nextP.next.prev = currentP
			}
		} else {
			// only move current forward if current endIP not cover next endIP
			currentP = nextP
		}
		nextP = currentP.next
	}
}

func aggregateAdj(cidrs []cidr, mergeFn Merge) {
	// check already done from Aggregate()
	currentP := &cidrs[0]
	nextP := currentP.next

	for nextP != nil {

		if currentP.ones == nextP.ones &&
			currentP.nextStartIP.Cmp(nextP.startIP) == 0 &&
			getIPPrefix(currentP.netIP) < currentP.ones {
			// change current endIP and prefix
			// no need to change the netIP
			currentP.nextStartIP = nextP.nextStartIP
			currentP.ones = currentP.ones - 1
			// run the merge func
			mergeFn(currentP.entry, nextP.entry)

			// redo the link
			currentP.next = nextP.next
			if nextP.next != nil {
				nextP.next.prev = currentP
			}

			// try to move up if possible
			if currentP.prev != nil {
				nextP = currentP
				currentP = currentP.prev
			} else {
				nextP = nextP.next
			}
			continue
		}

		// move forward
		currentP = nextP
		nextP = currentP.next
	}
}

func getEntries(cidrs []cidr) []CidrEntry {
	var r []CidrEntry
	currentP := &cidrs[0]
	for currentP != nil {
		// update the entry network
		prefix := netip.PrefixFrom(currentP.netIP, currentP.ones)
		currentP.entry.SetNetwork(prefix)
		// added to results
		r = append(r, currentP.entry)
		// move to next
		currentP = currentP.next
	}
	return r
}

func getIPPrefix(ip netip.Addr) int {

	if ip.Is4() {
		// ipv4
		return net.IPv4len*8 - getTrailingZeroV4(ip.As4())
	} else {
		// ipv6
		return net.IPv6len*8 - getTrailingZeroV6(ip.As16())
	}
}

func getTrailingZeroV4(ip [4]byte) int {
	var n int
	var v byte

	i := len(ip) - 1
	for i >= 0 {
		v = ip[i]
		if v == 0x00 {
			n += 8
			i--
			continue
		}
		// found non-00 byte
		// count 0 bits
		for v&0x01 != 1 {
			n++
			v >>= 1
		}
		break
	}
	return n
}

func getTrailingZeroV6(ip [16]byte) int {
	var n int
	var v byte

	i := len(ip) - 1
	for i >= 0 {
		v = ip[i]
		if v == 0x00 {
			n += 8
			i--
			continue
		}
		// found non-00 byte
		// count 0 bits
		for v&0x01 != 1 {
			n++
			v >>= 1
		}
		break
	}
	return n
}
