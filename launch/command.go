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
	socketPath  string
	restart     bool
	writer      writerHelper
	proc        *process
}

func newCommand(procName, cmdLine, socketPath string) (*command, error) {
	return &command{
		processName: procName,
		cmdLine:     cmdLine,
		socketPath:  socketPath,
	}, nil
}

func (c *command) Run() error {
	conn, err := c.establishConn()
	if err != nil {
		return err
	}

	c.writer = writerHelper{io.MultiWriter(conn, os.Stdout)}
	// c.writer = writerHelper{os.Stdout}

	tp, err := term.GetParams(os.Stdin)
	if err != nil {
		return err
	}

	if err := term.MakeRaw(os.Stdin); err != nil {
		return err
	}
	defer term.SetParams(os.Stdin, tp)

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
