package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/DarthSim/overmind/v2/utils"

	"github.com/urfave/cli"
)

type cmdStopHandler struct {
	SocketPath string
}

func (h *cmdStopHandler) Run(c *cli.Context) error {
	conn, err := net.Dial("unix", h.SocketPath)
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "stop %v\n", strings.Join(c.Args(), " "))

	return nil
}
