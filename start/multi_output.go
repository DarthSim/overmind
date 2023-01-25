package start

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/DarthSim/overmind/v2/utils"
)

type multiOutput struct {
	maxNameLength int

	ch   chan *bytes.Buffer
	done chan struct{}

	echoes    map[int64]io.Writer
	echoInd   int64
	echoMutex sync.Mutex

	bufPool sync.Pool
}

func newMultiOutput(maxNameLength int) *multiOutput {
	o := multiOutput{
		maxNameLength: utils.Max(maxNameLength, 6),
		ch:            make(chan *bytes.Buffer, 128),
		done:          make(chan struct{}),
		echoes:        make(map[int64]io.Writer),
		bufPool: sync.Pool{
			New: func() interface{} { return new(bytes.Buffer) },
		},
	}

	go o.listen()

	return &o
}

func (o *multiOutput) Offset() int {
	return o.maxNameLength + 3
}

func (o *multiOutput) listen() {
	for buf := range o.ch {
		b := buf.Bytes()

		os.Stdout.Write(b)

		if len(o.echoes) > 0 {
			o.writeToEchoes(b)
		}

		o.bufPool.Put(buf)
	}
	close(o.done)
}

func (o *multiOutput) Stop() {
	close(o.ch)
	<-o.done
}

func (o *multiOutput) writeToEchoes(b []byte) {
	o.echoMutex.Lock()
	defer o.echoMutex.Unlock()

	for i, e := range o.echoes {
		if _, err := e.Write(b); err != nil {
			delete(o.echoes, i)
			o.WriteBoldLinef(nil, "Echo #%d closed", i)
		}
	}
}

func (o *multiOutput) Echo(w io.Writer) {
	o.echoMutex.Lock()
	defer o.echoMutex.Unlock()

	o.echoInd++
	o.echoes[o.echoInd] = w

	o.WriteBoldLinef(nil, "Echo #%d opened", o.echoInd)
}

func (o *multiOutput) WriteLine(proc *process, p []byte) {
	var (
		name  string
		color int
	)

	buf := o.bufPool.Get().(*bytes.Buffer)
	buf.Reset()

	if proc != nil {
		name = proc.Name
		color = proc.Color
	} else {
		name = "system"
		color = 7
	}

	colorStr := fmt.Sprintf("\033[1;38;5;%vm", color)
	buf.WriteString(colorStr)
	utils.FprintRpad(buf, name, o.maxNameLength)
	buf.WriteString("\033[0m | ")

	buf.Write(p)
	buf.WriteByte('\n')

	o.ch <- buf
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
	for _, str := range strings.Split(err.Error(), "\n") {
		o.WriteLine(proc, []byte(fmt.Sprintf("\033[0;31m%v\033[0m", str)))
	}
}
