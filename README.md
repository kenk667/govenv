# govenv

`govenv` is a small CLI that automates per-project environment setup for Go modules. It builds your module's binary into a dedicated env directory and generates an `activate` script that puts that binary on your `PATH` — the way Python's `venv` puts a project-local `python` interpreter on your `PATH`.

## Why this exists

Coming from Python, the local dev loop is unbeatable: `python -m venv .venv && source .venv/bin/activate`, and now `python`, `pip`, and your project's entry points all resolve to a project-local sandbox. You can run, test, and iterate without thinking about which copy of which tool you're invoking.

Go does not really have this. The standard answers are some mix of `go run ./...`, `go install`, manually managing `$GOBIN`, or shell aliases pointing at `go build` outputs. None of it composes the way `venv` does, and switching between projects that produce same-named binaries is awkward.

`govenv` is an attempt to port the `venv` ergonomics to Go: one command to build the project's binary into an isolated env dir, one `source .../activate` to put it on your `PATH`, one `eval "$(govenv deactivate)"` to back out cleanly.

## Install

```sh
go install github.com/kenk667/govenv@latest
```

## Usage

Run from a Go module directory (one containing `go.mod`):

```sh
govenv .venv
source .venv/activate
```

That builds the module's binary (named after the `module` directive in `go.mod`) into `.venv/`, then `source .venv/activate` puts `.venv/` at the front of your `PATH`.

```sh
eval "$(govenv deactivate)"    # restore PATH and unset govenv vars
```

`deactivate` is a subcommand that prints shell code to restore your environment; `eval` applies it to the current shell. It is intentionally *not* a shell function named `deactivate`, so it won't collide with Python venv's `deactivate`.

### Flags

| Flag | Description |
| --- | --- |
| `--clear` | Delete the env directory before creating it. |
| `--upgrade` | Rebuild and reinstall the binary into an existing env. |
| `--prompt <name>` | Override the name shown on activation (defaults to the module's binary name). |
| `--symlinks` | Symlink the binary into the env dir. |
| `--copies` | Copy the binary into the env dir (default). |
| `--go <version>` | Use a specific Go toolchain installed under `~/sdk/` (e.g. `1.22.3`). |

### Pinning a Go version

`govenv` does not download Go for you — it detects toolchains the Go team's own installer puts under `~/sdk/`. To install a version:

```sh
go install golang.org/dl/go1.22.3@latest
go1.22.3 download
```

That places the toolchain at `~/sdk/go1.22.3/`. Then bind it to an env at create time:

```sh
govenv --go 1.22.3 .venv
source .venv/activate
go version    # go1.22.3
```

While the env is active, `go`, `gofmt`, and friends resolve to the chosen toolchain. `eval "$(govenv deactivate)"` restores the original `PATH`.

List detected toolchains:

```sh
govenv list-go
```

### Typical workflows

Create a fresh env:

```sh
govenv .venv
source .venv/activate
```

Rebuild after code changes:

```sh
govenv --upgrade .venv
```

Wipe and recreate:

```sh
govenv --clear .venv
```

## How it works

1. Reads `go.mod` in the current directory to determine the module path; the binary is named after `filepath.Base(module)`.
2. Runs `go build -o <ENV_DIR>/<bin>` (or builds to a tempfile and symlinks if `--symlinks` is set).
3. Writes `<ENV_DIR>/activate`, a shell script that:
   - Saves the current `PATH` into `_GOVENV_OLD_PATH`.
   - Exports `GOVENV`, `GOVENV_PROMPT`, and `GOVENV_GO_ROOT` (the last is empty unless `--go` was set).
   - Prepends `<ENV_DIR>` to `PATH`. If `GOVENV_GO_ROOT` is set, also prepends `$GOVENV_GO_ROOT/bin` ahead of it so `go`, `gofmt`, etc. resolve to the chosen toolchain.

To exit, `eval "$(govenv deactivate)"` reads `_GOVENV_OLD_PATH` and the `GOVENV*` variables from the environment and prints shell code that restores `PATH` and unsets them. It's a subcommand rather than a sourced shell function so it never shadows (or gets shadowed by) Python venv's `deactivate`.

There is no daemon, no cache, no global state — just a directory with a binary and a shell script.
