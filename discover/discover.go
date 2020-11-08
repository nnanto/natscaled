package discover

import (
	"context"
	"github.com/nnanto/tamed/client"
	"log"
	"time"
)

// discTimeout specifies duration to wait for search of existing server
var discTimeout = 15 * time.Second

// Service is discovery service helps in identifying peers.
// This runs over tailscale using nnanto/tamed package
type Service struct {
	client         *client.Client
	notificationCh chan client.Notify
	ip             string
}

// Start starts a tailscale client (and server if needed) and waits for notification status to
// be received
func Start(ctx context.Context, startServer bool) (*Service, error) {
	s := &Service{
		notificationCh: make(chan client.Notify),
	}
	var err error
	if s.client, err = startTS(ctx, s.notificationCh, startServer); err != nil {
		return nil, err
	}
	s.waitForStatusUpdate()
	return s, nil
}

// IP gets tailscale IP
func (s *Service) IP() string {
	return s.ip
}

// Peers gets list of tailscale IPs of peers
func (s *Service) Peers() []string {
	var peers []string
	for _, p := range s.client.AllPeers() {
		peers = append(peers, p.IP())
	}
	return peers
}

// Wait for status update from tailscale or till timeout
func (s *Service) waitForStatusUpdate() {

	select {
	case <-s.afterSelfStatusUpdate():
		log.Println("Status Update Received")
	case <-time.After(discTimeout):
		log.Println("Wait for status timed out")
	}

}

func (s *Service) afterSelfStatusUpdate() <-chan struct{} {
	log.Println("Start waiting for notification from client")
	ch := make(chan struct{})
	go func() {
		for notify := range s.notificationCh {
			if notify.Status != nil {
				if notify.Status.Self.TailAddr != "" {
					log.Printf("Self info: %v", notify.Status.Self.TailAddr)
					s.ip = notify.Status.Self.TailAddr
					close(ch)
					break
				}
			}
		}
	}()

	return ch
}
