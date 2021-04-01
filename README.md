# DevDrake

__DevDrake__, whose executable is abbreviated as `drake`, is a
developer-oriented command-line tool for executing DrakeSpec-compliant jobs and
pipelines using the local Docker daemon. The ability to do so offers developers
some major benefits:

* Makes it practical and convenient to define useful jobs like `test`, `lint`,
  or `build` (for example) in _one, common place_ then seamlessly reuse those
  jobs across both CI/CD pipelines and local development workflows.

* Ensures parity between CI/CD platforms and the local development environment
  by sandboxing critical work in containers with a single, common definition.

For more information about DrakeSpec-compliant jobs and pipelines, please refer
to the [DrakeSpec](https://github.com/lovethedrake/drakespec).

## THIS PROJECT HIGHLY VOLATILE!

DevDrake implements the highly volatile
[DrakeSpec](https://github.com/lovethedrake/drakespec) and, as such is, itself,
highly volatile. Users are warned that breaking changes to this tool are likely
at any point up until its eventual 1.0 release.

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

`drake` utilizes a `Drakefile.yaml` file that defines DrakeSpec-compliant jobs
and pipelines using the Drake DSL. By convention, `Drakefile.yaml` is typically
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

### Executing jobs and pipelines

To execute a single job:

```console
$ drake run <job-name>
```

To execute multiple jobs, just provide additional job names as
additional arguments:

```console
$ drake run <job-name-0> <job-name-1> ... <job-name-n>
```

Note, the command above creates an _ad hoc_ pipeline wherein execution of each
job is contingent upon successful execution of the previous job. (Concurrent
execution of jobs in a pipeline is discussed later in this section, but these
dependencies effectively preclude any concurrent job execution in such an _ad
hoc_ pipeline.) If you need to execute more complex pipelines or take advantage
of any concurrent job execution, it is recommended to explicitly create a named
pipeline for that worflow in your `Drakefile.yaml`.

Note: Because pipelines may contain multiple jobs and because jobs can involve
multiple, cooperating containers, `drake` prefixes every line of output with the
job name and container name to disambiguate its source.

To execute a named pipeline instead of a series of jobs, use either the
`--pipeline` or `-p` flag:

```console
$ drake run <pipeline-name> --pipeline
```

Note: Multiple pipelines _cannot_ be executed at once.

Pipelines are composed of jobs that MAY execute concurrently (wherever each
job's dependencies permit). However, by default, `drake` never executes jobs in
a pipeline concurrently. To enable concurrent execution of the jobs within a
pipeline, use either the `--concurrency` or `-c` flag with a value greater than
`1`.

```console
$ drake run <pipeline-name> --pipeline --concurrency 2
```

Note: The `--concurrency` or `-c` flag only tunes the _maximum_ allowed
concurrency at any given moment during pipeline execution. Dependencies between
jobs in the pipeline may sometimes prevent this maximum degree of concurrency
from being achieved. For instance, given a pipeline containing two jobs, with
the second dependent on successful completion of the first, the jobs in the
pipeline can only ever execute in sequence, regardless of how hight a value is
given to the `--concurrency` or `-c` flag.

### Using secrets

Jobs may, at times, require access to sensitive information that should not be
checked into source control. This means that including these values anywhere in
the job definitions in your `Drakefile.yaml` is not possible.

To permit `drake` access to such secrets, add key/value pairs, one per line,
with key and value delimited by `=` to a `Drakesecrets` file. By convention,
this file is stored at the root of your project.

For example:

```
DOCKER_REGISTRY_NAMESPACE=krancour
DOCKER_PASSWORD=y7o9htGWkiVBbsqFjmeh
GITHUB_TOKEN=7ddbd7a560b5b545b4bbeae0c006249adba0456a
```

If it exists at the root of your project, all variants of the `drake run`
sub-command will, by default, load secrets from the `Drakesecrets` file and
expose them as environment variables within every container of every job that it
executes.

If your `Drakesecrets` file does not exist at the root of your project, `drake
run` supports optional `--secrets` and `-s` flags-- either of which can be used
to point `drake` to the location of your `Drakesecrets`.

__Unless it is kept outside your project source tree, do not forget to add
`Drakesecrets` to your `.gitignore` file to avoid checking sensitive information
into source control.__

## Development

Normally, you should download the Drake CLI from the [releases page](https://github.com/lovethedrake/devdrake/releases). If one of the Drake versions from that page doesn't work, then you can use Drake to build itself. In rare cases, you might find that you can't do that, however.

For example, if you've upgraded the spec version of the Drakefile, then an older Drake release won't be able to build the new Drakefile, so you'll need to manually build Drake from source.

To do so, run the below command. This is approximately the same command that Drake will run behind the scenes if you were to use the Drakefile to build.

```console
go run mage.go build
```

After the command exits, your built binary will be inside of `./bin`, e.g. `./bin/drake-darwin-amd64`.

## Contributing

This project accepts contributions via GitHub pull requests. The
[Contributing](CONTRIBUTING.md) document outlines the process to help get your
contribution accepted.

## Code of Conduct

Although not a CNCF project, this project abides by the
[CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).
