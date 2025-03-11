package start

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
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
	infoTrig  chan os.Signal
	processes []*process
	scriptDir string
	daemonize bool
}

func newCommand(h *Handler) (*command, error) {
	pf := parseProcfile(h.Procfile, h.PortBase, h.PortStep, h.Formation, h.FormationPortStep, h.StopSignals)

	c := command{
		timeout:   h.Timeout,
		doneTrig:  make(chan bool, len(pf)),
		stopTrig:  make(chan os.Signal, 1),
		infoTrig:  make(chan os.Signal, 1),
		processes: make([]*process, 0, len(pf)),
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

	c.output = newMultiOutput(pf.MaxNameLength(), h.ShowTimestamps)
	c.tmux = newTmuxClient(session, instanceID, root, h.TmuxConfigPath, c.output.Offset())

	procNames := utils.SplitAndTrim(h.ProcNames)
	ignoredProcNames := utils.SplitAndTrim(h.IgnoredProcNames)

	colors := defaultColors
	if len(h.Colors) > 0 {
		colors = h.Colors
	}

	canDie := utils.SplitAndTrim(h.CanDie)
	autoRestart := utils.SplitAndTrim(h.AutoRestart)

	c.scriptDir = filepath.Join(os.TempDir(), instanceID)
	os.MkdirAll(c.scriptDir, 0700)

	for i, e := range pf {
		shouldRun := len(procNames) == 0 || utils.StringsContain(procNames, e.OrigName)
		isIgnored := len(ignoredProcNames) != 0 && utils.StringsContain(ignoredProcNames, e.OrigName)

		if shouldRun && !isIgnored {
			scriptFilePath := c.createScriptFile(&e, pf, h.Shell, !h.NoPort)

			c.processes = append(c.processes, newProcess(
				c.tmux,
				e.Name,
				colors[i%len(colors)],
				scriptFilePath,
				c.output,
				(h.AnyCanDie || utils.StringsContain(canDie, e.OrigName)),
				(utils.StringsContain(autoRestart, e.OrigName) || utils.StringsContain(autoRestart, "all")),
				e.StopSignal,
			))
		}
	}

	if len(c.processes) == 0 {
		return nil, errors.New("No processes to run")
	}

	c.cmdCenter = newCommandCenter(&c, h.SocketPath, h.Network)

	return &c, nil
}

func (c *command) Run() (int, error) {
	defer os.RemoveAll(c.scriptDir)
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

	go c.handleInfo()

	c.doneWg.Wait()

	exitCode := c.tmux.ExitCode()

	time.Sleep(time.Second)

	c.tmux.Shutdown()

	return exitCode, nil
}

func (c *command) Quit() {
	c.stopTrig <- syscall.SIGINT
}

func (c *command) createScriptFile(e *procfileEntry, procFile procfile, shell string, setPort bool) string {
	scriptFile, err := os.Create(filepath.Join(c.scriptDir, e.Name))
	utils.FatalOnErr(err)

	fmt.Fprintf(scriptFile, "#!/usr/bin/env %s\n", shell)
	if setPort {
		fmt.Fprintf(scriptFile, "export PORT=%d\n", e.Port)

		for _, pf := range procFile {
			if pf.Name != e.Name {
				safeProcessName := sanitizeProcName(pf.Name)
				fmt.Fprintf(scriptFile, "export OVERMIND_PROCESS_%s_PORT=%d\n", safeProcessName, pf.Port)
			}
		}
	}

	fmt.Fprintf(scriptFile, "export PS=%s\n", e.Name)
	fmt.Fprintln(scriptFile, e.Command)

	utils.FatalOnErr(scriptFile.Chmod(0744))

	utils.FatalOnErr(scriptFile.Close())

	return scriptFile.Name()
}

func (c *command) checkTmux() bool {
	return utils.RunCmd("tmux", "-V") == nil
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
	signal.Notify(c.stopTrig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	c.waitForDoneOrStop()

	for _, proc := range c.processes {
		proc.Stop(false)
	}

	c.waitForTimeoutOrStop()

	for _, proc := range c.processes {
		proc.Kill(false)
	}
}

func (c *command) handleInfo() {
	signal.Notify(c.infoTrig, SIGINFO)

	for range c.infoTrig {
		for _, proc := range c.processes {
			proc.Info()
		}
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

func sanitizeProcName(name string) string {
	return regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(name, "_")
}
