package main

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/DarthSim/overmind/v2/utils"
	"github.com/urfave/cli"
)

type cmdRunHandler struct{}

func (h *cmdRunHandler) Run(c *cli.Context) error {
	if !c.Args().Present() {
		utils.Fatal("Specify a command to run")
	}

	command := c.Args()[0]

	var args []string

	if c.NArg() > 1 {
		args = c.Args()[1:]
	} else {
		args = []string{}
	}

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			} else {
				os.Exit(1)
			}
		} else {
			utils.Fatal(err)
		}
	}

	return nil
}
