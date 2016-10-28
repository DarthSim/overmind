package launch

import (
	"bufio"
	"net"
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
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		if c.command.proc == nil {
			continue
		}

		cmd := scanner.Text()

		switch cmd {
		case "stop":
			c.command.Stop()
		case "restart":
			c.command.Restart()
		}
	}
}
