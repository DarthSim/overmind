package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/DarthSim/overmind/utils"

	"gopkg.in/urfave/cli.v1"
)

type cmdEchoHandler struct {
	ControlMode bool
	SocketPath  string
}

func (h *cmdEchoHandler) Run(c *cli.Context) error {
	if c.Args().Present() {
		utils.Fatal("Echo doesn't accept any arguments")
	}

	conn, err := net.Dial("unix", h.SocketPath)
	utils.FatalOnErr(err)

	stop := make(chan os.Signal)

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
