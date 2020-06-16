package start

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/DarthSim/overmind/v2/utils"
)

const runningCheckInterval = 100 * time.Millisecond

type process struct {
	output *multiOutput

	proc *os.Process

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
	out io.Reader

	Name    string
	Color   int
	Command string
	Dir     string
}

type processesMap map[string]*process

func newProcess(tmux *tmuxClient, name string, color int, command, dir string, port int, output *multiOutput, canDie bool, autoRestart bool, stopSignal syscall.Signal) *process {
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
		Command: fmt.Sprintf("export PORT=%d; %s", port, command),
		Dir:     dir,
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
	ticker := time.NewTicker(runningCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
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
	ticker := time.NewTicker(runningCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
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
	ticker := time.NewTicker(runningCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		if !p.Running() {
			if !p.keepingAlive {
				p.output.WriteBoldLine(p, []byte("Exited"))
				p.keepingAlive = true
			}

			if !p.canDieNow {
				p.keepingAlive = false
				p.proc = nil

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

	p.tmux.RespawnProcess(p)

	p.waitPid()
}
