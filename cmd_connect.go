package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/DarthSim/overmind/utils"

	"gopkg.in/urfave/cli.v1"
)

type cmdConnectHandler struct {
	ProcessName string
	SocketPath  string
}

func (h *cmdConnectHandler) Run(c *cli.Context) error {
	if !c.Args().Present() {
		return cli.NewExitError("Specify name of process to connect", 1)
	}

	if c.NArg() > 1 {
		return cli.NewExitError("Specify a single name of processe", 1)
	}

	conn, err := net.Dial("unix", h.SocketPath)
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "get-session %v\n", c.Args().First())

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
