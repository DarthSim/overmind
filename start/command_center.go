package start

import (
	"fmt"
	"net"
	"regexp"

	"github.com/DarthSim/overmind/utils"
)

type commandCenter struct {
	processes processesMap
	output    *multiOutput
	listener  net.Listener
	stop      bool

	SocketPath string
}

func newCommandCenter(processes processesMap, socket string, output *multiOutput) *commandCenter {
	return &commandCenter{
		processes: processes,
		output:    output,

		SocketPath: socket,
	}
}

func (c *commandCenter) Start() (err error) {
	if c.listener, err = net.Listen("unix", c.SocketPath); err != nil {
		return
	}

	c.serviceMsg("Command center opened at %v", c.SocketPath)

	go func(c *commandCenter) {
		for {
			conn, _ := c.listener.Accept()
			go c.handleConnection(conn)

			if c.stop {
				break
			}
		}
	}(c)

	return nil
}

func (c *commandCenter) Stop() {
	c.stop = true
	c.listener.Close()
	c.serviceMsg("Command center closed")
}

func (c *commandCenter) handleConnection(conn net.Conn) {
	re := regexp.MustCompile("\\S+")

	utils.ScanLines(conn, func(b []byte) bool {
		args := re.FindAllString(string(b), -1)

		if len(args) == 0 {
			return true
		}

		cmd := args[0]

		if len(args) > 1 {
			args = args[1:]
		} else {
			args = []string{}
		}

		switch cmd {
		case "attach":
			c.processAttach(cmd, args, conn)
			return false
		case "restart":
			c.processRestart(cmd, args)
		case "kill":
			c.processKill()
		case "get-session":
			c.processGetSession(cmd, args, conn)
		}

		return true
	})
}

func (c *commandCenter) processAttach(cmd string, args []string, conn net.Conn) {
	if len(args) > 0 {
		if proc, ok := c.processes[args[0]]; ok {
			proc.AttachConnection(conn)
		}
	}
}

func (c *commandCenter) processRestart(cmd string, args []string) {
	for _, n := range args {
		if p, ok := c.processes[n]; ok {
			p.Restart()
		}
	}
}

func (c *commandCenter) processKill() {
	for _, p := range c.processes {
		p.Kill()
	}
}

func (c *commandCenter) serviceMsg(f string, i ...interface{}) {
	c.output.WriteBoldLine(nil, []byte(fmt.Sprintf(f, i...)))
}

func (c *commandCenter) processGetSession(cmd string, args []string, conn net.Conn) {
	if len(args) > 0 {
		if proc, ok := c.processes[args[0]]; ok {
			fmt.Fprintln(conn, proc.sessionID)
		} else {
			fmt.Fprintf(conn, "Unknown process: %v\n", args[0])
		}
	}
}
