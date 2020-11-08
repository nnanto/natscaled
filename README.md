# NATScaled

NAT scaled server forms a cluster of NAT servers using TailScale to identify other remote peer servers. In addition to providing discovery mechanism, 
tailscale also provides secure data transfer.

## Building and Running

Run `make build` to generate binary

Usage: `./natserver.go` (*when running the server with **-tsd** , run it with **sudo** command*)

```
  -serverName string name of the server
  -tsd  start tailscale daemon 
  -v    enable logging
```
