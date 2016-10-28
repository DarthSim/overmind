package launch

import (
	"fmt"
	"io"
	"net"
	"os"
)

type command struct {
	processName string
	cmdLine     string
	socketPath  string
	restart     bool
	writer      writerHelper
	proc        *process
}

func newCommand(h *Handler) (*command, error) {
	return &command{
		processName: h.ProcessName,
		cmdLine:     h.CmdLine,
		socketPath:  h.SocketPath,
	}, nil
}

func (c *command) Run() error {
	conn, err := c.establishConn()
	if err != nil {
		return err
	}

	c.writer = writerHelper{io.MultiWriter(conn, os.Stdout)}

	tp, err := getTermParams(os.Stdin)
	if err != nil {
		return err
	}

	t, err := rawTerm()
	if err != nil {
		return err
	}
	defer closeTerm(t)

	for {
		if c.proc, err = runProcess(c.cmdLine, c.writer, tp); err != nil {
			return err
		}

		c.proc.Wait()

		if !c.restart {
			break
		}

		c.restart = false
		c.writer.WriteBoldLine("Restarting...")
	}

	return nil
}

func (c *command) Stop() {
	c.proc.Stop()
}

func (c *command) Restart() {
	c.restart = true
	c.Stop()
}

func (c *command) establishConn() (net.Conn, error) {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return nil, err
	}

	go newCommandCenter(c, conn).Start()

	fmt.Fprintf(conn, "attach %v\n", c.processName)

	return conn, nil
}
