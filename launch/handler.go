package launch

import (
	"github.com/DarthSim/overmind/utils"

	"gopkg.in/urfave/cli.v1"
)

// Run runs the launch command
func Run(c *cli.Context) error {
	keepAlive := len(c.Args().Get(4)) > 0

	cmd, err := newCommand(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2), c.Args().Get(3), keepAlive)
	utils.FatalOnErr(err)

	utils.FatalOnErr(cmd.Run())

	return nil
}
