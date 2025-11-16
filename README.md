# Refdir - Go linter that can enforce reference-based ordering of definitions in a _file_

[![Go Report Card](https://goreportcard.com/badge/github.com/ppipada/refdir)](https://goreportcard.com/report/github.com/ppipada/refdir)
[![lint](https://github.com/ppipada/refdir/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/ppipada/refdir/actions/workflows/lint.yml)
[![test](https://github.com/ppipada/refdir/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/ppipada/refdir/actions/workflows/test.yml)

This linter is a maintained fork of the [refdir linter by @devnev](https://github.com/devnev/refdir). It has the below bug fixes/enhancements as of November 2025:

- Don't report recursive functions as an issue. Original [issue](https://github.com/devnev/refdir/issues/10) with [PR](https://github.com/devnev/refdir/pull/11)
- Respect Ignore checks.
- Interface selections are treated as type references rather than function references. Avoids logical contradiction wrt interface type definition and reference inside same file.
- Lesser noise for universal scope identifiers.
- Working `golangci-lint` custom module plugin for version > 2.
- Chores: Stricter `golangci-lint` config compliant code; `taskfile.dev` tasks; github action integration, vscode settings folders, updated and pinned dependencies/tools; improved readme.

**Disclaimer**: false positives; practically this is useful for "exploration" rather than for "enforcement".

> **Vertical Ordering**
>
> In general we want function call dependencies to point in the downward direction. That is, a function that is called should be bellow a function that does the calling. This creates a nice flow down the source code module from the high level to low level.
>
> As in newspaper articles, we expect the most important concepts to come first, and we expect them to be expressed with the least amount of polluting detail. We expect the low-level details to come last. This allows us to skim source files, getting the gist from the first few functions, without having to immerge ourselves in the details.
>
> -- Clean Code, Chapter 5, p84, Robert C. Martin, 2009

## Usage

### Golangci-lint plugin

- The subpackage `golangci-lint` provides a [module plugin](https://golangci-lint.run/plugins/module-plugins) for golangci-lint.
- Example [.custom-gcl.yml](./golangci-lint/.custom-gcl.yml) and [.golangci.yml](./golangci-lint/.golangci.yml) configuration files are provided as a basis for use in your own project.

### Go analysis library

- Use `github.com/ppipada/refdir/analysis/refdir.Analyzer` as per `go/analysis` [docs](<(https://pkg.go.dev/golang.org/x/tools/go/analysis)>) to integrate `refdir` in a custom analysis binary.

### Standalone

```bash
go install github.com/ppipada/refdir@latest
refdir ./...
```

- For each reference type (`func`, `type`, `recvtype`, `var`, `const`) there is a flag `--${type}-dir=[up|down|ignore]` to configure the required direction of references of that type.

- Meaning of directions:

  - up: use must be after the declaration (declare above use).
  - down: use may be before the declaration (call/use first, define later).
  - ignore: skip checks for that kind.

- Options

  - `--func-dir={down|up|ignore}`

    - What: References to functions and concrete methods (calls, values).
    - Note: Interface method selections (i.M) are not func refs; they’re treated as Type refs.
    - Default (recommended): down

  - `--type-dir={down|up|ignore}`

    - What: References to named types (in signatures, conversions, literals, etc.) and interface method selections (i.M).
    - Excludes: The receiver type in a method declaration (that’s RecvType).
    - Default (recommended): up

  - `--recvtype-dir={down|up|ignore}`

    - What: The receiver type name in a method declaration (the T in func (t T) M()).
    - Counted once per method; other mentions of T inside that method are ignored.
    - Default (recommended): up

  - `--var-dir={down|up|ignore}`

    - What: References to variables.
    - Excludes: Struct fields and inner-scope vars.
    - Default (recommended): up

  - `--const-dir={down|up|ignore}`

    - What: References to constants.
    - Excludes: Inner-scope consts.
    - Default (recommended): up

  - `--verbose`

    - What: Include informational messages (skips, reasons, positions).
    - Default: false

  - `--color`
    - What: Colorize output (OK/info/error).
    - Default: true

## Known limitations

- Transitive recursion is reported as an issue in either direction. i.e., func A -> func B -> func A. A sample of that is present in this [test](./analysis/refdir/testdata/analysistest/defaultdirs/func_recursive.go).
- Promoted interface methods via embedding: `t.M()` where `M` comes from an embedded interface field won’t be reclassified as a `Type` reference to that interface.
- Type parameters: Calls through type-parameter receivers are skipped. There isn’t a meaningful per-file declaration position to compare against.

## Example

![Code graph](code-dep-viz.png)

![Output example](output-color.png)
