package main

import (
	"os"
	"path"
	"strings"

	"github.com/DarthSim/godotenv"
	"github.com/DarthSim/overmind/v2/start"

	"github.com/urfave/cli"
)

const version = "2.4.0"

func socketFlags(s, n *string) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "socket, s",
			EnvVar:      "OVERMIND_SOCKET",
			Usage:       "Path to overmind socket (in case of using unix socket) or address to bind (in other cases)",
			Value:       "./.overmind.sock",
			Destination: s,
		},
		cli.StringFlag{
			Name:        "network, S",
			EnvVar:      "OVERMIND_NETWORK",
			Usage:       "Network to use for commands. Can be 'tcp', 'tcp4', 'tcp6' or 'unix'",
			Value:       "unix",
			Destination: n,
		},
	}
}

func setupStartCmd() cli.Command {
	c := start.Handler{}

	return cli.Command{
		Name:    "start",
		Aliases: []string{"s"},
		Usage:   "Run procfile",
		Action:  c.Run,
		Flags: append(
			[]cli.Flag{
				cli.StringFlag{Name: "title, w", EnvVar: "OVERMIND_TITLE", Usage: "Specify a title of the application", Destination: &c.Title},
				cli.StringFlag{Name: "procfile, f", EnvVar: "OVERMIND_PROCFILE", Usage: "Specify a Procfile to load", Value: "./Procfile", Destination: &c.Procfile},
				cli.StringFlag{Name: "processes, l", EnvVar: "OVERMIND_PROCESSES", Usage: "Specify process names to launch. Divide names with comma", Destination: &c.ProcNames},
				cli.StringFlag{Name: "root, d", Usage: "Specify a working directory of application. Default: directory containing the Procfile", Destination: &c.Root},
				cli.IntFlag{Name: "timeout, t", EnvVar: "OVERMIND_TIMEOUT", Usage: "Specify the amount of time (in seconds) processes have to shut down gracefully before being brutally killed", Value: 5, Destination: &c.Timeout},
				cli.IntFlag{Name: "port, p", EnvVar: "OVERMIND_PORT,PORT", Usage: "Specify a port to use as the base", Value: 5000, Destination: &c.PortBase},
				cli.IntFlag{Name: "port-step, P", EnvVar: "OVERMIND_PORT_STEP", Usage: "Specify a step to increase port number", Value: 100, Destination: &c.PortStep},
				cli.BoolFlag{Name: "no-port, N", EnvVar: "OVERMIND_NO_PORT", Usage: "Don't set $PORT variable for processes", Destination: &c.NoPort},
				cli.StringFlag{Name: "can-die, c", EnvVar: "OVERMIND_CAN_DIE", Usage: "Specify names of process which can die without interrupting the other processes. Divide names with comma", Destination: &c.CanDie},
				cli.BoolFlag{Name: "any-can-die", EnvVar: "OVERMIND_ANY_CAN_DIE", Usage: "No dead processes should stop Overmind. Overrides can-die", Destination: &c.AnyCanDie},
				cli.StringFlag{Name: "auto-restart, r", EnvVar: "OVERMIND_AUTO_RESTART", Usage: "Specify names of process which will be auto restarted on death. Divide names with comma. Use 'all' as a process name to auto restart all processes on death.", Destination: &c.AutoRestart},
				cli.StringFlag{Name: "colors, b", EnvVar: "OVERMIND_COLORS", Usage: "Specify the xterm color codes that will be used to colorize process names. Divide codes with comma"},
				cli.BoolFlag{Name: "show-timestamps, T", EnvVar: "OVERMIND_SHOW_TIMESTAMPS", Usage: "Add timestamps to the output", Destination: &c.ShowTimestamps},
				cli.StringFlag{Name: "formation, m", EnvVar: "OVERMIND_FORMATION", Usage: "Specify the number of each process type to run. The value passed in should be in the format process=num,process=num. Use 'all' as a process name to set value for all processes"},
				cli.IntFlag{Name: "formation-port-step", EnvVar: "OVERMIND_FORMATION_PORT_STEP", Usage: "Specify a step to increase port number for the next instance of a process", Value: 10, Destination: &c.FormationPortStep},
				cli.StringFlag{Name: "stop-signals, i", EnvVar: "OVERMIND_STOP_SIGNALS", Usage: "Specify a signal that will be sent to each process when Overmind will try to stop them. The value passed in should be in the format process=signal,process=signal. Supported signals are: ABRT, INT, KILL, QUIT, STOP, TERM, USR1, USR2"},
				cli.BoolFlag{Name: "daemonize, D", EnvVar: "OVERMIND_DAEMONIZE", Usage: "Launch Overmind as a daemon. Use 'overmind echo' to view logs and 'overmind quit' to gracefully quit daemonized instance", Destination: &c.Daemonize},
				cli.StringFlag{Name: "tmux-config, F", EnvVar: "OVERMIND_TMUX_CONFIG", Usage: "Specify an alternative tmux config path to be used by Overmind", Destination: &c.TmuxConfigPath},
				cli.StringFlag{Name: "ignored-processes, x", EnvVar: "OVERMIND_IGNORED_PROCESSES", Usage: "Specify process names to prevent from launching. Useful if you want to run all but one or two processes. Divide names with comma. Takes precedence over the 'processes' flag.", Destination: &c.IgnoredProcNames},
				cli.StringFlag{Name: "delayed-processes, X", EnvVar: "OVERMIND_DELAYED_PROCESSES", Usage: "Specify process names to prevent from launching right now. Useful if you want to run one or two processes later. Divide names with comma. Takes precedence over the 'processes' flag.", Destination: &c.DelayedProcNames},
				cli.StringFlag{Name: "shell, H", EnvVar: "OVERMIND_SHELL", Usage: "Specify shell to run processes with.", Value: "sh", Destination: &c.Shell},
			},
			socketFlags(&c.SocketPath, &c.Network)...,
		),
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
		Flags:     socketFlags(&c.SocketPath, &c.Network),
	}
}

