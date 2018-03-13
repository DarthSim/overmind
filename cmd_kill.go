package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/DarthSim/overmind/utils"

	"gopkg.in/urfave/cli.v1"
)

type cmdKillHandler struct {
	SocketPath string
}

func (k *cmdKillHandler) Run(c *cli.Context) error {
	conn, err := net.Dial("unix", k.SocketPath)
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "kill %v\n", strings.Join(c.Args(), " "))

	return nil
}
