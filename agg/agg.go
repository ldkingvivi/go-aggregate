package Agg

import (
	"math/big"
	"net"
	"sort"
	"strconv"
)

type cidr struct {
	netIP       net.IP
	startIP     *big.Int
	nextStartIP *big.Int // this is end IP + 1
	ones        int
	bits        int

	prev *cidr
	next *cidr
}

func AggregateIPNet(cidrIPNet []*net.IPNet) ([]*net.IPNet, error) {

	if len(cidrIPNet) < 2 {
		// let's do nothing if there's only 0 or 1 element
		return cidrIPNet, nil
	}

	cidrs := run(cidrIPNet)
	return getIPNet(cidrs), nil
}

func AggregateStr(cidrStrings []string) ([]string, error) {

	if len(cidrStrings) < 2 {
		// let's do nothing if there's only 0 or 1 element
		return cidrStrings, nil
	}

	var ipnets []*net.IPNet
	// convert
	for _, c := range cidrStrings {
		// convert
		_, ipnet, err := net.ParseCIDR(c)
		if err != nil {
			return []string{}, err
		}
		ipnets = append(ipnets, ipnet)
	}

	cidrs := run(ipnets)
	return getStrings(cidrs), nil
}

func convertIPNetToCidr(ipnets []*net.IPNet) []cidr {
	var cidrs []cidr
	bigOne := big.NewInt(1)

	// convert
	for _, ipnet := range ipnets {
		// cover IPv6
		startIP := big.NewInt(0)
		startIP.SetBytes(ipnet.IP)

		nextStartIP := big.NewInt(0)

		ones, bits := ipnet.Mask.Size()
		diff := uint(bits) - uint(ones)

		nextStartIP.Lsh(bigOne, diff)
		nextStartIP.Add(nextStartIP, startIP)

		cidrs = append(cidrs, cidr{
			netIP:       ipnet.IP,
			startIP:     startIP,
			nextStartIP: nextStartIP,
			ones:        ones,
			bits:        bits,
		})
	}

	return cidrs
}

func run(ipnets []*net.IPNet) []cidr {
	// convert
	cidrs := convertIPNetToCidr(ipnets)
	// sort it
	sortIt(cidrs)
	// add pointer
	addPointer(cidrs)
	// unlink the smaller ones that already in bigger ones
	unlinkCovered(cidrs)
	// do the aggregate
	aggregateAdj(cidrs)

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

func unlinkCovered(cidrs []cidr) {
	// check already done from Aggregate()
	currentP := &cidrs[0]
	nextP := currentP.next

	for nextP != nil {
		if currentP.nextStartIP.Cmp(nextP.nextStartIP) >= 0 {
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

func aggregateAdj(cidrs []cidr) {
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

func getStrings(cidrs []cidr) []string {
	var r []string
	currentP := &cidrs[0]

	for currentP != nil {
		r = append(r, currentP.netIP.String()+"/"+strconv.Itoa(currentP.ones))
		currentP = currentP.next
	}
	return r
}

func getIPNet(cidrs []cidr) []*net.IPNet {
	var r []*net.IPNet
	currentP := &cidrs[0]

	for currentP != nil {

		r = append(r, &net.IPNet{
			IP:   currentP.netIP,
			Mask: net.CIDRMask(currentP.ones, currentP.bits),
		})
		currentP = currentP.next
	}
	return r
}

func getIPPrefix(ip net.IP) int {

	if ip.To4() != nil {
		// ipv4
		return net.IPv4len*8 - getTrailingZero(ip)
	} else {
		// ipv6
		return net.IPv6len*8 - getTrailingZero(ip)
	}
}

func getTrailingZero(ip net.IP) int {
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
