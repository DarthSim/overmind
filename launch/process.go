package launch

import (
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/DarthSim/overmind/term"
	"github.com/DarthSim/overmind/utils"
	"github.com/kr/pty"
)

const runningCheckInterval = 100 * time.Millisecond

type process struct {
	cmd         *exec.Cmd
	writer      writerHelper
	interrupted bool
}

func runProcess(cmdLine string, writer writerHelper, tp term.Params) (*process, error) {
	pty, tty, err := pty.Open()
	if err != nil {
		return nil, err
	}

	if err := term.SetParams(pty, tp); err != nil {
		return nil, err
	}

	proc := process{
		cmd:    exec.Command("/bin/sh", "-c", cmdLine),
		writer: writer,
	}

	go io.Copy(proc.writer, pty)
	go io.Copy(pty, os.Stdin)

	proc.cmd.Stdout = tty
	proc.cmd.Stderr = tty
	proc.cmd.Stdin = tty
	proc.cmd.SysProcAttr = &syscall.SysProcAttr{Setctty: true, Setsid: true}

	if err := proc.cmd.Start(); err != nil {
		proc.writer.WriteErr(utils.ConvertError(err))
		return nil, err
	}

	go func(p *process, pty, tty *os.File) {
		defer pty.Close()
		defer tty.Close()

		if err := p.cmd.Wait(); err != nil {
			p.writer.WriteErr(utils.ConvertError(err))
		}
	}(&proc, pty, tty)

	return &proc, nil
}

func (p *process) Wait() {
	for _ = range time.Tick(runningCheckInterval) {
		if !p.Running() {
			return
		}
	}
}

func (p *process) Running() bool {
	return p.cmd.Process != nil && p.cmd.ProcessState == nil
}

func (p *process) Stop() {
	if p.interrupted {
		// Ok, we tried this easy way, it's time to kill
		p.writer.WriteBoldLine("Killing...")
		p.signal(syscall.SIGKILL)
	} else {
		p.writer.WriteBoldLine("Interrupting...")
		p.signal(syscall.SIGINT)
		p.interrupted = true
	}
}

func (p *process) signal(sig os.Signal) {
	if !p.Running() {
		return
	}

	group, err := os.FindProcess(-p.cmd.Process.Pid)
	if err != nil {
		p.writer.WriteErr(err)
		return
	}

	if err = group.Signal(sig); err != nil {
		p.writer.WriteErr(err)
	}
}
