package launch

import (
	"net"

	"github.com/DarthSim/overmind/utils"
)

type commandCenter struct {
	command *command
	conn    net.Conn
}

func newCommandCenter(command *command, conn net.Conn) *commandCenter {
	return &commandCenter{
		command: command,
		conn:    conn,
	}
}

func (c *commandCenter) Start() {
	utils.ScanLines(c.conn, func(b []byte) bool {
		if c.command.proc == nil {
			return true
		}

		switch string(b) {
		case "stop":
			c.command.Stop()
		case "restart":
			c.command.Restart()
		}

		return true
	})
}
