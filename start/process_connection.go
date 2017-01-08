package start

import (
	"fmt"
	"io"
	"net"
)

type processConnection struct {
	conn net.Conn
}

func (c *processConnection) Reader() io.Reader {
	return io.Reader(c.conn)
}

func (c *processConnection) Stop() {
	fmt.Fprintln(c.conn, "stop")
}

func (c *processConnection) Restart() {
	fmt.Fprintln(c.conn, "restart")
}
