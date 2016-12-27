package start

import (
	"fmt"
	"net"
	"os"
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
}

type processesMap map[string]*process

func newProcess(name, hash string, color int, command, root string, output *multiOutput) *process {
	return &process{
		command:   command,
		root:      root,
		sessionID: fmt.Sprintf("overmind-%v-%v", name, hash),
		output:    output,
		Name:      name,
		Color:     color,
	}
}

func (p *process) Run(socketPath string) error {
	if p.Running() {
		return nil
	}

	p.serviceMsg("Starting in %v...", p.sessionID)

	args := []string{"new", "-d", "-s", p.sessionID, "-n", p.Name, "-c", p.root, os.Args[0], "launch", p.Name, p.command, socketPath}

	if err := utils.RunCmd("tmux", args...); err != nil {
		return err
	}

	p.wait()

	p.serviceMsg("Exited")

	return nil
}

func (p *process) Running() bool {
	if p.conn != nil {
		return !p.conn.Closed
	}

	return p.sessionActive()
}

func (p *process) Stop() {
	if !p.Running() {
		return
	}

	if p.conn != nil {
		p.conn.Stop()
		p.wait()
	}

	// Session should be closed after the process exit, but to be sure...
	p.Kill()
}

func (p *process) Kill() {
	utils.RunCmd("tmux", "kill-session", "-t", p.sessionID)
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

	p.conn = &processConnection{
		conn:   conn,
		Closed: false,
	}

	go p.scanConn()
}

func (p *process) scanConn() {
	err := utils.ScanLines(p.conn.Reader(), func(b []byte) bool {
		p.output.WriteLine(p, b)
		return true
	})
	if err != nil {
		p.serviceMsg("Connection error:", err)
	}
	p.conn.Closed = true
}

func (p *process) sessionActive() bool {
	return utils.RunCmd("tmux", "has", "-t", p.sessionID) == nil
}

func (p *process) serviceMsg(f string, i ...interface{}) {
	p.output.WriteBoldLine(p, []byte(fmt.Sprintf(f, i...)))
}

func (p *process) wait() {
	for p.Running() {
		time.Sleep(100 * time.Millisecond)
	}
}
