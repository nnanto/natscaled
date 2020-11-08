package main

import (
	"context"
	"flag"
	"fmt"
	nserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nnanto/tamed/client"
	"log"
	"natscaled/discover"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	clientPort  = 4222
	clusterPort = 4223
)

var myIp = ""

type Server struct {
	disc discover.Service
	srv  nserver.Server
	ip   string
}

func main() {
	// Read options
	doLog := flag.Bool("v", false, "enable logging")
	serverName := flag.String("serverName", "", "name of the server")
	flag.Parse()

	s, cancel := startDiscovery(25 * time.Second)
	routes := getRoutes(s.Client.AllPeers())
	log.Printf("Listening on IP [%v:%v] for routes from %v", myIp, clusterPort, routes)
	opts := &nserver.Options{
		ServerName:   *serverName,
		NoLog:        !(*doLog),
		TraceVerbose: true,
		Trace:        true,
		Debug:        true,
		Host:         myIp,
		Port:         clientPort,
		Cluster: nserver.ClusterOpts{
			Host:      myIp,
			Port:      clusterPort,
			Advertise: myIp,
		},
	}
	if routes != "" {
		opts.Routes = nserver.RoutesFromStr(routes)
	}
	ns, err := nserver.NewServer(opts)
	if err != nil {
		panic(err)
	}

	ns.ConfigureLogger()

	go func() {
		n := make(chan os.Signal)
		signal.Notify(n, os.Kill, os.Interrupt)
		<-n
		cancel()
	}()

	if err = nserver.Run(ns); err != nil {
		panic(err)
	}
}

func getRoutes(peers []client.PeerStatus) string {
	var natRoutes []string
	for _, p := range peers {
		natRoutes = append(natRoutes, fmt.Sprintf("nats://%s:%v", p.IP(), clusterPort))
	}
	if len(natRoutes) == 0 {
		return ""
	}
	return strings.Join(natRoutes, ",")
}

func startDiscovery(timeout time.Duration) (*discover.Service, context.CancelFunc) {
	op := client.DefaultOptions()
	//op.Logger = log.Printf
	op.ListenerCh = make(chan client.Notify)
	s := discover.NewService(op, true)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := s.Start(ctx); err != nil {
			log.Fatalf("Failed to discover: %v\n", err)
		}
	}()

	waitForStatusUpdate(op.ListenerCh, timeout)
	return s, cancel
}

func waitForStatusUpdate(notifyCh chan client.Notify, timeout time.Duration) {

	select {
	case <-waitForStatus(notifyCh):
		log.Println("Status Update Received")
	case <-time.After(timeout):
		log.Println("Wait for status timed out")
	}

}

func waitForStatus(notifyCh chan client.Notify) <-chan struct{} {
	log.Println("Start waiting for notification from client")
	ch := make(chan struct{})
	go func() {
		for notify := range notifyCh {
			if notify.Status != nil {
				if notify.Status.Self.TailAddr != "" {
					log.Printf("Self info: %v", notify.Status.Self.TailAddr)
					myIp = notify.Status.Self.TailAddr
					close(ch)
					break
				}

			}
			if notify.PingRequest != nil {

			}
		}
	}()

	return ch
}
