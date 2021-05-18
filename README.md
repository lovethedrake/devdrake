# Mallard

Mallard, is a developer-oriented command-line tool for executing Drake jobs and
pipelines using the local Docker daemon. The ability to do so offers developers
some major benefits:

* Makes it practical and convenient to define useful jobs like `test`, `lint`,
  or `build` (for example) in _one, common place_ then seamlessly reuse those
  jobs across both CI/CD pipelines and local development workflows.

* Ensures parity between CI/CD platforms and the local development environment
  by sandboxing critical work in containers with a single, common definition.

For more information about Drake-compliant jobs and pipelines, please refer
to the [Drake Specification](https://github.com/lovethedrake/spec).

## THIS PROJECT HIGHLY VOLATILE!

Mallard implements the highly volatile
[Drake Specification](https://github.com/lovethedrake/spec) and, as such is,
itself, highly volatile. Users are warned that breaking changes to this tool are
likely at any point up until its eventual 1.0 release.

## Installation

To install `mallard`, head over to the
[releases page](https://github.com/lovethedrake/mallard/releases) page to
download a pre-built binary appropriate for your operating system and CPU
architecture.

We recommend renaming the downloaded file as `mallard` (or `mallard.exe` for
Windows) then moving it to some file system location that is on your `PATH`.

## Use

`mallard` has interactive help that is available by typing `mallard -h`, so this
section covers only the most ubiquitous and most useful commands to help you get
started.

`mallard` utilizes a `Drakefile.yaml` file that defines jobs and pipelines using
the Drake DSL. By convention, `Drakefile.yaml` is typically located at the root
of the project tree.

If you need a sample project with a trivial `Drakefile.yaml` to experiment with,
consider cloning the
[simple-demo](https://github.com/lovethedrake/simple-demo) project.

By default, `mallard` assumes it is being executed with the project's root as
your current working directory, and therefore assumes `Drakefile.yaml` exists in
the current working directory. If this is not the case, all `mallard`
sub-commands support optional `--file` and `-f` flags-- either of which can be
used to point `mallard` to the location of your `Drakefile.yaml`.

### Listing available jobs or pipelines

To list all jobs that you can execute:

```console
$ mallard list
```

The `list` sub-command is also aliased as `ls`:

```console
$ mallard ls
```

To list pipelines instead of jobs, use either the `--pipeline` or `-p` flag:

```console
$ mallard list --pipeline
```

### Executing jobs and pipelines

To execute a single job:

```console
$ mallard run <job-name>
```

To execute multiple jobs, just provide additional job names as
additional arguments:

```console
$ mallard run <job-name-0> <job-name-1> ... <job-name-n>
```

Note, the command above creates an _ad-hoc_ pipeline wherein execution of each
job is contingent upon successful execution of the previous job. (Concurrent
execution of jobs in a pipeline is possible and discussed later in this section,
but these dependencies effectively preclude any concurrency.) If you need to
execute more complex pipelines or take advantage of any concurrent job
execution, it is recommended to explicitly create a named pipeline for that
workflow in your `Drakefile.yaml`.

Note: Because pipelines may contain multiple jobs and because jobs can involve
multiple, cooperating containers, `mallard` prefixes every line of output with
the job name and container name to disambiguate its source.

To execute a named pipeline instead of a series of jobs, use either the
`--pipeline` or `-p` flag:

```console
$ mallard run <pipeline-name> --pipeline
```

Note: Multiple pipelines _cannot_ be executed at once.

Pipelines are composed of jobs that MAY execute concurrently (wherever each
job's dependencies permit). However, by default, `mallard` does not executes
jobs in a pipeline concurrently. To enable concurrent execution of the jobs
within a pipeline, use either the `--concurrency` or `-c` flag with a value
greater than `1`.

```console
$ mallard run <pipeline-name> --pipeline --concurrency 2
```

Note: The `--concurrency` or `-c` flag only tunes the _maximum_ allowed
concurrency at any given moment during pipeline execution. Dependencies between
jobs in the pipeline may sometimes prevent this maximum degree of concurrency
from being achieved. For instance, given a pipeline containing two jobs, with
the second dependent on successful completion of the first, the jobs in the
pipeline can only ever execute in sequence, regardless of how high a value is
given to the `--concurrency` or `-c` flag.

### Using secrets

Jobs may, at times, require access to sensitive information that should not be
checked into source control. This means that including these values anywhere in
the job definitions in your `Drakefile.yaml` is not possible.

To permit `mallard` access to such secrets, add key/value pairs, to a "flat"
YAML file. By convention, this file is named `Drakesecrets.yaml` stored at the
root of your project.

For example:

```yaml
DOCKER_REGISTRY_NAMESPACE: krancour
DOCKER_PASSWORD: y7o9htGWkiVBbsqFjmeh
GITHUB_TOKEN: 7ddbd7a560b5b545b4bbeae0c006249adba0456a
```

It is also possible to source secret values from the values of _local_
environment variables using a simple expression syntax.

For instance:

```yaml
DOCKER_REGISTRY_NAMESPACE: ${DOCKER_ORG}
```

The above creates a secret named `DOCKER_REGISTRY_NAMESPACE` with the value
of the local environment variable `DOCKER_ORG`.

If it exists at the root of your project, all variants of the `mallard run`
sub-command will, by default, load secrets from the `Drakesecrets.yaml` file and
expose them as environment variables within _every container of every job that
it executes_.

Note: This behavior will change in a future release, as it is, at best, unwise
and, at worst, insecure, to inject every secret into every container. The Drake
specification is currently in the process of being amended to address this.
Mallard will remediate this issue once the specification's approach to solving
this is solidified.

If your `Drakesecrets` file does not exist at the root of your project,
`mallard run` supports optional `--secrets` and `-s` flags-- either of which can
be used to point `drake` to the location of your `Drakesecrets.yaml`.

__Unless it is kept outside your project source tree, do not forget to add
`Drakesecrets` to your `.gitignore` file to avoid checking sensitive information
into source control.__

## Development

If you're adventurous, Mallard can be used to build itself from source, but you
pretty quickly run into a chicken and egg problem (or is it a duck and egg
problem) and other complications.

For the moment, the most reliable way to build Mallard from source is to use the
following command:

```console
go run mage.go build
```

After the command exits, your new binary will be available inside the `./bin`
directory.

## Contributing

This project accepts contributions via GitHub pull requests. The
[Contributing](CONTRIBUTING.md) document outlines the process to help get your
contribution accepted.

## Code of Conduct

Although not a CNCF project, this project abides by the
[CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).
