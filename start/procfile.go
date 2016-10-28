package start

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/DarthSim/overmind/utils"
)

type procfileEntry struct {
	Name    string
	Command string
}

type procfile []procfileEntry

func parseProcfile(procfile string, portBase, portStep int) (pf procfile) {
	re, _ := regexp.Compile("^(\\w+):\\s+(.+)$")

	f, err := os.Open(procfile)
	utils.FatalOnErr(err)

	port := portBase

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if len(scanner.Text()) > 0 {
			params := re.FindStringSubmatch(scanner.Text())
			if len(params) < 2 {
				utils.Fatal("Invalid process format: ", scanner.Text())
			}

			pf = append(pf, procfileEntry{
				params[1],
				strings.Replace(params[2], "$PORT", strconv.Itoa(port), -1),
			})

			port += portStep
		}
	}

	utils.FatalOnErr(scanner.Err())

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
