package start

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/DarthSim/overmind/utils"
)

type process struct {
	command   string
	root      string
	sessionID string
	output    *multiOutput
	conn      *processConnection

	Name  string
	Color int
	Pid   string
}

type processesMap map[string]*process

func newProcess(name, sessionID string, color int, command, root string, output *multiOutput) *process {
	return &process{
		command:   command,
		root:      root,
		sessionID: sessionID,
		output:    output,
		Name:      name,
		Color:     color,
	}
}

func (p *process) WindowID() string {
	return fmt.Sprintf("%v:%v", p.sessionID, p.Name)
}

func (p *process) Start(socketPath string, newSession bool) (err error) {
	if p.Running() {
		return
	}

	args := []string{
		"-n", p.Name, "-P", "-F", "#{pane_pid}",
		"-c", p.root, os.Args[0], "launch", p.Name, p.command, socketPath,
		"\\;", "allow-rename", "off",
	}

	if newSession {
		args = append([]string{"new", "-d", "-s", p.sessionID}, args...)
	} else {
		args = append([]string{"neww", "-t", p.sessionID}, args...)
	}

	p.Pid, err = utils.RunCmdOutput("tmux", args...)
	p.Pid = strings.TrimSpace(p.Pid)

	return
}

func (p *process) Wait() {
	for p.Running() {
		time.Sleep(100 * time.Millisecond)
	}
}

func (p *process) Running() bool {
	if len(p.Pid) == 0 {
		return false
	}

	return utils.RunCmd("/bin/sh", "-c", fmt.Sprintf("kill -0 %v", p.Pid)) == nil
}

func (p *process) Stop() {
	if !p.Running() {
		return
	}

	if p.conn != nil {
		p.conn.Stop()
	}
}

func (p *process) Kill() {
	if !p.Running() {
		return
	}

	p.output.WriteBoldLine(p, []byte("Killing..."))
	utils.RunCmd("/bin/sh", "-c", fmt.Sprintf("kill -9 %v", p.Pid))
}

func (p *process) Restart() {
	if p.conn == nil {
		return
	}

	p.conn.Restart()
}

func (p *process) AttachConnection(conn net.Conn) {
	if p.conn != nil {
		return
	}

	p.conn = &processConnection{conn}

	go p.scanConn()
}

func (p *process) scanConn() {
	err := utils.ScanLines(p.conn.Reader(), func(b []byte) bool {
		p.output.WriteLine(p, b)
		return true
	})
	if err != nil {
		p.output.WriteErr(p, fmt.Errorf("Connection error: %v", err))
	}
}
