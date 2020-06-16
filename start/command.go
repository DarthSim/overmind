package start

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/DarthSim/overmind/v2/utils"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/sevlyar/go-daemon"
)

var defaultColors = []int{2, 3, 4, 5, 6, 42, 130, 103, 129, 108}

type command struct {
	title     string
	timeout   int
	tmux      *tmuxClient
	output    *multiOutput
	cmdCenter *commandCenter
	doneTrig  chan bool
	doneWg    sync.WaitGroup
	stopTrig  chan os.Signal
	processes processesMap
	daemonize bool
}

func newCommand(h *Handler) (*command, error) {
	procs := resolveProcs(h)

	c := command{
		timeout:   h.Timeout,
		doneTrig:  make(chan bool, len(procs)),
		stopTrig:  make(chan os.Signal),
		processes: make(processesMap),
		daemonize: h.Daemonize,
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

	session := utils.EscapeTitle(c.title)
	nanoid, err := gonanoid.Nanoid()
	if err != nil {
		return nil, err
	}

	instanceID := fmt.Sprintf("overmind-%s-%s", session, nanoid)

	c.tmux = newTmuxClient(session, instanceID, root, h.TmuxConfigPath)
	c.output = newMultiOutput(procs.MaxNameLength())

	procNames := utils.SplitAndTrim(h.ProcNames)

	colors := defaultColors
	if len(h.Colors) > 0 {
		colors = h.Colors
	}

	canDie := utils.SplitAndTrim(h.CanDie)
	autoRestart := utils.SplitAndTrim(h.AutoRestart)

	for i, e := range procs {
		dir, err := filepath.Abs(filepath.Dir(e.Procfile))
		if err != nil {
			utils.FatalOnErr(err)
		}
		if len(procNames) == 0 || utils.StringsContain(procNames, e.OrigName) {
			c.processes[e.Name] = newProcess(
				c.tmux,
				e.Name,
				colors[i%len(colors)],
				e.Command,
				dir,
				e.Port,
				c.output,
				utils.StringsContain(canDie, e.OrigName),
				utils.StringsContain(autoRestart, e.OrigName),
				e.StopSignal,
			)
		}
	}

	c.cmdCenter = newCommandCenter(&c, h.SocketPath)

	return &c, nil
}

func (c *command) Run() (int, error) {
	defer c.output.Stop()

	fmt.Printf("\033]0;%s | overmind\007", c.title)

	if !c.checkTmux() {
		return 1, errors.New("Can't find tmux. Did you forget to install it?")
	}

	c.output.WriteBoldLinef(nil, "Tmux socket name: %v", c.tmux.Socket)
	c.output.WriteBoldLinef(nil, "Tmux session ID: %v", c.tmux.Session)
	c.output.WriteBoldLinef(nil, "Listening at %v", c.cmdCenter.SocketPath)

	c.startCommandCenter()
	defer c.stopCommandCenter()

	if c.daemonize {
		if !daemon.WasReborn() {
			c.stopCommandCenter()
		}

		ctx := new(daemon.Context)
		child, err := ctx.Reborn()

		if child != nil {
			c.output.WriteBoldLinef(nil, "Daemonized. Use `overmind echo` to view logs and `overmind quit` to gracefully quit daemonized instance")
			return 0, err
		}

		defer ctx.Release()
	}

	c.runProcesses()

	go c.waitForExit()

	c.doneWg.Wait()

	exitCode := c.tmux.ExitCode()

	time.Sleep(time.Second)

	c.tmux.Shutdown()

	return exitCode, nil
}

func (c *command) Quit() {
	c.stopTrig <- syscall.SIGINT
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

			p.StartObserving()
		}(p, c.doneTrig, &c.doneWg)
	}

	utils.FatalOnErr(c.tmux.Start())
}

func (c *command) waitForExit() {
	signal.Notify(c.stopTrig, syscall.SIGINT, syscall.SIGTERM)

	c.waitForDoneOrStop()

	for _, proc := range c.processes {
		proc.Stop(false)
	}

	c.waitForTimeoutOrStop()

	for _, proc := range c.processes {
		proc.Kill(false)
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
