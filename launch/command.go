package launch

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/DarthSim/overmind/term"
)

type command struct {
	processName string
	cmdLine     string
	port        string
	socketPath  string
	keep_alive  bool
	restart     bool
	writer      writerHelper
	proc        *process
}

func newCommand(procName, cmdLine, port, socketPath string, keep_alive bool) (*command, error) {
	return &command{
		processName: procName,
		cmdLine:     cmdLine,
		port:        port,
		socketPath:  socketPath,
		keep_alive:  keep_alive,
	}, nil
}

func (c *command) Run() error {
	conn, err := c.establishConn()
	if err != nil {
		return err
	}

	c.writer = writerHelper{io.MultiWriter(conn, os.Stdout)}

	tp, err := term.GetParams(os.Stdin)
	if err != nil {
		return err
	}

	if err = term.MakeRaw(os.Stdin); err != nil {
		return err
	}
	defer term.SetParams(os.Stdin, tp)

	os.Setenv("PORT", c.port)

	for {
		if c.proc, err = runProcess(c.cmdLine, c.writer, tp, c.keep_alive); err != nil {
			return err
		}

		c.proc.Wait()

		c.writer.WriteBoldLine("Exited")

		c.proc.WaitKeepAlive()

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
