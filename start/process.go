package start

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/DarthSim/overmind/utils"
)

const runningCheckInterval = 100 * time.Millisecond

type process struct {
	output *multiOutput

	proc      *os.Process
	procGroup *os.Process

	stopSignal   syscall.Signal
	canDie       bool
	canDieNow    bool
	keepingAlive bool
	dead         bool
	interrupted  bool
	restart      bool

	tmux     *tmuxClient
	tmuxPane string

	in  io.Writer
	out io.Reader

	Name    string
	Color   int
	Command string
}

type processesMap map[string]*process

func newProcess(tmux *tmuxClient, name string, color int, command string, port int, output *multiOutput, canDie bool, scriptDir string, stopSignal syscall.Signal) *process {
	out, in := io.Pipe()

	scriptFile, err := os.Create(filepath.Join(scriptDir, name))
	utils.FatalOnErr(err)

	fmt.Fprintln(scriptFile, "#!/bin/sh")
	fmt.Fprintf(scriptFile, "export PORT=%d\n", port)
	fmt.Fprintln(scriptFile, command)

	utils.FatalOnErr(scriptFile.Chmod(0744))

	utils.FatalOnErr(scriptFile.Close())

	proc := &process{
		output: output,
		tmux:   tmux,

		stopSignal: stopSignal,
		canDie:     canDie,
		canDieNow:  canDie,

		in:  in,
		out: out,

		Name:    name,
		Color:   color,
		Command: scriptFile.Name(),
	}

	tmux.AddProcess(proc)

	return proc
}

func (p *process) WindowID() string {
	return fmt.Sprintf("%s:%s", p.tmux.Session, p.Name)
}

func (p *process) StartObserving() {
	if !p.Running() {
		p.waitPid()

		p.output.WriteBoldLinef(p, "Started with pid %v...", p.Pid())

		go p.scanOuput()
		go p.observe()
	}

	p.Wait()
}

func (p *process) Pid() int {
	if p.proc == nil {
		return 0
	}

	return -p.proc.Pid
}

func (p *process) Wait() {
	for range time.Tick(runningCheckInterval) {
		if p.dead {
			return
		}
	}
}

func (p *process) Running() bool {
	return p.proc != nil && p.proc.Signal(syscall.Signal(0)) == nil
}

func (p *process) Stop(keepAlive bool) {
	p.canDieNow = keepAlive

	if p.interrupted {
		// Ok, we have tried once, time to go brutal
		p.Kill(keepAlive)
		return
	}

	if p.Running() {
		p.output.WriteBoldLine(p, []byte("Interrupting..."))
		p.proc.Signal(p.stopSignal)
	}

	p.interrupted = true
}

func (p *process) Kill(keepAlive bool) {
	p.canDieNow = keepAlive

	if p.Running() {
		p.output.WriteBoldLine(p, []byte("Killing..."))
		p.proc.Signal(syscall.SIGKILL)
	}
}

func (p *process) Restart() {
	p.restart = true
	p.Stop(false)
}

func (p *process) waitPid() {
	for range time.Tick(runningCheckInterval) {
		if p.Pid() != 0 {
			break
		}
	}
}

func (p *process) scanOuput() {
	err := utils.ScanLines(p.out, func(b []byte) bool {
		p.output.WriteLine(p, b)
		return true
	})
	if err != nil {
		p.output.WriteErr(p, fmt.Errorf("Output error: %v", err))
	}
}

func (p *process) observe() {
	for range time.Tick(runningCheckInterval) {
		if !p.Running() {
			if !p.keepingAlive {
				p.output.WriteBoldLine(p, []byte("Exited"))
				p.keepingAlive = true
			}

			if !p.canDieNow {
				p.keepingAlive = false
				p.proc = nil

				if p.restart {
					p.respawn()
				} else {
					p.dead = true
					break
				}
			}
		}
	}
}

func (p *process) respawn() {
	p.output.WriteBoldLine(p, []byte("Restarting..."))

	p.restart = false
	p.canDieNow = p.canDie
	p.interrupted = false

	p.tmux.RespawnProcess(p)

	p.waitPid()
}
