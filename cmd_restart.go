package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/DarthSim/overmind/utils"

	"gopkg.in/urfave/cli.v1"
)

type cmdRestartHandler struct {
	SocketPath string
}

func (h *cmdRestartHandler) Run(c *cli.Context) error {
	if !c.Args().Present() {
		return cli.NewExitError("Specify names of processes to be restarted", 1)
	}

	conn, err := net.Dial("unix", h.SocketPath)
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "restart %v\n", strings.Join(c.Args(), " "))

	return nil
}
