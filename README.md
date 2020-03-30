# go-aggregate
[![License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](http://opensource.org/licenses/MIT)
[![Actions Status](https://github.com/ldkingvivi/go-aggregate/workflows/Go/badge.svg)](https://github.com/ldkingvivi/go-aggregate/actions)
[![Build Status](https://travis-ci.org/ldkingvivi/go-aggregate.png?branch=master)](https://travis-ci.org/ldkingvivi/go-aggregate)
[![codecov](https://codecov.io/gh/ldkingvivi/go-aggregate/branch/master/graph/badge.svg)](https://codecov.io/gh/ldkingvivi/go-aggregate)

this is the go implementation of the original aggregate from [@horms]( https://github.com/horms) on linux back in 2002

```
package main

import (
	agg "github.com/ldkingvivi/go-aggregate"
	"log"
	"net"
)

func main() {
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
	// use string
	x, err := agg.AggregateStr(input)
	if err != nil {
		log.Printf("%+v", err)
	} else {
		log.Printf("%+v", x) // [192.0.2.0/24 2001:db8::/62 2001:db8:0:4::/64]
	}

	var ipNets []*net.IPNet
	for _, s := range input {
		_, ipnet, err := net.ParseCIDR(s)
		if err != nil {
			log.Printf("%+v", err)
			continue
		}
		ipNets = append(ipNets, ipnet)
	}

	// use *net.IPNet
	y, err := agg.AggregateIPNet(ipNets)
	if err != nil {
		log.Printf("%+v", err)
	} else {
		log.Printf("%+v", y) // [192.0.2.0/24 2001:db8::/62 2001:db8:0:4::/64]
	}

}

```

BenchMark with following string

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
BenchmarkAggregateIPNet-12    	   61604	     18624 ns/op	   18056 B/op	     204 allocs/op
```