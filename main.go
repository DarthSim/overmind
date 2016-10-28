package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/DarthSim/overmind/launch"
	"github.com/DarthSim/overmind/start"

	"gopkg.in/alecthomas/kingpin.v2"
)

func setupStartCmd(app *kingpin.Application) {
	c := start.Handler{}

	run := app.Command("start", "Run procfile.").Default().Action(c.Run)

	run.Flag("procfile", "Specify a Procfile to load").Default("./Procfile").Short('f').StringVar(&c.Procfile)
	run.Flag("processes", "Specify process names to lunch. Process should be specified in the Procfile.").StringVar(&c.ProcNames)
	run.Flag("root", "Specify a working directory of application. Default: directory containing the Procfile").Short('d').StringVar(&c.Root)
	run.Flag("timeout", "Specify the amount of time (in seconds) processes have to shut down gracefully before being brutally killed").Default("5").Short('t').IntVar(&c.Timeout)
	run.Flag("port", "Specify a port to use as the base").Default("5000").Short('p').IntVar(&c.PortBase)
	run.Flag("port-step", "Specify a step to increase port number").Default("100").Short('P').IntVar(&c.PortStep)
	run.Flag("socket", "Specify a path to the command center socket").Default("./.overmind.sock").Short('s').StringVar(&c.SocketPath)
}

func setupLaunchCmd(app *kingpin.Application) {
	c := launch.Handler{}

	run := app.Command("launch", "Launch process, connect to overmind socket, wait for instructions.").Hidden().Action(c.Run)

	run.Arg("name", "Process name").Required().StringVar(&c.ProcessName)
	run.Arg("cmd", "Shell command").Required().StringVar(&c.CmdLine)
	run.Arg("socket", "Path to overmind socket").Required().StringVar(&c.SocketPath)
}

func setupRestartCmd(app *kingpin.Application) {
	c := cmdRestartHandler{}

	run := app.Command("restart", "Restart specified processes").Action(c.Run)

	run.Arg("name", "Process name").Required().StringsVar(&c.ProcessNames)
	run.Flag("socket", "Path to overmind socket").Default("./.overmind.sock").Short('s').StringVar(&c.SocketPath)
}

func setupConnectCmd(app *kingpin.Application) {
	c := cmdConnectHandler{}

	run := app.Command("connect", "Connect to the tmux session of specified process").Action(c.Run)

	run.Arg("name", "Process name").Required().StringVar(&c.ProcessName)
	run.Flag("socket", "Path to overmind socket").Default("./.overmind.sock").Short('s').StringVar(&c.SocketPath)
}

func setupKillCmd(app *kingpin.Application) {
	c := cmdKillHandler{}

	run := app.Command("kill", "Kills tmux sessions of all processes").Action(c.Run)

	run.Flag("socket", "Path to overmind socket").Default("./.overmind.sock").Short('s').StringVar(&c.SocketPath)
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	app := kingpin.New("overmind", "The mind to rule your development processes.")

	setupStartCmd(app)
	setupLaunchCmd(app)
	setupRestartCmd(app)
	setupConnectCmd(app)
	setupKillCmd(app)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
