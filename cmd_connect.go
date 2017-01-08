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
		utils.Fatal("Specify a name of process to connect")
	}

	if c.NArg() > 1 {
		utils.Fatal("Specify a single name of process")
	}

	conn, err := net.Dial("unix", h.SocketPath)
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "get-window %v\n", c.Args().First())

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
