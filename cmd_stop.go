package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/DarthSim/overmind/utils"

	"gopkg.in/urfave/cli.v1"
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
