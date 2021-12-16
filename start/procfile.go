package start

import (
	"fmt"
	"os"
	"regexp"
	"syscall"

	"github.com/DarthSim/overmind/v2/utils"
)

type procfileEntry struct {
	Name       string
	OrigName   string
	Command    string
	Port       int
	StopSignal syscall.Signal
}

type procfile []procfileEntry

func parseProcfile(procfile string, portBase, portStep int, formation map[string]int, formationPortStep int, stopSignals map[string]syscall.Signal) (pf procfile) {
	re, _ := regexp.Compile(`^([\w-]+):\s+(.+)$`)

	f, err := os.Open(procfile)
	utils.FatalOnErr(err)

	port := portBase
	names := make(map[string]bool)

	err = utils.ScanLines(f, func(b []byte) bool {
		if len(b) == 0 {
			return true
		}

		params := re.FindStringSubmatch(string(b))
		if len(params) != 3 {
			return true
		}

		name, cmd := params[1], params[2]

		num := 1
		if fnum, ok := formation[name]; ok {
			num = fnum
		} else if fnum, ok := formation["all"]; ok {
			num = fnum
		}

		signal := syscall.SIGINT
		if s, ok := stopSignals[name]; ok {
			signal = s
		}

		for i := 0; i < num; i++ {
			iname := name

			if num > 1 {
				iname = fmt.Sprintf("%s#%d", name, i+1)
			}

			if names[iname] {
				utils.Fatal("Process names must be unique")
			}
			names[iname] = true

			pf = append(
				pf,
				procfileEntry{
					Name:       iname,
					OrigName:   name,
					Command:    cmd,
					Port:       port + (i * formationPortStep),
					StopSignal: signal,
				},
			)
		}

		port += portStep

		return true
	})

	utils.FatalOnErr(err)

	if len(pf) == 0 {
		utils.Fatal("No entries were found in Procfile")
	}

	return
}

func (p procfile) MaxNameLength() (nl int) {
	for _, e := range p {
		if l := len(e.Name); nl < l {
			nl = l
		}
	}
	return
}
