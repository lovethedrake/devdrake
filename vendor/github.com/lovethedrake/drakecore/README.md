# DrakeCore

[![codecov](https://codecov.io/gh/lovethedrake/drakecore/branch/master/graph/badge.svg)](https://codecov.io/gh/lovethedrake/drakecore)


__DrakeCore__ is a library of generally useful components for implementing
[drakespec](https://github.com/lovethedrake/drakespec)-compliant tools in Go.

## THIS PROJECT HIGHLY VOLATILE!

DrakeCore implements the highly volatile
[DrakeSpec](https://github.com/lovethedrake/drakespec) and, as such is, itself,
highly volatile. Users are warned that breaking changes to this software are
likely at any point up until its eventual 1.0 release.

## Use

Use the `github.com/lovethedrake/drakecore/config` package's `NewConfigFromYAML`
function to parse a `[]byte` containing DrakeSpec-compliant YAML into a rich
and (by design) immutable configuration tree.

```golang
import "github.com/lovethedrake/drakecore/config"

// ...

cfg, err := config.NewConfigFromYAML(yamlBytes)
if err != nil {
  // Handle the error
}
```

Alternatively, use the `config.NewConfigFromFile` function to read a file
containing DrakeSpec-compliant YAML from a specified file system path and parse
its contents.

```golang
import "github.com/lovethedrake/drakecore/config"

// ...

cfg, err := config.NewConfigFromFile(filePath)
if err != nil {
  // Handle the error
}
```

Navigation of the resulting configuration tree is intuitive and learning to do
so is left as an exercise for the reader.

## Contributing

This project accepts contributions via GitHub pull requests. The
[Contributing](CONTRIBUTING.md) document outlines the process to help get your
contribution accepted.

## Code of Conduct

Although not a CNCF project, this project abides by the
[CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).
