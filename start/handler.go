package start

import (
	"errors"
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
	Formation          map[string]int
	FormationPortStep  int
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
	if strings.ContainsAny(h.Title, ".:") {
		return errors.New("Due to the tmux restrictions, title can't contain . (dot) and : (colon)")
	}

	if len(c.String("colors")) > 0 {
		colors := strings.Split(c.String("colors"), ",")
		h.Colors = make([]int, len(colors))
		for i, s := range colors {
			color, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil || color < 0 || color > 255 {
				return fmt.Errorf("Invalid xterm color code: %s", s)
			}
			h.Colors[i] = color
		}
	}

	if len(c.String("formation")) > 0 {
		maxProcNum := h.PortStep / h.FormationPortStep

		entries := strings.Split(c.String("formation"), ",")
		h.Formation = make(map[string]int)
		for _, entry := range entries {
			pair := strings.Split(entry, "=")

			if len(pair) != 2 {
				return errors.New("Invalid formation format")
			}

			name := strings.TrimSpace(pair[0])
			if len(name) == 0 {
				return errors.New("Invalid formation format")
			}

			num, err := strconv.Atoi(strings.TrimSpace(pair[1]))
			if err != nil || num < 0 {
				return fmt.Errorf("Invalid number of processes: %s", pair[1])
			}
			if num > maxProcNum {
				return fmt.Errorf("You can spawn only %d instances of the same process with port step of %d and formation port step of %d", maxProcNum, h.PortStep, h.FormationPortStep)
			}

			h.Formation[name] = num
		}
	}

	cmd, err := newCommand(h)
	utils.FatalOnErr(err)

	utils.FatalOnErr(cmd.Run())

	return nil
}
