package discover

import (
	"context"
	"github.com/nnanto/tamed/client"
	"github.com/nnanto/tamed/server"
	"log"
)

// startTS starts a tailscale client (and server)
func startTS(ctx context.Context, notificationCh chan client.Notify, startServer bool) (cl *client.Client, err error) {

	if startServer {
		return startTamedService(ctx, notificationCh)
	} else {
		return startTamedClient(ctx, notificationCh)
	}
}

func startTamedService(ctx context.Context, ch chan client.Notify) (cl *client.Client, err error) {
	serverNotification := make(chan server.Notify)

	// start server
	go func() {
		if err := server.Start(ctx, nil, serverNotification); err != nil {
			log.Fatalf("Unable to start tailscale server: %v\n", err)
		}
	}()

	// receive notification from server
	for notify := range serverNotification {
		if notify.Authenticated != nil {
			// start client once we get authenticated message
			if cl, err = startTamedClient(ctx, ch); err != nil {
				log.Fatalf("Unable to start tamed client: %v\n", err)
			}
			log.Println("Started Tamed Client")
			serverNotification = nil
			break
		} else if notify.LoginURL != nil {
			log.Printf("Please login: %v\n", notify.LoginURL)
		}
	}

	return
}

// startTamedClient starts client
func startTamedClient(ctx context.Context, ch chan client.Notify) (*client.Client, error) {
	options := client.DefaultOptions()
	options.ListenerCh = ch
	return client.Start(ctx, options)
}
