package main

import (
	"fmt"
	"os"

	"github.com/DarthSim/overmind/v2/utils"

	"github.com/urfave/cli"
)

type cmdStatusHandler struct{ dialer }

func (h *cmdStatusHandler) Run(c *cli.Context) error {
	if c.Args().Present() {
		utils.Fatal("Status doesn't accept any arguments")
	}

	conn, err := h.Dial()
	utils.FatalOnErr(err)

	fmt.Fprintln(conn, "status")

	utils.ScanLines(conn, func(b []byte) bool {
		fmt.Fprintf(os.Stdout, "%s\n", b)
		return true
	})

	return nil
}
