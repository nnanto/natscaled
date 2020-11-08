package main

import (
	"flag"
	"natscaled/nat"
)

func main() {
	// Read options
	verbose := flag.Bool("v", false, "enable logging")
	startTailscaleServer := flag.Bool("tsd", false, "start tailscale daemon")
	serverName := flag.String("serverName", "", "name of the server")
	flag.Parse()

	// Start server
	s := &nat.Server{}
	if err := s.Start(*serverName, *startTailscaleServer, *verbose); err != nil {
		panic(err)
	}
}
