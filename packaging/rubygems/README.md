## Ruby wrapper Overmind
This gem wraps the [Overmind](https://github.com/DarthSim/overmind) library and includes all the dependencies needed to work with it.

Overmind is a process manager for Procfile-based applications and [tmux](https://tmux.github.io/). 
With Overmind, you can easily run several processes from your Procfile in a single terminal.

Learn more about Overmind [here](https://github.com/DarthSim/overmind).

## Installation
**Note:** At the moment, Overmind supports Linux, *BSD, and macOS only.

### Requirements
This gem already has all the necessary dependencies and doesn't require any libraries to be installed.
But for users of the *BSD system, it requires the installation of `tmux`.

- **FreeBSD:**
```bash
  pkg install tmux
```

**Note:**: You can find more information about the `tmux` installation [here](https://github.com/tmux/tmux)

### Installation with Ruby

```bash
  gem install overmind
```

### Installation with Rails
Overmind can improve your DX of working on multi-process Ruby on Rails applications.
First, add it to your Gemfile:

```ruby
  group :development do
    gem "overmind"
  end
```

We recommend installing it as a part of your project and not globally, so the version is kept in sync for everyone on your team.

Then, for simplicity, we suggest generating a bin stub:

```bash
  bundle binstubs overmind
```

Finally, change the contents of `bin/dev` (or add this file) as follows:

```shell
  #!/usr/bin/env sh

  bin/overmind start -f Procfile.dev
```

Now, your `bin/dev` command uses Overmind under the hood.

One of the biggest benefits is that now you can connect to any process and interact with it, for example, for debugging a web process:

```bash
  bin/overmind connect web
```

## Usage

### Running processes

Overmind reads the list of processes you want to manage from a file named `Procfile`. It may look like this:

```Procfile
web: bin/rails server
worker: bundle exec sidekiq
assets: gulp watch
```

To get started, you just need to run Overmind from your working directory containing a `Procfile`:

```bash
  # in Rails project
  bin/overmind start
  
  # Ruby
  overmind start
```

#### Specifying a Procfile

If a `Procfile` isn't located in your working directory, you can specify the exact path:

```bash
  bin/overmind start -f path/to/your/Procfile
  OVERMIND_PROCFILE=path/to/your/Procfile bin/overmind start
```

### Connecting to a process

If you need to gain access to process input, you can connect to its `tmux` window:

```bash
  bin/overmind connect <process_name>
```

You can safely disconnect from the window by hitting `Ctrl b` (or your tmux prefix) and then `d`.

You can omit the process name to connect to the first process defined in the Procfile.

### Restarting a process

You can restart a single process without restarting all the other ones:

```bash
  bin/overmind restart sidekiq
```

### Stopping a process

You can stop a single process without stopping all the other ones:

```bash
  bin/overmind stop sidekiq
```

More features and abilities are [here](https://github.com/DarthSim/overmind)
