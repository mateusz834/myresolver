package main

import (
	"errors"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"strings"

	"github.com/mateusz834/myresolver"
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
	adndb := flag.String("asndb", "", "")
	flag.Parse()

	var ipdb *myresolver.IPDB

	if *adndb != "" {
		db, err := myresolver.ParseIPDBFromFile(*adndb)
		if err != nil {
			return fmt.Errorf("failed while oppening the mmdb file: %v", err)
		}
		ipdb = &db
	}

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

	srv := NewServer(ipdb, *baseDomain)
	return srv.Run(dnsAddrs, *httpAddr)
}
