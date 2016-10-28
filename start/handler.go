package start

import (
	"path/filepath"

	"github.com/DarthSim/overmind/utils"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Handler handles args and flags for the start command
type Handler struct {
	Procfile           string
	Root               string
	Timeout            int
	PortBase, PortStep int
	ProcNames          string
	SocketPath         string
}

// AbsRoot returns absolute path to the working directory
func (h *Handler) AbsRoot() (string, error) {
	var absRoot string

	if len(h.Root) > 0 {
		absRoot = h.Root
	} else {
		absRoot = filepath.Dir(h.Procfile)
	}

	return filepath.Abs(absRoot)
}

// Run runs the start command
func (h *Handler) Run(_ *kingpin.ParseContext) error {
	cmd, err := newCommand(h)
	utils.FatalOnErr(err)

	utils.FatalOnErr(cmd.Run())

	return nil
}
