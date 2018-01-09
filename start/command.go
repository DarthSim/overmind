package start

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/DarthSim/overmind/utils"
	gonanoid "github.com/matoous/go-nanoid"
)

var defaultColors = []int{2, 3, 4, 5, 6, 42, 130, 103, 129, 108}

type command struct {
	title       string
	timeout     int
	output      *multiOutput
	cmdCenter   *commandCenter
	doneTrig    chan bool
	doneWg      sync.WaitGroup
	stopTrig    chan os.Signal
	processes   processesMap
	tmuxSocket  string
	tmuxSession string
	canDie      []string
}

func newCommand(h *Handler) (*command, error) {
	pf := parseProcfile(h.Procfile, h.PortBase, h.PortStep)

	c := command{
		timeout:   h.Timeout,
		doneTrig:  make(chan bool, len(pf)),
		stopTrig:  make(chan os.Signal),
		processes: make(processesMap),
	}

	root, err := h.AbsRoot()
	if err != nil {
		return nil, err
	}

	if len(h.Title) > 0 {
		c.title = h.Title
	} else {
		c.title = filepath.Base(root)
	}

	nanoid, err := gonanoid.Nanoid()
	if err != nil {
		return nil, err
	}

	c.tmuxSession = utils.EscapeTitle(c.title)
	c.tmuxSocket = fmt.Sprintf("overmind-%s-%s", c.tmuxSession, nanoid)

	c.output = newMultiOutput(pf.MaxNameLength())

	procNames := utils.SplitAndTrim(h.ProcNames)

	colors := defaultColors
	if len(h.Colors) > 0 {
		colors = h.Colors
	}

	c.canDie = utils.SplitAndTrim(h.CanDie)

	for i, e := range pf {
		if len(procNames) == 0 || utils.StringsContain(procNames, e.Name) {
			c.processes[e.Name] = newProcess(e.Name, c.tmuxSocket, c.tmuxSession, colors[i%len(colors)], e.Command, root, e.Port, c.output, utils.StringsContain(c.canDie, e.Name))
		}
	}

	c.cmdCenter, err = newCommandCenter(c.processes, h.SocketPath, c.output)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *command) Run() error {
	fmt.Printf("\033]0;%s | overmind\007", c.title)

	if !c.checkTmux() {
		return errors.New("Can't find tmux. Did you forget to install it?")
	}

	c.startCommandCenter()
	defer c.stopCommandCenter()

	c.runProcesses()

	go c.waitForExit()

	c.doneWg.Wait()

	// Session should be killed after all windows exit, but just for sure...
	c.killSession()

	return nil
}

func (c *command) checkTmux() bool {
	return utils.RunCmd("which", "tmux") == nil
}

func (c *command) startCommandCenter() {
	utils.FatalOnErr(c.cmdCenter.Start())
}

func (c *command) stopCommandCenter() {
	c.cmdCenter.Stop()
}

func (c *command) runProcesses() {
	c.output.WriteBoldLinef(nil, "Tmux socket name: %v", c.tmuxSocket)
	c.output.WriteBoldLinef(nil, "Tmux session ID: %v", c.tmuxSession)

	newSession := true

	for _, p := range c.processes {
		if err := p.Start(c.cmdCenter.SocketPath, newSession); err != nil {
			c.output.WriteErr(p, err)
			c.doneTrig <- true
			break
		}

		newSession = false

		c.output.WriteBoldLinef(p, "Started with pid %v...", p.Pid())
		c.doneWg.Add(1)

		go func(p *process, trig chan bool, wg *sync.WaitGroup) {
			defer wg.Done()
			defer func() { trig <- true }()

			p.Wait()
		}(p, c.doneTrig, &c.doneWg)
	}
}

func (c *command) waitForExit() {
	signal.Notify(c.stopTrig, os.Interrupt, os.Kill)

	c.waitForDoneOrStop()

	for {
		for _, proc := range c.processes {
			proc.Stop()
		}

		c.waitForTimeoutOrStop()
	}
}

func (c *command) waitForDoneOrStop() {
	select {
	case <-c.doneTrig:
	case <-c.stopTrig:
	}
}

func (c *command) waitForTimeoutOrStop() {
	select {
	case <-time.After(time.Duration(c.timeout) * time.Second):
	case <-c.stopTrig:
	}
}

func (c *command) killSession() {
	utils.RunCmd("tmux", "-L", c.tmuxSocket, "kill-session", "-t", c.tmuxSession)
}
