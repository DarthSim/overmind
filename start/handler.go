package start

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/DarthSim/overmind/v2/utils"

	"github.com/urfave/cli"
)

var signalMap = map[string]syscall.Signal{
	"ABRT": syscall.SIGABRT,
	"INT":  syscall.SIGINT,
	"KILL": syscall.SIGKILL,
	"QUIT": syscall.SIGQUIT,
	"STOP": syscall.SIGSTOP,
	"TERM": syscall.SIGTERM,
	"USR1": syscall.SIGUSR1,
	"USR2": syscall.SIGUSR2,
}

// Handler handles args and flags for the start command
type Handler struct {
	Title              string
	Procfile           string
	Root               string
	Timeout            int
	PortBase, PortStep int
	ProcNames          string
	SocketPath         string
	SocketName         string
	CanDie             string
	AutoRestart        string
	Colors             []int
	Formation          map[string]int
	FormationPortStep  int
	StopSignals        map[string]syscall.Signal
	Daemonize          bool
	TmuxConfigPath     string
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
	err := h.parseColors(c.String("colors"))
	utils.FatalOnErr(err)

	err = h.parseFormation(c.String("formation"))
	utils.FatalOnErr(err)

	err = h.parseStopSignals(c.String("stop-signals"))
	utils.FatalOnErr(err)

	cmd, err := newCommand(h)
	utils.FatalOnErr(err)

	exitCode, err := cmd.Run()
	utils.FatalOnErr(err)

	os.Exit(exitCode)

	return nil
}

func (h *Handler) parseColors(colorsStr string) error {
	if len(colorsStr) > 0 {
		colors := strings.Split(colorsStr, ",")

		h.Colors = make([]int, len(colors))

		for i, s := range colors {
			color, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil || color < 0 || color > 255 {
				return fmt.Errorf("Invalid xterm color code: %s", s)
			}
			h.Colors[i] = color
		}
	}

	return nil
}

func (h *Handler) parseFormation(formation string) error {
	if len(formation) > 0 {
		maxProcNum := h.PortStep / h.FormationPortStep

		entries := strings.Split(formation, ",")

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

	return nil
}

func (h *Handler) parseStopSignals(signals string) error {
	if len(signals) > 0 {
		entries := strings.Split(signals, ",")

		h.StopSignals = make(map[string]syscall.Signal)

		for _, entry := range entries {
			pair := strings.Split(entry, "=")

			if len(pair) != 2 {
				return errors.New("Invalid stop-signals format")
			}

			name := strings.TrimSpace(pair[0])
			if len(name) == 0 {
				return errors.New("Invalid stop-signals format")
			}

			if signal, ok := signalMap[pair[1]]; ok {
				h.StopSignals[name] = signal
			} else {
				return fmt.Errorf("Invalid signal: %s", pair[1])
			}
		}
	}

	return nil
}
