package main

import (
	"bufio"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/DarthSim/overmind/launch"
	"github.com/DarthSim/overmind/start"

	"gopkg.in/urfave/cli.v1"
)

const version = "0.0.1"

func setupStartCmd() cli.Command {
	c := start.Handler{}

	return cli.Command{
		Name:    "start",
		Aliases: []string{"s"},
		Usage:   "Run procfile",
		Action:  c.Run,
		Flags: []cli.Flag{
			cli.StringFlag{Name: "procfile, f", EnvVar: "OVERMIND_PROCFILE", Usage: "Specify a Procfile to load", Value: "./Procfile", Destination: &c.Procfile},
			cli.StringFlag{Name: "processes, l", EnvVar: "OVERMIND_PROCESSES", Usage: "Specify process names to lunch. Process should be specified in the Procfile.", Destination: &c.ProcNames},
			cli.StringFlag{Name: "root, d", Usage: "Specify a working directory of application. Default: directory containing the Procfile", Destination: &c.Root},
			cli.IntFlag{Name: "timeout, t", EnvVar: "OVERMIND_TIMEOUT", Usage: "Specify the amount of time (in seconds) processes have to shut down gracefully before being brutally killed", Value: 5, Destination: &c.Timeout},
			cli.IntFlag{Name: "port, p", EnvVar: "OVERMIND_PORT", Usage: "Specify a port to use as the base", Value: 5000, Destination: &c.PortBase},
			cli.IntFlag{Name: "port-step, P", EnvVar: "OVERMIND_PORT_STEP", Usage: "Specify a step to increase port number", Value: 100, Destination: &c.PortStep},
			cli.StringFlag{Name: "socket, s", Usage: "Specify a path to the command center socket", Value: "./.overmind.sock", Destination: &c.SocketPath},
		},
	}
}

func setupLaunchCmd() cli.Command {
	return cli.Command{
		Name:      "launch",
		Usage:     "Launch process, connect to overmind socket, wait for instructions",
		Action:    launch.Run,
		ArgsUsage: "[process name] [shell command] [path to overmind socket]",
		Hidden:    true,
	}
}

func setupRestartCmd() cli.Command {
	c := cmdRestartHandler{}

	return cli.Command{
		Name:      "restart",
		Aliases:   []string{"r"},
		Usage:     "Restart specified processes",
		Action:    c.Run,
		ArgsUsage: "[process name...]",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "socket, s", Usage: "Path to overmind socket", Value: "./.overmind.sock", Destination: &c.SocketPath},
		},
	}
}

func setupConnectCmd() cli.Command {
	c := cmdConnectHandler{}

	return cli.Command{
		Name:      "connect",
		Aliases:   []string{"c"},
		Usage:     "Connect to the tmux session of the specified process",
		Action:    c.Run,
		ArgsUsage: "[process name]",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "socket, s", Usage: "Path to overmind socket", Value: "./.overmind.sock", Destination: &c.SocketPath},
		},
	}
}

func setupKillCmd() cli.Command {
	c := cmdKillHandler{}

	return cli.Command{
		Name:    "kill",
		Aliases: []string{"k"},
		Usage:   "Kills tmux sessions of all processes",
		Action:  c.Run,
		Flags: []cli.Flag{
			cli.StringFlag{Name: "socket, s", Usage: "Path to overmind socket", Value: "./.overmind.sock", Destination: &c.SocketPath},
		},
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	loadEnvFile()

	app := cli.NewApp()

	app.Name = "Overmind"
	app.HelpName = "overmind"
	app.Usage = "The mind to rule processes of your development environment"
	app.Description = strings.Join([]string{
		"Overmind runs commands specified in procfile in separate tmux sessions.",
		"This allows to connect to each process and manage processes on fly.",
	}, " ")
	app.Author = "Sergey \"DarthSim\" Alexandrovich"
	app.Email = "darthsim@gmail.com"
	app.Version = version

	app.Commands = []cli.Command{
		setupStartCmd(),
		setupLaunchCmd(),
		setupRestartCmd(),
		setupConnectCmd(),
		setupKillCmd(),
	}

	app.Run(os.Args)
}

func loadEnvFile() {
	re, _ := regexp.Compile("^(\\w+)=(.+)$")

	f, err := os.Open("./.overmind.env")
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if env := re.FindStringSubmatch(scanner.Text()); len(env) == 3 {
			os.Setenv(env[1], env[2])
		}
	}
}
