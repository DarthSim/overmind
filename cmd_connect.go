package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/DarthSim/overmind/utils"

	"gopkg.in/alecthomas/kingpin.v2"
)

type cmdConnectHandler struct {
	ProcessName string
	SocketPath  string
}

func (c *cmdConnectHandler) Run(_ *kingpin.ParseContext) error {
	conn, err := net.Dial("unix", c.SocketPath)
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "get-session %v\n", c.ProcessName)

	sid, err := bufio.NewReader(conn).ReadString('\n')
	utils.FatalOnErr(err)

	// For some reason this doesn't work without sh
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("tmux attach -t %v", sid))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	utils.FatalOnErr(cmd.Run())

	return nil
}
