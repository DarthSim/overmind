package start

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DarthSim/overmind/utils"

	"gopkg.in/urfave/cli.v1"
)

// Handler handles args and flags for the start command
type Handler struct {
	Title              string
	Procfile           string
	Root               string
	Timeout            int
	PortBase, PortStep int
	ProcNames          string
	SocketPath         string
	CanDie             string
	Colors             []int
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
func (h *Handler) Run(c *cli.Context) error {
	colors := strings.Split(c.String("colors"), ",")
	if len(colors) > 0 {
		h.Colors = make([]int, len(colors))
		for i, s := range colors {
			c, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil || c < 0 || c > 255 {
				return fmt.Errorf("Invalid xterm color code: %s", s)
			}
			h.Colors[i] = c
		}
	}

	cmd, err := newCommand(h)
	utils.FatalOnErr(err)

	utils.FatalOnErr(cmd.Run())

	return nil
}
