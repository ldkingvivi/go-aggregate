package Agg

import (
	"net/netip"
	"reflect"
	"strconv"
	"testing"
)

func TestGetIPPrefix(t *testing.T) {
	ip1 := netip.MustParseAddr("0.0.0.0")

	got1 := getIPPrefix(ip1)
	if got1 != 0 {
		t.Errorf("expect 0 but got %+v", got1)
	}

	ip2 := netip.MustParseAddr("8.8.8.0")
	got2 := getIPPrefix(ip2)
	if got2 != 21 {
		t.Errorf("expect 21 but got %+v", got2)
	}

	ip3 := netip.MustParseAddr("8.8.8.8")
	got3 := getIPPrefix(ip3)
	if got3 != 29 {
		t.Errorf("expect 29 but got %+v", got3)
	}

	ip4 := netip.MustParseAddr("2620:108:700f::3645:f643")
	got4 := getIPPrefix(ip4)
	if got4 != 128 {
		t.Errorf("expect 128 but got %+v", got4)
	}

	ipnet4 := netip.MustParsePrefix("2620:108:700f::3645:f643/64")
	got5 := getIPPrefix(ipnet4.Masked().Addr())
	if got5 != 48 {
		t.Errorf("expect 48 but got %+v", got5)
	}
}

type testResults struct {
	ipnetString string
	count       int
}

type customCidrEntry struct {
	ipNet netip.Prefix
	count int
	note  string
}

func (c *customCidrEntry) GetNetwork() netip.Prefix {
	return c.ipNet
}

func (c *customCidrEntry) SetNetwork(ipNet netip.Prefix) {
	c.ipNet = ipNet
}

func (c *customCidrEntry) GetCount() int {
	return c.count
}

func (c *customCidrEntry) SetCount(count int) {
	c.count = count
}

func NewCustomCidrEntry(ipNet netip.Prefix, count int, note string) CidrEntry {
	return &customCidrEntry{
		ipNet: ipNet,
		count: count,
		note:  note,
	}
}

func mergeAddCount(k, d CidrEntry) {
	sk, _ := k.(*customCidrEntry)
	sd, _ := d.(*customCidrEntry)

	sk.count += sd.count
}

func mergeUseDeleteNote(k, d CidrEntry) {

	sk, _ := k.(*customCidrEntry)
	sd, _ := d.(*customCidrEntry)

	sk.note = sd.note
}

func mergeDoNothing(_, _ CidrEntry) {
}

