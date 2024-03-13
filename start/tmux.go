package start

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DarthSim/overmind/v2/utils"
	"golang.org/x/term"
)

var tmuxUnescapeRe = regexp.MustCompile(`\\(\d{3})`)
var tmuxOutputRe = regexp.MustCompile(`%(\S+) (.+)`)
var tmuxProcessRe = regexp.MustCompile(`%(\d+) (.+) (\d+)`)
var outputRe = regexp.MustCompile(`%(\d+) (.+)`)

const tmuxPaneMsgFmt = `%%overmind-process #{pane_id} %s #{pane_pid}`

type tmuxClient struct {
	inReader, outReader io.Reader
	inWriter, outWriter io.Writer

	processes       []*process
	processesByPane map[string]*process

	cmdMutex sync.Mutex

	cmd *exec.Cmd

	configPath string

	shutdown bool

	outputOffset int

	Root    string
	Socket  string
	Session string
}

func newTmuxClient(session, socket, root, configPath string, outputOffset int) *tmuxClient {
	t := tmuxClient{
		processes:       make([]*process, 0),
		processesByPane: make(map[string]*process),

		configPath: configPath,

		outputOffset: outputOffset,

		Root:    root,
		Session: session,
		Socket:  socket,
	}

	t.inReader, t.inWriter = io.Pipe()
	t.outReader, t.outWriter = io.Pipe()

	return &t
}

func (t *tmuxClient) Start() error {
	go t.listen()

	args := []string{"-C", "-L", t.Socket}

	if len(t.configPath) != 0 {
		args = append(args, "-f", t.configPath)
	}

	first := true
	for _, p := range t.processes {
		tmuxPaneMsg := fmt.Sprintf(tmuxPaneMsgFmt, p.Name)

		if first {
			first = false

			args = append(args, "new", "-n", p.Name, "-s", t.Session, "-P", "-F", tmuxPaneMsg, p.Command, ";")

			if w, h, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
				if w > t.outputOffset {
					w -= t.outputOffset
				}

				args = append(args, "refresh", "-C", fmt.Sprintf("%d,%d", w, h), ";")
			}

			args = append(args, "setw", "-g", "remain-on-exit", "on", ";")
			args = append(args, "setw", "-g", "allow-rename", "off", ";")
		} else {
			args = append(args, "neww", "-n", p.Name, "-P", "-F", tmuxPaneMsg, p.Command, ";")
		}
	}

	t.cmd = exec.Command("tmux", args...)
	t.cmd.Stdout = t.outWriter
	t.cmd.Stderr = os.Stderr
	t.cmd.Stdin = t.inReader
	t.cmd.Dir = t.Root

	err := t.cmd.Start()
	if err != nil {
		return err
	}

	go t.observe()

	return nil
}

func (t *tmuxClient) sendCmd(cmd string, arg ...interface{}) {
	t.cmdMutex.Lock()
	defer t.cmdMutex.Unlock()

	fmt.Fprintln(t.inWriter, fmt.Sprintf(cmd, arg...))
}

func (t *tmuxClient) listen() {
	scanner := bufio.NewScanner(t.outReader)

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

func (t *tmuxClient) observe() {
	t.cmd.Process.Wait()

	if t.shutdown {
		return
	}

	exec.Command("tmux", "-L", t.Socket, "kill-session", "-t", t.Session).Run()

	utils.Fatal("Tmux client unexpectedly exited")
}

func (t *tmuxClient) mapProcess(pane, name, pid string) {
	for _, p := range t.processes {
		if p.Name != name {
			continue
		}

		t.processesByPane[pane] = p
		p.paneID = pane // save the tmux paneID in the process

		if ipid, err := strconv.Atoi(pid); err == nil {
			p.pid = ipid
		}

		break
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
	t.processes = append(t.processes, p)
}

func (t *tmuxClient) RespawnProcess(p *process) {
	command := strings.ReplaceAll(fmt.Sprintf("%q", p.Command), "$", "\\$")
	tmuxPaneMsg := fmt.Sprintf(tmuxPaneMsgFmt, p.Name)
	t.sendCmd("neww -d -k -t %s -n %s -P -F %q %s", p.Name, p.Name, tmuxPaneMsg, command)
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
	t.shutdown = true

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

func (t *tmuxClient) PaneExitCode(paneID string) (status int) {
	cmd := exec.Command("tmux", "-L", t.Socket, "list-panes", "-t", paneID, "-F", "#{pane_dead_status}")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	cmd.Run()

	output := strings.TrimSpace(out.String())
	if output == "" {
		return 0
	}

	status, err := strconv.Atoi(output)
	if err != nil {
		utils.Fatal(fmt.Sprintf("Unknown status pane. paneID: %s", paneID))
	}

	return
}
