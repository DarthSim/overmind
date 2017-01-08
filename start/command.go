package start

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
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
	sessionID string
	canDie    []string
}

func newCommand(h *Handler) (*command, error) {
	pf := parseProcfile(h.Procfile, h.PortBase, h.PortStep)

	c := command{
		timeout:   h.Timeout,
		doneTrig:  make(chan bool, len(pf)),
		stopTrig:  make(chan os.Signal),
		processes: make(processesMap),
		sessionID: fmt.Sprintf("overmind-%v", utils.RandomString(32)),
	}

	root, err := h.AbsRoot()
	if err != nil {
		return nil, err
	}

	c.output = newMultiOutput(pf.MaxNameLength())

	procNames := utils.SplitAndTrim(h.ProcNames)

	for i, e := range pf {
		if len(procNames) == 0 || utils.StringsContain(procNames, e.Name) {
			c.processes[e.Name] = newProcess(e.Name, c.sessionID, baseColor+i, e.Command, root, c.output)
		}
	}

	c.canDie = utils.SplitAndTrim(h.CanDie)

	c.cmdCenter, err = newCommandCenter(c.processes, h.SocketPath, c.output)
	if err != nil {
		return nil, err
	}

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
	c.output.WriteBoldLinef(nil, "Tmux session ID: %v", c.sessionID)

	newSession := true

	for _, p := range c.processes {
		if err := p.Start(c.cmdCenter.SocketPath, newSession); err != nil {
			c.output.WriteErr(p, err)
			c.doneTrig <- true
			break
		}

		newSession = false

		c.output.WriteBoldLinef(p, "Started with pid %v...", p.Pid)
		c.doneWg.Add(1)

		go func(p *process, trig chan bool, wg *sync.WaitGroup) {
			defer wg.Done()
			defer func() {
				if !utils.StringsContain(c.canDie, p.Name) {
					trig <- true
				}
			}()

			p.Wait()

			c.output.WriteBoldLine(p, []byte("Exited"))
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
	utils.RunCmd("tmux", "kill-session", "-t", c.sessionID)
}
