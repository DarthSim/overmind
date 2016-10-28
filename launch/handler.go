package launch

import (
	"github.com/DarthSim/overmind/utils"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Handler handles args and flags for the launch command
type Handler struct {
	ProcessName string
	CmdLine     string
	SocketPath  string
}

// Run runs the launch command
func (h *Handler) Run(_ *kingpin.ParseContext) error {
	cmd, err := newCommand(h)
	utils.FatalOnErr(err)

	utils.FatalOnErr(cmd.Run())

	return nil
}
