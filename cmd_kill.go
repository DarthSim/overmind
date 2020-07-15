package main

import (
	"fmt"

	"github.com/DarthSim/overmind/v2/utils"

	"github.com/urfave/cli"
)

type cmdKillHandler struct{ dialer }

func (h *cmdKillHandler) Run(_ *cli.Context) error {
	conn, err := h.Dial()
	utils.FatalOnErr(err)

	fmt.Fprintf(conn, "kill")

	return nil
}
