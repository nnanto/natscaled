package nat

import (
	"context"
	"fmt"
	nserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nnanto/tamed/client"
	"natscaled/discover"
	"os"
	"os/signal"
	"strings"
)

const (
	clientPort  = 4222 // Port that nat client connects to
	clusterPort = 4223 // Port that other nat servers in cluster communicates to
)

type Server struct {
	disc               *discover.Service  // tailscale service
	discNotificationCh chan client.Notify // notification from tailscale client
	srv                *nserver.Server    // nat server

}

// Start starts NAT server along with tailscale client (and server if startTailscaleDaemon is true)
func (s *Server) Start(serverName string, startTailscaleDaemon bool, verbose bool) error {
	ctx, cancel := context.WithCancel(context.Background())

	// start tailscale client
	ds, err := discover.Start(ctx, startTailscaleDaemon)
	if err != nil {
		return err
	}

	// initialize NAT options
	option := &nserver.Options{
		ServerName:   serverName,
		NoLog:        !verbose,
		TraceVerbose: true,
		Trace:        true,
		Debug:        true,
		Host:         ds.IP(),
		Port:         clientPort,
		Cluster: nserver.ClusterOpts{
			Host:      ds.IP(),
			Port:      clusterPort,
			Advertise: ds.IP(),
		},
	}

	// adds peers from tailscale as routes
	if routes := getRoutes(ds.Peers()); routes != "" {
		option.Routes = nserver.RoutesFromStr(routes)
	}

	// handle kill signals
	go func() {
		n := make(chan os.Signal)
		signal.Notify(n, os.Kill, os.Interrupt)
		<-n
		cancel()
	}()

	// start NAT server

	ns, err := nserver.NewServer(option)
	if err != nil {
		panic(err)
	}

	ns.ConfigureLogger()

	if err = nserver.Run(ns); err != nil {
		return err
	}
	return nil
}

// getRoutes gets cluster ip:host for peer NAT servers
func getRoutes(peers []string) string {
	var natRoutes []string
	for _, ip := range peers {
		natRoutes = append(natRoutes, fmt.Sprintf("nats://%s:%v", ip, clusterPort))
	}
	if len(natRoutes) == 0 {
		return ""
	}
	return strings.Join(natRoutes, ",")
}
