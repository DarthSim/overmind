package start

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/DarthSim/overmind/utils"
)

type multiOutput struct {
	maxNameLength int
	mutex         sync.Mutex
}

func newMultiOutput(maxNameLength int) *multiOutput {
	return &multiOutput{
		maxNameLength: utils.Max(maxNameLength, 6),
	}
}

func (o *multiOutput) WriteLine(proc *process, p []byte) {
	var (
		buf   bytes.Buffer
		name  string
		color int
	)

	if proc != nil {
		name = proc.Name
		color = proc.Color
	} else {
		name = "system"
		color = 37
	}

	colorStr := fmt.Sprintf("\033[1;%vm", color)

	buf.WriteString(colorStr)
	buf.WriteString(name)

	for buf.Len()-len(colorStr) < o.maxNameLength {
		buf.WriteByte(' ')
	}

	buf.WriteString("\033[0m | ")
	buf.Write(p)
	buf.WriteByte('\n')

	o.mutex.Lock()
	defer o.mutex.Unlock()

	buf.WriteTo(os.Stdout)
}

func (o *multiOutput) WriteBoldLine(proc *process, p []byte) {
	o.WriteLine(proc, []byte(
		fmt.Sprintf("\033[1m%s\033[0m", p),
	))
}

func (o *multiOutput) WriteBoldLinef(proc *process, format string, i ...interface{}) {
	o.WriteBoldLine(proc, []byte(fmt.Sprintf(format, i...)))
}

func (o *multiOutput) WriteErr(proc *process, err error) {
	o.WriteLine(proc, []byte(fmt.Sprintf("\033[0;31m%v\033[0m", err)))
}
