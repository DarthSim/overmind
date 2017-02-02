package start

import (
	"fmt"
	"net"
	"path/filepath"
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

func newCommandCenter(processes processesMap, socket string, output *multiOutput) (*commandCenter, error) {
	s, err := filepath.Abs(socket)

	return &commandCenter{
		processes: processes,
		output:    output,

		SocketPath: s,
	}, err
}

func (c *commandCenter) Start() (err error) {
	if c.listener, err = net.Listen("unix", c.SocketPath); err != nil {
		return
	}

	c.output.WriteBoldLinef(nil, "Listening at %v", c.SocketPath)

	go func(c *commandCenter) {
		for {
			if conn, err := c.listener.Accept(); err == nil {
				go c.handleConnection(conn)
			}

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
		case "get-window":
			c.processGetWindow(cmd, args, conn)
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

func (c *commandCenter) processGetWindow(cmd string, args []string, conn net.Conn) {
	if len(args) > 0 {
		if proc, ok := c.processes[args[0]]; ok {
			fmt.Fprintln(conn, proc.WindowID())
		} else {
			fmt.Fprintf(conn, "Unknown process: %v\n", args[0])
		}
	}
}
