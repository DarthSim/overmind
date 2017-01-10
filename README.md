# Overmind

[![Build Status](https://travis-ci.org/DarthSim/overmind.svg?branch=master)](https://travis-ci.org/DarthSim/overmind)

Overmind allows you to launch a number of processes in a single terminal using [tmux](https://tmux.github.io/) and Procfile.

<a href="https://evilmartians.com/?utm_source=overmind">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="236" height="54">
</a>

## The Overmind powers

There are a lot of Procfile runners written in different languages, but Overmind has some superpowers, that the other runners don't:

* Overmind starts processes in a tmux session, so you can connect to a process and get control over it;
* Overmind can restart processes on a fly, so you don't need to restart a whole stack;
* Overmind allows specified processes to die without interrupting all other ones;
* Overmind uses pty to capture processes output, so it won't be cut, delayed, and colors will remain the same;
* Overmind reads environment variables from a file and uses them as params, so you can configure Overmind's behavior globally and/or per directory.

_Don't need all this powers? You may be interested in the Overmind's little sister - [Hivemind](https://github.com/DarthSim/hivemind)_

## Installation

_Because of using pty and tmux, Overmind doesn't support Windows._

Overmind works with [tmux](https://tmux.github.io/), so you need to install it first:

```bash
# on Ubuntu
$ apt-get install tmux
# on Mac OS X (with homebrew)
$ brew install tmux
```

__Note:__ You can find installation manual for other systems here: https://github.com/tmux/tmux

There are two ways to install Overmind itself:

### Download the last Overmind release binary

You can download the last release [here](https://github.com/DarthSim/overmind/releases/latest).

### Build Overmind from source

You need Go 1.6 or later to build the project.

```bash
$ go get -u -f github.com/DarthSim/overmind
```

__Note:__ You can update Overmind the same way.

## Usage

**TL;DR:** You can get help by running `overmind -h` and `overmind help [command]`.

### Running processes

Overmind reads the processes you want to run from a Procfile, that looks like this:

```Procfile
web: bin/rails server
worker: bundle exec sidekiq
assets: gulp watch
```

To get started, you just need to run Overmind from your working directory containing a Procfile.

```bash
$ overmind start
```

You can also use the short alias:

```bash
$ overmind s
```

#### Specifying a Procfile

If a Procfile isn't located in your working directory, you can specify it:

```bash
$ overmind start -f path/to/your/Procfile
$ OVERMIND_PROCFILE=path/to/your/Procfile overmind start
```

#### Running only specified processes

You can specify the names of the processes you want to run:

```bash
$ overmind start -l web,sidekiq
$ OVERMIND_PROCESSES=web,sidekiq overmind start
```

#### Processes that can die

Normally when some process dies, Overmind interrupts all other processes. But you can specify processes, that can die without interrupting all other ones:

```bash
$ overmind start -c assets,npm_install
$ OVERMIND_CAN_DIE=assets,npm_install overmind start
```

### Connecting to a process

If you need to get access to a process input, you can connect to it's tmux window:

```bash
$ overmind connect [process_name]
```

### Restarting a process

You can restart a single process without restarting the other ones:

```bash
$ overmind restart sidekiq
```

You can restart multiple process the same way:

```bash
$ overmind restart sidekiq assets
```

### Killing processes

If something goes wrong, you can kill all running processes:

```bash
$ overmind kill
```

### Overmind environment

If you need to set specific environment variables before running a Procfile, you can specify them in `.overmind.env` in the current working directory and/or your home directory. The file should contain variable=value pairs one by line:

```
PATH=$PATH:/additional/path
OVERMIND_CAN_DIE=npm_install
OVERMIND_PORT_BASE=3000
```

### Specifying a socket

Overmind receives commands via a unix socket. Normally it opens a socket named `.overmind.sock` in a working directory, but you can specify it:

```bash
$ overmind start -s path/to/socket
```

All other commands support the same flag:

```bash
$ overmind connect -s path/to/socket web
$ overmind restart -s path/to/socket sidekiq
$ overmind kill -s path/to/socket
```

## Author

Sergey "DarthSim" Aleksandrovich

Highly inspired by [Foreman](https://github.com/ddollar/foreman).

## License

Overmind is licensed under the MIT license.

See LICENSE for the full license text.