func TestAggregateAddCount(t *testing.T) {

	var got []CidrEntry

	for i, c := range []struct {
		in   []string
		want []testResults
	}{
		// Empty
		{
			[]string{},
			[]testResults{},
		},
		// Single
		{
			[]string{"8.8.8.0/24"},
			[]testResults{
				{
					"8.8.8.0/24",
					1,
				},
			},
		},
		// IPv4 prefixes
		{
			[]string{
				"8.8.8.8/29", "8.8.8.0/24",
			},
			[]testResults{
				{
					"8.8.8.0/24",
					2,
				},
			},
		},
		{
			[]string{
				"8.8.8.8/29", "8.8.8.0/29",
			},
			[]testResults{
				{
					"8.8.8.0/28",
					2,
				},
			},
		},
		{
			[]string{
				"8.8.8.8/29", "8.8.8.16/29",
			},
			[]testResults{
				{
					"8.8.8.8/29",
					1,
				},
				{
					"8.8.8.16/29",
					1,
				},
			},
		},
		{
			[]string{
				"8.8.8.0/25", "9.9.9.0/25", "8.8.8.128/25",
			},
			[]testResults{
				{
					"8.8.8.0/24",
					2,
				},
				{
					"9.9.9.0/25",
					1,
				},
			},
		},
		{
			[]string{
				"192.0.2.0/25", "192.0.2.128/25",
			},
			[]testResults{
				{
					"192.0.2.0/24",
					2,
				},
			},
		},
		{
			[]string{
				"192.0.2.0/26", "192.0.2.64/26", "192.0.2.128/26", "192.0.2.192/26",
			},
			[]testResults{
				{
					"192.0.2.0/24",
					4,
				},
			},
		},
		{
			[]string{
				"192.0.2.0/27", "192.0.2.32/27", "192.0.2.64/27", "192.0.2.96/27",
				"192.0.2.128/27", "192.0.2.160/27", "192.0.2.192/27", "192.0.2.224/27",
			},
			[]testResults{
				{
					"192.0.2.0/24",
					8,
				},
			},
		},
		{
			[]string{
				"192.0.2.0/28", "192.0.2.16/28", "192.0.2.32/28", "192.0.2.48/28",
				"192.0.2.64/28", "192.0.2.80/28", "192.0.2.96/28", "192.0.2.112/28",
				"192.0.2.128/28", "192.0.2.144/28", "192.0.2.160/28", "192.0.2.176/28",
				"192.0.2.192/28", "192.0.2.208/28", "192.0.2.224/28", "192.0.2.240/28",
			},
			[]testResults{
				{
					"192.0.2.0/24",
					16,
				},
			},
		},
		{
			[]string{
				"192.0.2.0/29", "192.0.2.8/29", "192.0.2.16/29", "192.0.2.24/29",
				"192.0.2.32/29", "192.0.2.40/29", "192.0.2.48/29", "192.0.2.56/29",
				"192.0.2.64/29", "192.0.2.72/29", "192.0.2.80/29", "192.0.2.88/29",
				"192.0.2.96/29", "192.0.2.104/29", "192.0.2.112/29", "192.0.2.120/29",
				"192.0.2.128/29", "192.0.2.136/29", "192.0.2.144/29", "192.0.2.152/29",
				"192.0.2.160/29", "192.0.2.168/29", "192.0.2.176/29", "192.0.2.184/29",
				"192.0.2.192/29", "192.0.2.200/29", "192.0.2.208/29", "192.0.2.216/29",
				"192.0.2.224/29", "192.0.2.232/29", "192.0.2.240/29", "192.0.2.248/29",
			},
			[]testResults{
				{
					"192.0.2.0/24",
					32,
				},
			},
		},
		{
			[]string{
				"192.0.2.0/26", "192.0.2.64/26", "192.0.2.192/26",
				"192.0.2.128/28", "192.0.2.144/28", "192.0.2.160/28", "192.0.2.176/28",
			},
			[]testResults{
				{
					"192.0.2.0/24",
					7,
				},
			},
		},
		{
			[]string{
				"192.0.2.1/32", "192.0.2.1/32",
			},
			[]testResults{
				{
					"192.0.2.1/32",
					2,
				},
			},
		},
		{
			[]string{
				"192.0.2.0/25", "192.0.2.128/25",
				"192.0.2.248/29",
			},
			[]testResults{
				{
					"192.0.2.0/24",
					3,
				},
			},
		},
		{
			[]string{
				"192.0.2.0/24",
				"198.51.100.0/24",
				"203.0.113.0/24",
			},
			[]testResults{
				{
					"192.0.2.0/24",
					1,
				},
				{
					"198.51.100.0/24",
					1,
				},
				{
					"203.0.113.0/24",
					1,
				},
			},
		},
		{
			[]string{
				"192.0.2.0/25",
				"192.0.2.0/26",
				"192.0.2.0/27",
				"192.0.2.0/28",
				"192.0.2.0/29",
				"192.0.2.0/30",
			},
			[]testResults{
				{
					"192.0.2.0/25",
					6,
				},
			},
		},
		{
			[]string{
				"0.0.0.0/0",
				"192.0.2.0/24", "198.51.100.0/24", "203.0.113.0/24",
				"255.255.255.255/32",
			},
			[]testResults{
				{
					"0.0.0.0/0",
					5,
				},
			},
		},
		{
			[]string{
				"0.0.0.0/0", "0.0.0.0/0",
				"255.255.255.255/32", "255.255.255.255/32",
			},
			[]testResults{
				{
					"0.0.0.0/0",
					4,
				},
			},
		},
		{
			[]string{
				"192.168.0.0/25", "192.168.0.128/25",
				"192.168.1.0/24", "192.168.3.0/24", "192.168.4.0/24",
				"192.168.5.0/26",
				"192.168.128.0/22", "192.168.132.0/22",
				"192.168.128.0/21",
			},
			[]testResults{
				{
					"192.168.0.0/23",
					3,
				},
				{
					"192.168.3.0/24",
					1,
				},
				{
					"192.168.4.0/24",
					1,
				},
				{
					"192.168.5.0/26",
					1,
				},
				{
					"192.168.128.0/21",
					3,
				},
			},
		},
		{
			[]string{
				"192.168.0.0/25", "192.168.0.128/25",
				"192.168.1.0/24", "192.168.3.0/24", "192.168.4.0/24",
				"192.168.5.0/26",
			},
			[]testResults{
				{
					"192.168.0.0/23",
					3,
				},
				{
					"192.168.3.0/24",
					1,
				},
				{
					"192.168.4.0/24",
					1,
				},
				{
					"192.168.5.0/26",
					1,
				},
			},
		},
		{
			[]string{
				"192.0.2.0/25", "198.51.100.0/25", "192.0.2.128/25",
			},
			[]testResults{
				{
					"192.0.2.0/24",
					2,
				},
				{
					"198.51.100.0/25",
					1,
				},
			},
		},
		// IPv6 prefixes
		{
			[]string{
				"2001:db8::/64", "2001:db8:0:1::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64",
				"2001:db8:0:4::/64",
			},
			[]testResults{
				{
					"2001:db8::/62",
					4,
				},
				{
					"2001:db8:0:4::/64",
					1,
				},
			},
		},
		{
			[]string{
				"::/0",
				"2001:db8::/32",
				"2001:db8::/126",
				"2001:db8::/127",
				"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128",
			},
			[]testResults{
				{
					"::/0",
					5,
				},
			},
		},
		{
			[]string{
				"::/0", "::/0",
				"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128",
			},
			[]testResults{
				{
					"::/0",
					4,
				},
			},
		},
		// Mix IPv4 and IPv6
		{
			[]string{
				"192.0.2.0/29", "192.0.2.8/29", "192.0.2.16/29", "192.0.2.24/29",
				"192.0.2.32/29", "192.0.2.40/29", "192.0.2.48/29", "192.0.2.56/29",
				"192.0.2.64/29", "192.0.2.72/29", "192.0.2.80/29", "192.0.2.88/29",
				"192.0.2.96/29", "192.0.2.104/29", "192.0.2.112/29", "192.0.2.120/29",
				"2001:db8::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64", "2001:db8:0:1::/64",
				"192.0.2.128/29", "192.0.2.136/29", "192.0.2.144/29", "192.0.2.152/29",
				"192.0.2.160/29", "192.0.2.176/29", "192.0.2.184/29", "192.0.2.168/32",
				"192.0.2.192/29", "192.0.2.200/29", "192.0.2.208/29", "192.0.2.216/29",
				"192.0.2.224/29", "192.0.2.232/29", "192.0.2.240/29", "192.0.2.248/29",
				"2001:db8:0:4::/64", "192.0.2.171/32", "192.0.2.172/32", "192.0.2.174/32",
				"192.0.2.169/32", "192.0.2.170/32", "192.0.2.173/32", "192.0.2.175/32",
			},
			[]testResults{
				{
					"192.0.2.0/24",
					39,
				},
				{
					"2001:db8::/62",
					4,
				},
				{
					"2001:db8:0:4::/64",
					1,
				},
			},
		},
	} {

		var cidrEntries []CidrEntry
		var cidrWant []CidrEntry

		for _, s := range c.in {
			ipnet := netip.MustParsePrefix(s)
			cidrEntries = append(cidrEntries, NewCustomCidrEntry(ipnet, 1, "US"))
		}

		for _, s := range c.want {
			ipnet := netip.MustParsePrefix(s.ipnetString)
			cidrWant = append(cidrWant, NewCustomCidrEntry(ipnet, s.count, "US"))
		}

		got = Aggregate(cidrEntries, mergeAddCount)

		if !reflect.DeepEqual(got, cidrWant) {
			t.Errorf("#%d: expect: %+v , but got %+v", i, cidrWant, got)
		}
	}
}

