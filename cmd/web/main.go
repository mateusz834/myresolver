package main

import (
	"errors"
	"flag"
	"fmt"
	"net/netip"
	"os"
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

	dns, err := netip.ParseAddrPort(*dnsAddr)
	if err != nil {
		return err
	}

	srv := NewServer(*baseDomain)
	return srv.Run(dns, *httpAddr)
}
