package main

import (
	"fmt"
	"net"

	"github.com/DarthSim/overmind/v2/utils"

	"github.com/urfave/cli"
)

type cmdQuitHandler struct {
	SocketPath string
}

func (c *cmdQuitHandler) Run(_ *cli.Context) error {
	conn, err := net.Dial("unix", c.SocketPath)
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "quit")

	return nil
}