func TestAggregateWithGivenCount(t *testing.T) {

	var input = []testResults{
		{
			"8.8.9.128/25",
			1,
		},
		{
			"8.8.8.0/24",
			39,
		},
		{
			"8.8.9.0/25",
			4,
		},
	}

	var want = []testResults{
		{
			"8.8.8.0/23",
			44,
		},
	}

	var inputCidrs []CidrEntry
	for _, s := range input {
		ipnet := netip.MustParsePrefix(s.ipnetString)
		inputCidrs = append(inputCidrs, NewCustomCidrEntry(ipnet, s.count, "US"))
	}

	got := Aggregate(inputCidrs, mergeAddCount)

	var cidrWant []CidrEntry
	for _, s := range want {
		ipnet := netip.MustParsePrefix(s.ipnetString)
		cidrWant = append(cidrWant, NewCustomCidrEntry(ipnet, s.count, "US"))
	}

	if !reflect.DeepEqual(got, cidrWant) {
		t.Errorf("expect: %+v , but got %+v", cidrWant, got)
	}
}

func TestAggregateWithMergeDeleteNote(t *testing.T) {

	// example use custom interface
	xNet := netip.MustParsePrefix("8.8.8.128/25")
	yNet := netip.MustParsePrefix("8.8.8.0/25")

	x := NewCustomCidrEntry(xNet, 10, "US")
	y := NewCustomCidrEntry(yNet, 20, "CA")

	got := Aggregate([]CidrEntry{x, y}, mergeUseDeleteNote)

	if len(got) != 1 {
		t.Errorf("expect single results")
	}

	gotS, ok := got[0].(*customCidrEntry)
	if !ok {
		t.Errorf("error to map type back")
	}

	expect := "US"

	if gotS.note != expect {
		t.Errorf("expect %s, but got %s", expect, gotS.note)
	}
}

