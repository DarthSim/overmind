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

	"golang.org/x/crypto/ssh/terminal"
)

var tmuxVersionRe = regexp.MustCompile(`(\d+)\.(\d+)`)
var tmuxUnescapeRe = regexp.MustCompile(`\\(\d{3})`)
var tmuxOutputRe = regexp.MustCompile(`%(\S+) (.+)`)
var tmuxProcessRe = regexp.MustCompile(`%(\d+) (.+) (\d+)`)
var outputRe = regexp.MustCompile(`%(\d+) (.+)`)

const tmuxPaneFmt = "%overmind-process #{pane_id} #{window_name} #{pane_pid}"

type tmuxClient struct {
	inReader, outReader io.Reader
	inWriter, outWriter io.Writer

	processes       []*process
	processesByPane processesMap

	cmdMutex sync.Mutex

	cmd *exec.Cmd

	configPath string

	shutdown bool

	Root    string
	Socket  string
	Session string
}

func tmuxVersion() (int, int) {
	output, err := exec.Command("tmux", "-V").Output()
	if err != nil {
		return 0, 0
	}

	version := tmuxVersionRe.FindStringSubmatch(string(output))
	if len(version) < 3 {
		return 0, 0
	}

	major, _ := strconv.Atoi(version[1])
	minor, _ := strconv.Atoi(version[2])

	return major, minor
}

func newTmuxClient(session, socket, root, configPath string) *tmuxClient {
	t := tmuxClient{
		processes:       make([]*process, 0),
		processesByPane: make(processesMap),

		configPath: configPath,

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
		if first {
			first = false

			args = append(args, "new", "-c", p.Dir, "-n", p.Name, "-s", t.Session, "-P", "-F", tmuxPaneFmt, p.Command, ";")

			if major, minor := tmuxVersion(); major < 2 || (major == 2 && minor < 6) {
				if w, h, err := terminal.GetSize(int(os.Stdin.Fd())); err == nil {
					args = append(args, "refresh", "-C", fmt.Sprintf("%d,%d", w, h), ";")
				}
			}

			args = append(args, "setw", "-g", "remain-on-exit", "on", ";")
			args = append(args, "setw", "-g", "allow-rename", "off", ";")
		} else {
			args = append(args, "neww", "-c", p.Dir, "-n", p.Name, "-P", "-F", tmuxPaneFmt, p.Command, ";")
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

	utils.Fatal("Tmux client unexpectidly exited")
}

func (t *tmuxClient) mapProcess(pane, name, pid string) {
	for _, p := range t.processes {
		if p.Name != name {
			continue
		}

		t.processesByPane[pane] = p

		if ipid, err := strconv.Atoi(pid); err == nil {
			p.proc, _ = os.FindProcess(-ipid)
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
	t.sendCmd("neww -d -k -t %s -n %s -P -F %q %s", p.Name, p.Name, tmuxPaneFmt, command)
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
