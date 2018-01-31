package start

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"syscall"

	"github.com/pkg/term/termios"
)

var tmuxUnescapeRe = regexp.MustCompile(`\\(\d{3})`)
var tmuxOutputRe = regexp.MustCompile("%(\\S+) (.+)")
var tmuxProcessRe = regexp.MustCompile("%(\\d+) (.+) (\\d+)")
var outputRe = regexp.MustCompile("%(\\d+) (.+)")

const tmuxPanesCmd = "list-windows -F \"%%overmind-process #{pane_id} #{window_name} #{pane_pid}\""

type tmuxClient struct {
	pty *os.File

	processesByPane processesMap
	processesByName processesMap

	cmdMutex sync.Mutex

	initKilled bool

	Socket  string
	Session string
}

func newTmuxClient(session, socket, root string) (*tmuxClient, error) {
	pty, tty, err := termios.Pty()
	if err != nil {
		return nil, err
	}

	t := tmuxClient{
		pty:             pty,
		processesByName: make(processesMap),
		processesByPane: make(processesMap),

		Session: session,
		Socket:  socket,
	}

	cmd := exec.Command("tmux", "-CC", "-L", socket, "new", "-n", "__init__", "-s", session)
	cmd.Stdout = tty
	// cmd.Stdout = io.MultiWriter(tty, os.Stdout)
	cmd.Stderr = os.Stderr
	cmd.Stdin = tty
	cmd.SysProcAttr = &syscall.SysProcAttr{Setctty: true, Setsid: true}
	cmd.Dir = root

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	t.sendCmd("setw -g remain-on-exit on")

	go t.listen()

	return &t, nil
}

func (t *tmuxClient) sendCmd(cmd string, arg ...interface{}) {
	t.cmdMutex.Lock()
	defer t.cmdMutex.Unlock()

	fmt.Fprintln(t.pty, fmt.Sprintf(cmd, arg...))
}

func (t *tmuxClient) listen() {
	scanner := bufio.NewScanner(t.pty)

	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		tmuxOut := tmuxOutputRe.FindStringSubmatch(scanner.Text())

		if len(tmuxOut) < 2 {
			continue
		}

		switch tmuxOut[1] {
		case "overmind-process":
			procbind := tmuxProcessRe.FindStringSubmatch(tmuxOut[2])
			if len(procbind) > 3 {
				t.mapProcess(procbind[1], procbind[2], procbind[3])
			}
		case "output":
			output := outputRe.FindStringSubmatch(tmuxOut[2])
			if len(output) > 2 {
				t.sendOutput(output[1], output[2])
			}
		}
	}
}

func (t *tmuxClient) mapProcess(pane, name, pid string) {
	if p, ok := t.processesByName[name]; ok {
		t.processesByPane[pane] = p
		p.tmuxPane = pane

		if ipid, err := strconv.Atoi(pid); err == nil {
			p.proc, _ = os.FindProcess(-ipid)
		}
	}
}

func (t *tmuxClient) sendOutput(name, str string) {
	if proc, ok := t.processesByPane[name]; ok {
		unescaped := tmuxUnescapeRe.ReplaceAllStringFunc(str, func(src string) string {
			code, _ := strconv.ParseUint(src[1:], 8, 8)
			return string([]byte{byte(code)})
		})

		fmt.Fprint(proc.in, unescaped)
	}
}

func (t *tmuxClient) listPanes() {
	t.sendCmd(tmuxPanesCmd)
}

func (t *tmuxClient) AddProcess(p *process) {
	t.processesByName[p.Name] = p
	t.sendCmd("neww -n %s %s", p.Name, p.Command)

	if !t.initKilled {
		// Ok, we have at least one process running, we can kill __init__ window
		t.sendCmd("movew -s %s -t __init__", p.Name)
		t.sendCmd("killw -t __init__")
		t.initKilled = true
	}

	t.listPanes()
}

func (t *tmuxClient) RespawnProcess(p *process) {
	t.sendCmd("respawn-pane -t %s %s", p.Name, p.Command)
	t.listPanes()
}
