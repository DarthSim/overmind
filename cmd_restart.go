package main

import (
	"fmt"
	"strings"

	"github.com/DarthSim/overmind/v2/utils"

	"github.com/urfave/cli"
)

type cmdRestartHandler struct{ dialer }

func (h *cmdRestartHandler) Run(c *cli.Context) error {
	conn, err := h.Dial()
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "restart %v\n", strings.Join(c.Args(), " "))

	return nil
}
