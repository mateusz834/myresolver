package main

import (
	"errors"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	baseDomain := flag.String("basedomain", "", "")
	dnsAddr := flag.String("dnsaddr", "[::]:53", "")
	httpAddr := flag.String("httpaddr", "[::]:80", "")
	flag.Parse()

	if *baseDomain == "" {
		return errors.New("basedomain flag is requierd")
	}

	dnsAddrs := make([]netip.AddrPort, 0)
	for _, dnsAddr := range strings.Split(*dnsAddr, ",") {
		addr, err := netip.ParseAddrPort(dnsAddr)
		if err != nil {
			return err
		}
		dnsAddrs = append(dnsAddrs, addr)
	}

	srv := NewServer(*baseDomain)
	return srv.Run(dnsAddrs, *httpAddr)
}
