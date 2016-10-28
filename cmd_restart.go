package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/DarthSim/overmind/utils"

	"gopkg.in/alecthomas/kingpin.v2"
)

type cmdRestartHandler struct {
	ProcessNames []string
	SocketPath   string
}

func (c *cmdRestartHandler) Run(_ *kingpin.ParseContext) error {
	conn, err := net.Dial("unix", c.SocketPath)
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "restart %v\n", strings.Join(c.ProcessNames, " "))

	return nil
}
