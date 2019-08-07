# devdrake

__devdrake__, whose executable is abbreviated as `drake`, is a
developer-oriented command-line tool for executing Drake jobs and pipelines
using the local Docker daemon. The ability to do so offers developers some major benefits:

* Makes it practical and convenient to define useful jobs like `test`, `lint`,
  or `build` (for example) in _one, common place_ then seamlessly reuse those
  jobs across both CI/CD pipelines and local development workflows.

* Ensures parity between CI/CD platforms and the local development environment
  by sandboxing critical work in containers with a single, common definition.

For more information about Drake jobs and pipelines, please refer to the
[drakespec](https://github.com/lovethedrake/drakespec).

## Installation

To install `drake`, head over to the
[releases page](https://github.com/lovethedrake/devdrake/releases) page to
download a pre-built binary appropriate for your operating system and CPU
architecture.

We recommend renaming the downloaded file as `drake` (or `drake.exe` for
Windows) then moving it to some file system location that is on your `PATH`.

## Use

`drake` has interactive help that is available by typing `drake -h`, so this
section covers only the most ubiquitous and most useful commands to help you get
started.

As with all Drake-compliant tools, `drake` utilizes a `Drakefile.yaml` file that
defines Drake jobs and pipelines. By convention, `Drakefile.yaml` is typically
located at the root of the project tree.

If you need a sample project with a trivial `Drakefile.yaml` to experiment with,
consider cloning the
[simple-demo](https://github.com/lovethedrake/simple-demo) project.

By default, `drake` assumes it is being executed with the project's root as your
current working directory, and therefore assumes `Drakefile.yaml` exists in the
current working directory. If this is not the case, all `drake` sub-commands
support optional `--file` and `-f` flags-- either of which can be used to point
`drake` to the location of your `Drakefile.yaml`.

### Listing available jobs or pipelines

To list all jobs that you can execute:

```console
$ drake list
```

The `list` sub-command is also aliased as `ls`:

```console
$ drake ls
```

To list pipelines instead of jobs, use either the `--pipeline` or `-p` flag:

```console
$ drake list --pipeline
```

### Executing jobs and pipeline

To execute a single job:

```console
$ drake run <job-name>
```

To execute multiple jobs in sequence, just provide additional job names as
additional arguments:

```console
$ drake run <job-name-0> <job-name-1> ... <job-name-n>
```

By default, multiple jobs execute in sequence, but it is possible to execute
multiple jobs concurrently using either the `--concurrently` or `-c` flag:

```console
$ drake run <job-name-0> <job-name-1> ... <job-name-n> --concurrently
```

Note: Because multiple jobs can execute concurrently and because jobs can
involve multiple, co-operating containers, `drake` prefixes every line of output
with the job name and container name to disambiguate its source.

To execute a pipeline instead of a job, use either the `--pipeline` or `-p` flag:

```console
$ drake run <pipeline-name>
```

Drake pipelines are composed of stages that always execute in sequence. Stages,
in turn, are composed of jobs that MAY execute concurrently. `drake` never
executes jobs concurrently by default. To allow concurrent execution of the jobs within each stage, once again, utilize either the `--concurrently` or `-c`
flag:

```console
$ drake run <pipeline-name> --concurrently
```

It is also possible to execute multiple pipelines in sequence:

```console
$ drake run <pipeline-name-0> <pipeline-name-1> ... <pipeline-name-n>
```

Note: When executing pipelines, the `--concurrently` or `-c` flag only enables
concurrent execution of _jobs_ within each _stage_ of a pipeline. Stages will
_always_ execute in sequence, and never concurrently. Similarly, when naming
multiple pipelines for execution, those pipelines also will execute in sequence,
and never concurrently.

## Contributing

This project accepts contributions via GitHub pull requests. The
[Contributing](CONTRIBUTING.md) document outlines the process to help get your
contribution accepted.

## Code of Conduct

Although not a CNCF project, this project abides by the
[CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).
