package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/DarthSim/overmind/v2/utils"

	"github.com/urfave/cli"
)

type cmdConnectHandler struct {
	dialer

	ControlMode bool
}

func (h *cmdConnectHandler) Run(c *cli.Context) error {
	if c.NArg() > 1 {
		utils.Fatal("Specify a single name of process")
	}

	conn, err := h.Dial()
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "get-connection %s\n", c.Args().First())

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

	args := []string{"-L", parts[0], "attach", "-t", parts[1]}

	if h.ControlMode {
		args = append([]string{"-CC"}, args...)
	}

	cmd := exec.Command("tmux", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	utils.FatalOnErr(cmd.Run())

	return nil
}
