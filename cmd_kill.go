package main

import (
	"fmt"
	"net"

	"github.com/DarthSim/overmind/utils"

	"github.com/urfave/cli"
)

type cmdKillHandler struct {
	SocketPath string
}

func (c *cmdKillHandler) Run(_ *cli.Context) error {
	conn, err := net.Dial("unix", c.SocketPath)
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "kill")

	return nil
}
