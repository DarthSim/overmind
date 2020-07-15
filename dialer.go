package main

import "net"

type dialer struct {
	SocketPath string
	Network    string
}

func (d *dialer) Dial() (net.Conn, error) {
	return net.Dial(d.Network, d.SocketPath)
}
