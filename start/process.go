package start

import (
	"fmt"
	"io"
	"syscall"
	"time"

	"github.com/DarthSim/overmind/v2/utils"
)

const runningCheckInterval = 100 * time.Millisecond

const SIGINFO syscall.Signal = 29

type process struct {
	output *multiOutput

	pid int

	stopSignal   syscall.Signal
	canDie       bool
	canDieNow    bool
	autoRestart  bool
	keepingAlive bool
	dead         bool
	interrupted  bool
	restart      bool

	tmux *tmuxClient

	in  io.Writer
	out io.ReadCloser

	Name    string
	Color   int
	Command string
}

func newProcess(tmux *tmuxClient, name string, color int, command string, output *multiOutput, canDie bool, autoRestart bool, stopSignal syscall.Signal) *process {
	out, in := io.Pipe()

	proc := &process{
		output: output,
		tmux:   tmux,

		stopSignal:  stopSignal,
		canDie:      canDie,
		canDieNow:   canDie,
		autoRestart: autoRestart,

		in:  in,
		out: out,

		Name:    name,
		Color:   color,
		Command: command,
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

		p.output.WriteBoldLinef(p, "Started with pid %v...", p.pid)

		go p.scanOuput()
		go p.observe()
	}

	p.Wait()
}

func (p *process) Wait() {
	ticker := time.NewTicker(runningCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		if p.dead {
			return
		}
	}
}

func (p *process) Running() bool {
	if pid := p.pid; pid > 0 {
		err := syscall.Kill(pid, syscall.Signal(0))
		return err == nil || err == syscall.EPERM
	}

	return false
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
		if err := p.groupSignal(p.stopSignal); err != nil {
			p.output.WriteErr(p, fmt.Errorf("Can't stop: %s", err))
		}
	}

	p.interrupted = true
}

func (p *process) Kill(keepAlive bool) {
	p.canDieNow = keepAlive

	if p.Running() {
		p.output.WriteBoldLine(p, []byte("Killing..."))
		if err := p.groupSignal(syscall.SIGKILL); err != nil {
			p.output.WriteErr(p, fmt.Errorf("Can't kill: %s", err))
		}
	}
}

func (p *process) Info() {
	p.groupSignal(SIGINFO)
}

func (p *process) Restart() {
	p.restart = true
	p.Stop(false)
}

func (p *process) waitPid() {
	ticker := time.NewTicker(runningCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		if p.pid > 0 {
			break
		}
	}
}

func (p *process) groupSignal(s syscall.Signal) error {
	if pid := p.pid; pid > 0 {
		return syscall.Kill(-pid, s)
	}

	return nil
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
	ticker := time.NewTicker(runningCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		if !p.Running() {
			if !p.keepingAlive {
				p.out.Close()

				p.reportExitCode()
				p.keepingAlive = true
			}

			if !p.canDieNow {
				p.keepingAlive = false
				p.pid = 0

				if p.restart || (!p.interrupted && p.autoRestart) {
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

	p.out, p.in = io.Pipe()
	go p.scanOuput()

	p.tmux.RespawnProcess(p)

	p.waitPid()
	p.output.WriteBoldLinef(p, "Restarted with pid %v...", p.pid)
}

func (p *process) reportExitCode() {
	exitCode := p.tmux.WindowExitCode(p.WindowID())
	message := fmt.Sprintf("Exited with code %d", exitCode)

	p.output.WriteBoldLine(p, []byte(message))
}
