package start

import (
	"errors"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/DarthSim/overmind/utils"
)

const baseColor = 32

type command struct {
	timeout   int
	output    *multiOutput
	cmdCenter *commandCenter
	doneTrig  chan bool
	doneWg    sync.WaitGroup
	stopTrig  chan os.Signal
	processes processesMap
	hash      string
}

func newCommand(h *Handler) (*command, error) {
	pf := parseProcfile(h.Procfile, h.PortBase, h.PortStep)

	c := command{
		timeout:   h.Timeout,
		doneTrig:  make(chan bool, len(pf)),
		stopTrig:  make(chan os.Signal),
		processes: make(processesMap),
		hash:      utils.RandomString(32),
	}

	root, err := h.AbsRoot()
	if err != nil {
		return nil, err
	}

	c.output = newMultiOutput(pf.MaxNameLength())

	procNames := strings.Split(h.ProcNames, ",")

	for i, e := range pf {
		if len(procNames) == 0 || utils.StringsContain(procNames, e.Name) {
			c.processes[e.Name] = newProcess(e.Name, c.hash, baseColor+i, e.Command, root, c.output)
		}
	}

	c.cmdCenter = newCommandCenter(c.processes, h.SocketPath, c.output)

	return &c, nil
}

func (c *command) Run() error {
	if !c.checkTmux() {
		return errors.New("Can't find tmux. Did you forget to install it?")
	}

	c.startCommandCenter()
	defer c.stopCommandCenter()

	c.runProcesses()

	go c.waitForExit()

	c.doneWg.Wait()

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
	for _, p := range c.processes {
		c.doneWg.Add(1)

		go func(p *process, trig chan bool, wg *sync.WaitGroup) {
			defer wg.Done()
			defer func() { trig <- true }()

			if err := p.Run(c.cmdCenter.SocketPath); err != nil {
				c.output.WriteErr(p, err)
				return
			}
		}(p, c.doneTrig, &c.doneWg)
	}
}

func (c *command) waitForExit() {
	signal.Notify(c.stopTrig, os.Interrupt, os.Kill)

	c.waitForDoneOrStop()

	for {
		for _, proc := range c.processes {
			go proc.Stop()
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
