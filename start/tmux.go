package start

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/DarthSim/overmind/utils"

	"github.com/pkg/term/termios"
)

var tmuxUnescapeRe = regexp.MustCompile(`\\(\d{3})`)
var tmuxOutputRe = regexp.MustCompile("%(\\S+) (.+)")
var tmuxProcessRe = regexp.MustCompile("%(\\d+) (.+) (\\d+)")
var outputRe = regexp.MustCompile("%(\\d+) (.+)")

const tmuxPaneFmt = "%overmind-process #{pane_id} #{window_name} #{pane_pid}"

type tmuxClient struct {
	pty, tty *os.File

	processesByPane processesMap
	processesByName processesMap

	cmdMutex sync.Mutex

	cmd *exec.Cmd

	initKilled bool

	Root    string
	Socket  string
	Session string
}

func newTmuxClient(session, socket, root string) (*tmuxClient, error) {
	t := tmuxClient{
		processesByName: make(processesMap),
		processesByPane: make(processesMap),

		Root:    root,
		Session: session,
		Socket:  socket,
	}

	var err error

	t.pty, t.tty, err = termios.Pty()
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (t *tmuxClient) Start() error {
	go t.listen()

	args := []string{"-CC", "-L", t.Socket}

	first := true
	for name, p := range t.processesByName {
		if first {
			first = false
			args = append(args, "new", "-n", name, "-s", t.Session, "-P", "-F", tmuxPaneFmt, p.Command, ";")
			args = append(args, "setw", "-g", "remain-on-exit", "on", ";")
		} else {
			args = append(args, "neww", "-n", name, "-P", "-F", tmuxPaneFmt, p.Command, ";")
		}
	}

	t.cmd = exec.Command("tmux", args...)
	t.cmd.Stdout = t.tty
	// t.cmd.Stdout = io.MultiWriter(t.tty, os.Stdout)
	t.cmd.Stderr = os.Stderr
	t.cmd.Stdin = t.tty
	t.cmd.SysProcAttr = &syscall.SysProcAttr{Setctty: true, Setsid: true}
	t.cmd.Dir = t.Root

	if err := t.cmd.Start(); err != nil {
		return err
	}

	return nil
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

	utils.FatalOnErr(scanner.Err())
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

func (t *tmuxClient) AddProcess(p *process) {
	t.processesByName[p.Name] = p
}

func (t *tmuxClient) RespawnProcess(p *process) {
	t.sendCmd("neww -k -t %s -n %s -P -F \"%s\" %s", p.Name, p.Name, tmuxPaneFmt, p.Command)
}

func (t *tmuxClient) ExitCode() (status int) {
	buf := new(bytes.Buffer)

	cmd := exec.Command("tmux", "-L", t.Socket, "list-windows", "-t", t.Session, "-F", "#{pane_dead_status}")
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	cmd.Run()

	scanner := bufio.NewScanner(buf)

	for scanner.Scan() {
		if s, err := strconv.Atoi(scanner.Text()); err == nil && s > status {
			status = s
		}
	}

	return
}

func (t *tmuxClient) Shutdown() {
	t.sendCmd("kill-session")

	stopped := make(chan struct{})

	go func() {
		t.cmd.Process.Wait()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
	}
}
