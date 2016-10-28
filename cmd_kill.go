package main

import (
	"fmt"
	"net"

	"github.com/DarthSim/overmind/utils"

	"gopkg.in/alecthomas/kingpin.v2"
)

type cmdKillHandler struct {
	SocketPath string
}

func (c *cmdKillHandler) Run(_ *kingpin.ParseContext) error {
	conn, err := net.Dial("unix", c.SocketPath)
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "kill")

	return nil
}
