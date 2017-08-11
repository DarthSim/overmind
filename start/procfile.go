package start

import (
	"os"
	"regexp"

	"github.com/DarthSim/overmind/utils"
)

type procfileEntry struct {
	Name    string
	Command string
	Port    int
}

type procfile []procfileEntry

func parseProcfile(procfile string, portBase, portStep int) (pf procfile) {
	re, _ := regexp.Compile("^([\\w-]+):\\s+(.+)$")

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

		if names[name] {
			utils.Fatal("Process names must be uniq")
		}
		names[name] = true

		pf = append(pf, procfileEntry{name, cmd, port})

		port += portStep

		return true
	})

	utils.FatalOnErr(err)

	if len(pf) == 0 {
		utils.Fatal("No entries was found in Procfile")
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
