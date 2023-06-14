package main

import (
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
	srv := NewServer("rand.api.get.my-resolver.834834.xyz")
	return srv.Run(netip.MustParseAddrPort("95.216.184.1:53"), ":80")
}
