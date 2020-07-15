package main

import (
	"fmt"
	"strings"

	"github.com/DarthSim/overmind/v2/utils"

	"github.com/urfave/cli"
)

type cmdStopHandler struct{ dialer }

func (h *cmdStopHandler) Run(c *cli.Context) error {
	conn, err := h.Dial()
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "stop %v\n", strings.Join(c.Args(), " "))

	return nil
}
