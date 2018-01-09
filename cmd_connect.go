package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

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

	fmt.Fprintf(conn, "get-connection %v\n", c.Args().First())

	response, err := bufio.NewReader(conn).ReadString('\n')
	utils.FatalOnErr(err)

	response = strings.TrimSpace(response)
	if response == "" {
		utils.Fatal(fmt.Sprintf("Unknown process name: %s", c.Args().First()))
	}

	parts := strings.Split(response, " ")
	if len(parts) < 2 {
		utils.Fatal("Invalid server response")
	}

	cmd := exec.Command("tmux", "-L", parts[0], "attach", "-t", parts[1])
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	utils.FatalOnErr(cmd.Run())

	return nil
}
