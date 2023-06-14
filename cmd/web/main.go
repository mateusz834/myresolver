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
	srv := NewServer("rand.api.who-resolved.example.com.")
	return srv.Run(netip.MustParseAddrPort("[::]:5333"), ":8080")
}
