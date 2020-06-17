package start

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"syscall"

	"github.com/DarthSim/overmind/v2/utils"
)

type procfileEntry struct {
	Name     string
	Command  string
	Procfile string
}

type procEntry struct {
	Name       string
	OrigName   string
	Command    string
	Port       int
	Procfile   string
	StopSignal syscall.Signal
}

type procfile []procfileEntry
type procs []procEntry

func resolveProcs(h *Handler) (ps procs) {
	port := h.PortBase
	root, _ := h.AbsRoot()
	names := make(map[string]bool)

	pf := parseProcfile(filepath.Clean(h.Procfile), root)
	for i, p := range pf {
		name := p.Name
		iname := name

		shortPath := p.Procfile
		if filepath.Base(p.Procfile) == "Procfile" {
			shortPath = filepath.Dir(p.Procfile)
		}

		if len(shortPath) > 1 {
			iname = fmt.Sprintf("%s/%s", shortPath, iname)
		}

		num := 1
		if fnum, ok := h.Formation[name]; ok {
			num = fnum
		} else if fnum, ok := h.Formation["all"]; ok {
			num = fnum
		}

		if num > 1 {
			iname = fmt.Sprintf("%s#%d", iname, i+1)
		}

		if names[iname] {
			utils.Fatal("Process names must be uniq")
		}
		names[iname] = true

		signal := syscall.SIGINT
		if s, ok := h.StopSignals[name]; ok {
			signal = s
		}

		ps = append(
			ps,
			procEntry{
				Name:       iname,
				OrigName:   name,
				Command:    p.Command,
				Procfile:   p.Procfile,
				Port:       port + (i * h.FormationPortStep),
				StopSignal: signal,
			},
		)
		port += h.PortStep
	}
	return
}

// procfile: relative path to the Procfile to be parsed
// absDir: absolute path from where overmind handler were started
func parseProcfile(procfile string, absDir string) (pf procfile) {
	reProc, _ := regexp.Compile(`^([\w-]+):\s+(.+)$`)
	reFile, _ := regexp.Compile(`^-f\s+(.+)$`)

	f, err := os.Open(procfile)
	utils.FatalOnErr(err)

	err = utils.ScanLines(f, func(b []byte) bool {
		if len(b) == 0 {
			return true
		}

		fparams := reFile.FindStringSubmatch(string(b))
		if len(fparams) == 2 {
			nextProcfile := fparams[1]

			if filepath.IsAbs(nextProcfile) {
				utils.Fatal("Nested Procfile must use relative path")
			}

			nextPFPath := filepath.Join(filepath.Dir(procfile), nextProcfile)

			npf := parseProcfile(nextPFPath, absDir)
			pf = append(pf, npf...)
		}

		pparams := reProc.FindStringSubmatch(string(b))
		if len(pparams) == 3 {
			name, cmd := pparams[1], pparams[2]
			pf = append(
				pf,
				procfileEntry{
					Name:       name,
					Command:    cmd,
					Procfile:   procfile,
				},
			)
		}

		return true
	})

	utils.FatalOnErr(err)

	if len(pf) == 0 {
		utils.Fatal(fmt.Sprintf("No entries was found in '%s'", procfile))
	}

	return
}

func (p procs) MaxNameLength() (nl int) {
	for _, e := range p {
		if l := len(e.Name); nl < l {
			nl = l
		}
	}
	return
}
