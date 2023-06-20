package main

import (
	"flag"
	"fmt"
	"net/netip"
	"os"

	"github.com/mateusz834/myresolver"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	addrStr := flag.String("addr", "[::]:53", "")
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

	addr, err := netip.ParseAddrPort(*addrStr)
	if err != nil {
		return fmt.Errorf("failed while parsing the addr: %v", err)
	}

	errChan := make(chan error, 1)

	go func() {
		errChan <- myresolver.ListenUDPDNS(addr, ipdb, nil)
	}()

	go func() {
		errChan <- myresolver.ListenTCPDNS(addr, ipdb, nil)
	}()

	return <-errChan
}