func TestAggregateWithMergeDoNothing(t *testing.T) {

	var input = []string{
		"8.8.9.128/25",
		"8.8.8.0/24",
		"8.8.9.0/25",
	}

	var want = []string{
		"8.8.8.0/23",
	}

	var inputCidrs []CidrEntry
	for _, s := range input {
		ipnet := netip.MustParsePrefix(s)
		inputCidrs = append(inputCidrs, NewBasicCidrEntry(ipnet))
	}

	got := Aggregate(inputCidrs, mergeDoNothing)

	var cidrWant []CidrEntry
	for _, s := range want {
		ipnet := netip.MustParsePrefix(s)
		cidrWant = append(cidrWant, NewBasicCidrEntry(ipnet))
	}

	if !reflect.DeepEqual(got, cidrWant) {
		t.Errorf("expect: %+v , but got %+v", cidrWant, got)
	}
}

func TestAggregateWithMergeDoNothing65K(t *testing.T) {
	var inputCidrs []CidrEntry

	var c, d int
	var cStr, dStr string

	for c = 0; c < 256; c++ {
		cStr = strconv.Itoa(c)
		for d = 0; d < 256; d++ {
			dStr = strconv.Itoa(d)
			ipnet := netip.MustParsePrefix("1.1." + cStr + "." + dStr + "/32")
			inputCidrs = append(inputCidrs, NewBasicCidrEntry(ipnet))
		}
	}

	got := Aggregate(inputCidrs, mergeDoNothing)

	var cidrWant []CidrEntry
	ipnet := netip.MustParsePrefix("1.1.0.0/16")
	cidrWant = append(cidrWant, NewBasicCidrEntry(ipnet))

	if !reflect.DeepEqual(got, cidrWant) {
		t.Errorf("expect: %+v , but got %+v", cidrWant, got)
	}

}

func BenchmarkAggregateMergeAddCount(b *testing.B) {
	input := []string{
		"192.0.2.160/29", "192.0.2.176/29", "192.0.2.184/29", "192.0.2.168/32",
		"192.0.2.0/29", "192.0.2.8/29", "192.0.2.16/29", "192.0.2.24/29",
		"192.0.2.32/29", "192.0.2.40/29", "192.0.2.48/29", "192.0.2.56/29",
		"192.0.2.64/29", "192.0.2.72/29", "192.0.2.80/29", "192.0.2.88/29",
		"2001:db8::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64", "2001:db8:0:1::/64",
		"192.0.2.128/29", "192.0.2.136/29", "192.0.2.144/29", "192.0.2.152/29",
		"192.0.2.192/29", "192.0.2.200/29", "192.0.2.208/29", "192.0.2.216/29",
		"192.0.2.224/29", "192.0.2.232/29", "192.0.2.240/29", "192.0.2.248/29",
		"2001:db8:0:4::/64", "192.0.2.171/32", "192.0.2.172/32", "192.0.2.174/32",
		"192.0.2.169/32", "192.0.2.170/32", "192.0.2.173/32", "192.0.2.175/32",
		"192.0.2.96/29", "192.0.2.104/29", "192.0.2.112/29", "192.0.2.120/29",
	}

	var cidrEntries []CidrEntry
	for _, s := range input {
		ipnet := netip.MustParsePrefix(s)
		cidrEntries = append(cidrEntries, NewCustomCidrEntry(ipnet, 1, "US"))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Aggregate(cidrEntries, mergeAddCount)
	}
}

