# Overmind

<img align="right" width="224" height="74" title="Overmind logo"
     src="https://cdn.rawgit.com/DarthSim/overmind/master/logo.svg">

[![Build Status](https://travis-ci.org/DarthSim/overmind.svg?branch=master)](https://travis-ci.org/DarthSim/overmind)

Overmind is a process manager for Procfile-based applications and [tmux](https://tmux.github.io/). With Overmind, you can easily run several processes from your `Procfile` in a single terminal.

Procfile is a simple format to specify types of processes your application provides (such as web application server, background queue process, front-end builder) and commands to run those processes. It can significantly simplify process management for developers and is used by popular hosting platforms, such as Heroku and Deis. You can learn more about the `Procfile` format [here](https://devcenter.heroku.com/articles/procfile) or [here](http://docs.deis.io/en/latest/using_deis/process-types/).

There are some good Procfile-based process management tools, including [foreman](https://github.com/ddollar/foreman) by David Dollar, which started it all. The problem with most of those tools is that processes you want to manage start to think they are logging their output into a file, and that can lead to all sorts of problems: severe lagging, losing or breaking colored output. Tools can also add vanity information (unneeded timestamps in logs). Overmind was created to fix those problems once and for all.

See this article for a good intro and all the juicy details! [Introducing
Overmind and Hivemind](https://evilmartians.com/chronicles/introducing-overmind-and-hivemind)

<a href="https://evilmartians.com/?utm_source=overmind">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="236" height="54">
</a>

## Overmind features

You may know several Procfile process management tools, but Overmind has some unique, _extraterrestrial_ powers others don't:

* Overmind starts processes in a tmux session, so you can easily connect to any process and gain control over it;
* Overmind can restart a single process on the fly — you don't need to restart the whole stack;
* Overmind allows a specified process to die without interrupting all of the other ones;
* Overmind uses `pty` to capture process output — so it won't be clipped, delayed, and it won't break colored output;
* Overmind can read environment variables from a file and use them as parameters so that you can configure Overmind behavior globally and/or per directory.

**If a lot of those features seem like overkill for you, especially the tmux integration, you should take a look at Overmind's little sister — [Hivemind](https://github.com/DarthSim/hivemind)!**

![Overmind screenshot](http://i.imgur.com/lfrFKMf.png)

## Installation

**Note:** At the moment, Overmind supports Linux, *BSD, and macOS only.

Overmind works with [tmux](https://tmux.github.io/), so you need to install it first:

```bash
# on macOS (with homebrew)
$ brew install tmux

# on Ubuntu
$ apt-get install tmux
```

**Note:** You can find installation manual for other systems here: https://github.com/tmux/tmux

There are three ways to install Overmind:

### With Homebrew (macOS)

```bash
brew install overmind
```

### Download the latest Overmind release binary

You can download the latest release [here](https://github.com/DarthSim/overmind/releases/latest).

### Build Overmind from source

You need Go 1.6 or later to build the project.

```bash
$ go get -u -f github.com/DarthSim/overmind
```

**Note:** You can update Overmind the same way.

## Usage

**In short:** You can get help by running `overmind -h` and `overmind help [command]`.

### Running processes

Overmind reads the list of processes you want to manage from a file named `Procfile`. It may look like this:

```Procfile
web: bin/rails server
worker: bundle exec sidekiq
assets: gulp watch
```

To get started, you just need to run Overmind from your working directory containing a `Procfile`:

```bash
$ overmind start
```

You can also use the short alias:

```bash
$ overmind s
```

#### Specifying a Procfile

If a `Procfile` isn't located in your working directory, you can specify the exact path:

```bash
$ overmind start -f path/to/your/Procfile
$ OVERMIND_PROCFILE=path/to/your/Procfile overmind start
```

#### Specifying the ports

Overmind sets environment variable `PORT` for each process in your Procfile so that you can do things like this:

```Procfile
web: bin/rails server -p $PORT
```

Overmind assigns the port base (5000 by default) to `PORT` for the first process and increases `PORT` by port step (100 by default) for the each next one. You can specify port base and port step like this:

```bash
$ overmind start -p 3000 -P 10
$ OVERMIND_PORT=3000 OVERMIND_PORT_STEP=10 overmind start
```

#### Running only the specified processes

You can specify the names of processes you want to run:

```bash
$ overmind start -l web,sidekiq
$ OVERMIND_PROCESSES=web,sidekiq overmind start
```

#### Processes that can die

Usually, when a process dies, Overmind will interrupt all other processes. However, you can specify processes that can die without interrupting all other ones:

```bash
$ overmind start -c assets,npm_install
$ OVERMIND_CAN_DIE=assets,npm_install overmind start
```

#### Specifying the colors

Overmind colorizes process names with different colors. May happen that these colors don't match well with your color scheme. In this case, you can define your own colors using xterm color codes:

```bash
$ overmind start -b 123,123,125,126,127
$ OVERMIND_COLORS=123,123,125,126,127 overmind start
```

If you want Overmind to always use these colors, you can specify them in the [environment file](https://github.com/DarthSim/overmind#overmind-environment) located in your home directory.

### Connecting to a process

If you need to gain access to process input, you can connect to its `tmux` window:

```bash
$ overmind connect [process_name]
```

You can safely disconnet from the window by hitting `Ctrl b` and then `d`.

### Restarting a process

You can restart a single process without restarting all the other ones:

```bash
$ overmind restart sidekiq
```

You can restart multiple processes the same way:

```bash
$ overmind restart sidekiq assets
```

### Stopping a process

You can stop a single process without stopping all the other ones:

```bash
$ overmind stop sidekiq
```

You can stop multiple processes the same way:

```bash
$ overmind stop sidekiq assets
```

### Killing processes

If something goes wrong, you can kill all running processes:

```bash
$ overmind kill
```

### Overmind environment

If you need to set specific environment variables before running a `Procfile`, you can specify them in the `.overmind.env` file in the current working directory, your home directory, or/and in the `.env` file in in the current working directory. The file should contain `variable=value` pairs, one per line:

```
PATH=$PATH:/additional/path
OVERMIND_CAN_DIE=npm_install
OVERMIND_PORT=3000
```

You can specify additional env file to load with `OVERMIND_ENV` variable:

```
$ OVERMIND_ENV=path/to/env overmind s
```

### Specifying a socket

Overmind receives commands via a Unix socket. Usually, it opens a socket named `.overmind.sock` in a working directory, but you can specify the full path:

```bash
$ overmind start -s path/to/socket
$ OVERMIND_SOCKET=path/to/socket overmind start
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

Many thanks to @antiflasher for the awesome logo.

## License

Overmind is licensed under the MIT license.

See LICENSE for the full license text.
