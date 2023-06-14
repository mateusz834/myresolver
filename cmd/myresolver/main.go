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
	flag.Parse()

	addr, err := netip.ParseAddrPort(*addrStr)
	if err != nil {
		return fmt.Errorf("failed while parsing the addr: %v", err)
	}

	errChan := make(chan error, 1)

	go func() {
		errChan <- myresolver.ListenUDPDNS(addr, nil)
	}()

	go func() {
		errChan <- myresolver.ListenTCPDNS(addr, nil)
	}()

	return <-errChan
}
