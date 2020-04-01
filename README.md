# go-aggregate
[![License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](http://opensource.org/licenses/MIT)
[![Actions Status](https://github.com/ldkingvivi/go-aggregate/workflows/Go/badge.svg)](https://github.com/ldkingvivi/go-aggregate/actions)
[![Build Status](https://travis-ci.org/ldkingvivi/go-aggregate.png?branch=master)](https://travis-ci.org/ldkingvivi/go-aggregate)
[![codecov](https://codecov.io/gh/ldkingvivi/go-aggregate/branch/master/graph/badge.svg)](https://codecov.io/gh/ldkingvivi/go-aggregate)

# What is this
This is the go implementation of the original aggregate from [@horms]( https://github.com/horms) on linux back in 2002, but more generic, you can implement the interface and make it very flexible

### Basic Example

```
package main

import (
	agg "github.com/ldkingvivi/go-aggregate"
	"log"
	"net"
)

func main() {

	// example use NewBasicCidrEntry for basic aggregate
	_, aNet, _ := net.ParseCIDR("8.8.8.0/25")
	a := agg.NewBasicCidrEntry(aNet)

	_, bNet, _ := net.ParseCIDR("9.9.9.0/25")
	b := agg.NewBasicCidrEntry(bNet)

	_, cNet, _ := net.ParseCIDR("8.8.8.128/25")
	c := agg.NewBasicCidrEntry(cNet)

	// empty merge func will do the basic merge
	result := agg.Aggregate([]agg.CidrEntry{a, b, c}, func(_, _ agg.CidrEntry) {})
	for _, cidr := range result {
		log.Printf("%s", cidr.GetNetwork())
		//2020/03/29 22:02:12 8.8.8.0/24
		//2020/03/29 22:02:12 9.9.9.0/25
	}
}
```

### Custom Struct Example

```
package main

import (
	agg "github.com/ldkingvivi/go-aggregate"
	"log"
	"net"
)

type customCidrEntry struct {
	ipNet *net.IPNet
	count int
	note  string
}

func (c *customCidrEntry) GetNetwork() *net.IPNet {
	return c.ipNet
}

func (c *customCidrEntry) SetNetwork(ipNet *net.IPNet) {
	c.ipNet = ipNet
}

func NewCustomCidrEntry(ipNet *net.IPNet, count int, note string) agg.CidrEntry {
	return &customCidrEntry{
		ipNet: ipNet,
		count: count,
		note:  note,
	}
}

func main() {
	// example use custom interface with client's own merge logic
	_, xNet, _ := net.ParseCIDR("8.8.8.128/25")
	_, yNet, _ := net.ParseCIDR("8.8.8.0/25")

	x := NewCustomCidrEntry(xNet, 10, "US")
	y := NewCustomCidrEntry(yNet, 20, "US")

	// add CIDR's count when merged
	result := agg.Aggregate([]agg.CidrEntry{x, y}, func(keep, delete agg.CidrEntry) {
		specificKeep, _ := keep.(*customCidrEntry)
		specificDelete, _ := delete.(*customCidrEntry)
		specificKeep.count += specificDelete.count
	})

	for _, cidr := range result {
		custom, ok := cidr.(*customCidrEntry)
		if ok {
			log.Printf("%s count : %d with note: %s",
				custom.GetNetwork(), custom.count, custom.note)
			//2020/03/29 22:25:10 8.8.8.0/24 count : 30 with note: US
		}
	}
}

```

### BenchMark with following string
```
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
```

```
goos: darwin
goarch: amd64
pkg: github.com/ldkingvivi/go-aggregate
BenchmarkAggregateMergeAddCount-12        	   65290	     17640 ns/op	   18056 B/op	     204 allocs/op
BenchmarkAggregateMergeUseDeletNote-12    	   66913	     17600 ns/op	   18056 B/op	     204 allocs/op
BenchmarkAggregateMergeDoNothing-12       	   67716	     17702 ns/op	   18056 B/op	     204 allocs/op
```