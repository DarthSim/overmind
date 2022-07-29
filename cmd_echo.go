package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/DarthSim/overmind/v2/utils"

	"github.com/urfave/cli"
)

type cmdEchoHandler struct{ dialer }

func (h *cmdEchoHandler) Run(c *cli.Context) error {
	if c.Args().Present() {
		utils.Fatal("Echo doesn't accept any arguments")
	}

	conn, err := h.Dial()
	utils.FatalOnErr(err)

	stop := make(chan os.Signal, 1)

	go func() {
		utils.ScanLines(conn, func(b []byte) bool {
			fmt.Fprintf(os.Stdout, "%s\n", b)
			return true
		})

		stop <- syscall.SIGINT
	}()

	fmt.Fprintln(conn, "echo")

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	return nil
}