func setupStopCmd() cli.Command {
	c := cmdStopHandler{}

	return cli.Command{
		Name:      "stop",
		Aliases:   []string{"interrupt", "i"},
		Usage:     "Stop specified processes without quitting Overmind itself",
		Action:    c.Run,
		ArgsUsage: "[process name...]",
		Flags:     socketFlags(&c.SocketPath, &c.Network),
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
		Flags: append(
			[]cli.Flag{
				cli.BoolFlag{Name: "control-mode, c", EnvVar: "OVERMIND_CONTROL_MODE", Usage: "Connect to the tmux session in control mode", Destination: &c.ControlMode},
			},
			socketFlags(&c.SocketPath, &c.Network)...,
		),
	}
}

func setupQuitCmd() cli.Command {
	c := cmdQuitHandler{}

	return cli.Command{
		Name:    "quit",
		Aliases: []string{"q"},
		Usage:   "Gracefully quits Overmind. Same as sending SIGINT",
		Action:  c.Run,
		Flags:   socketFlags(&c.SocketPath, &c.Network),
	}
}

func setupKillCmd() cli.Command {
	c := cmdKillHandler{}

	return cli.Command{
		Name:    "kill",
		Aliases: []string{"k"},
		Usage:   "Kills all processes",
		Action:  c.Run,
		Flags:   socketFlags(&c.SocketPath, &c.Network),
	}
}

func setupRunCmd() cli.Command {
	c := cmdRunHandler{}

	return cli.Command{
		Name:            "run",
		Aliases:         []string{"exec", "e"},
		Usage:           "Runs provided command within the Overmind environment",
		Action:          c.Run,
		SkipFlagParsing: true,
	}
}

func setupEchoCmd() cli.Command {
	c := cmdEchoHandler{}

	return cli.Command{
		Name:   "echo",
		Usage:  "Echoes output from master Overmind instance",
		Action: c.Run,
		Flags:  socketFlags(&c.SocketPath, &c.Network),
	}
}

func setupStatusCmd() cli.Command {
	c := cmdStatusHandler{}

	return cli.Command{
		Name:    "status",
		Aliases: []string{"ps"},
		Usage:   "Prints process statuses",
		Action:  c.Run,
		Flags:   socketFlags(&c.SocketPath, &c.Network),
	}
}

func main() {
	loadEnvFiles()

	app := cli.NewApp()

	app.Name = "Overmind"
	app.HelpName = "overmind"
	app.Usage = "The mind to rule processes of your development environment"
	app.Description = strings.Join([]string{
		"Overmind runs commands specified in procfile in a tmux session.",
		"This allows to connect to each process and manage processes on fly.",
	}, " ")
	app.Author = "Sergey \"DarthSim\" Alexandrovich"
	app.Email = "darthsim@gmail.com"
	app.Version = version

	app.Commands = []cli.Command{
		setupStartCmd(),
		setupRestartCmd(),
		setupStopCmd(),
		setupConnectCmd(),
		setupQuitCmd(),
		setupKillCmd(),
		setupRunCmd(),
		setupEchoCmd(),
		setupStatusCmd(),
	}

	app.Run(os.Args)
}

func loadEnvFiles() {
	// First load the specifically named overmind env files
	userHomeDir, _ := os.UserHomeDir()
	godotenv.Overload(path.Join(userHomeDir, ".overmind.env"))
	godotenv.Overload("./.overmind.env")

	_, skipEnv := os.LookupEnv("OVERMIND_SKIP_ENV")
	if !skipEnv {
		godotenv.Overload("./.env")
	}

	envs := strings.Split(os.Getenv("OVERMIND_ENV"), ",")
	for _, e := range envs {
		if len(e) > 0 {
			godotenv.Overload(e)
		}
	}
}
