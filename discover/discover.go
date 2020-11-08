package discover

import (
	"context"
	"github.com/nnanto/tamed/client"
	"github.com/nnanto/tamed/server"
	"log"
)

type Service struct {
	Client      *client.Client
	Option      *client.Option
	StartServer bool
}

func NewService(option *client.Option, startServer bool) *Service {
	if option == nil {
		option = client.DefaultOptions()
	}
	return &Service{
		Option:      option,
		StartServer: startServer,
	}
}

func (s *Service) Start(ctx context.Context) error {
	var err error
	if s.StartServer {
		err = s.startTamedService(ctx)
	} else {
		err = s.startTamedClient(ctx)
	}
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) startTamedService(ctx context.Context) error {
	serverNotification := make(chan server.Notify)
	go func() {
		for notify := range serverNotification {
			if notify.Authenticated != nil {
				// start client once we get authenticated message
				if err := s.startTamedClient(ctx); err != nil {
					log.Fatalf("Unable to start tamed client: %v\n", err)
				}
				log.Println("Started Tamed Client")
			} else if notify.LoginURL != nil {
				log.Printf("Please login: %v\n", notify.LoginURL)
			}
		}
	}()

	// start a server
	if err := server.Start(ctx, nil, serverNotification); err != nil {
		return err
	}
	return nil
}

func (s *Service) startTamedClient(ctx context.Context) (err error) {
	if s.Client, err = client.Start(ctx, s.Option); err != nil {
		return err
	}
	return nil
}