func BenchmarkAggregateMergeUseDeletNote(b *testing.B) {
	input := []string{
		"192.0.2.160/29", "192.0.2.176/29", "192.0.2.184/29", "192.0.2.168/32",
		"192.0.2.0/29", "192.0.2.8/29", "192.0.2.16/29", "192.0.2.24/29",
		"192.0.2.32/29", "192.0.2.40/29", "192.0.2.48/29", "192.0.2.56/29",
		"192.0.2.64/29", "192.0.2.72/29", "192.0.2.80/29", "192.0.2.88/29",
		"2001:db8::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64", "2001:db8:0:1::/64",
		"192.0.2.128/29", "192.0.2.136/29", "192.0.2.144/29", "192.0.2.152/29",
		"192.0.2.192/29", "192.0.2.200/29", "192.0.2.208/29", "192.0.2.216/29",
		"192.0.2.224/29", "192.0.2.232/29", "192.0.2.240/29", "192.0.2.248/29",
		"2001:db8:0:4::/64", "192.0.2.171/32", "192.0.2.172/32", "192.0.2.174/32",
		"192.0.2.169/32", "192.0.2.170/32", "192.0.2.173/32", "192.0.2.175/32",
		"192.0.2.96/29", "192.0.2.104/29", "192.0.2.112/29", "192.0.2.120/29",
	}

	var cidrEntries []CidrEntry
	for _, s := range input {
		ipnet := netip.MustParsePrefix(s)
		cidrEntries = append(cidrEntries, NewCustomCidrEntry(ipnet, 10, "US"))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Aggregate(cidrEntries, mergeUseDeleteNote)
	}
}

func BenchmarkAggregateMergeDoNothing(b *testing.B) {
	input := []string{
		"192.0.2.160/29", "192.0.2.176/29", "192.0.2.184/29", "192.0.2.168/32",
		"192.0.2.0/29", "192.0.2.8/29", "192.0.2.16/29", "192.0.2.24/29",
		"192.0.2.32/29", "192.0.2.40/29", "192.0.2.48/29", "192.0.2.56/29",
		"192.0.2.64/29", "192.0.2.72/29", "192.0.2.80/29", "192.0.2.88/29",
		"2001:db8::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64", "2001:db8:0:1::/64",
		"192.0.2.128/29", "192.0.2.136/29", "192.0.2.144/29", "192.0.2.152/29",
		"192.0.2.192/29", "192.0.2.200/29", "192.0.2.208/29", "192.0.2.216/29",
		"192.0.2.224/29", "192.0.2.232/29", "192.0.2.240/29", "192.0.2.248/29",
		"2001:db8:0:4::/64", "192.0.2.171/32", "192.0.2.172/32", "192.0.2.174/32",
		"192.0.2.169/32", "192.0.2.170/32", "192.0.2.173/32", "192.0.2.175/32",
		"192.0.2.96/29", "192.0.2.104/29", "192.0.2.112/29", "192.0.2.120/29",
	}

	var cidrEntries []CidrEntry
	for _, s := range input {
		ipnet := netip.MustParsePrefix(s)
		cidrEntries = append(cidrEntries, NewBasicCidrEntry(ipnet))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Aggregate(cidrEntries, mergeDoNothing)
	}
}

func BenchmarkAggregateMergeDoNothing16M(b *testing.B) {
	var cidrEntries []CidrEntry

	var x, c, d int
	var bStr, cStr, dStr string

	for x = 0; x < 256; x++ {
		bStr = strconv.Itoa(x)
		for c = 0; c < 256; c++ {
			cStr = strconv.Itoa(c)
			for d = 0; d < 256; d++ {
				dStr = strconv.Itoa(d)
				ipnet := netip.MustParsePrefix("1." + bStr + "." + cStr + "." + dStr + "/32")
				cidrEntries = append(cidrEntries, NewBasicCidrEntry(ipnet))
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Aggregate(cidrEntries, mergeDoNothing)
	}
}
