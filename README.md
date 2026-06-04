# govenv

`govenv` is a small CLI that automates per-project environment setup for Go modules. It builds your module's binary into a dedicated env directory and generates an `activate` script that puts that binary on your `PATH` — the way Python's `venv` puts a project-local `python` interpreter on your `PATH`.

## Why this exists

Coming from Python, the local dev loop is unbeatable: `python -m venv .venv && source .venv/bin/activate`, and now `python`, `pip`, and your project's entry points all resolve to a project-local sandbox. You can run, test, and iterate without thinking about which copy of which tool you're invoking.

Go does not really have this. The standard answers are some mix of `go run ./...`, `go install`, manually managing `$GOBIN`, or shell aliases pointing at `go build` outputs. None of it composes the way `venv` does, and switching between projects that produce same-named binaries is awkward.

`govenv` is an attempt to port the `venv` ergonomics to Go: one command to build the project's binary into an isolated env dir, one `source .../activate` to put it on your `PATH`, one `govenv-deactivate` to back out cleanly.

## Install

```sh
go install github.com/kenk667/govenv@latest
```

Run `govenv --help` for an overview of commands and options, or `govenv <command> -h` (e.g. `govenv list-go -h`) for command-specific help. `govenv --version` (or `-v`) prints the installed version.

### Man page

The repo ships a `govenv.1` man page. To install it, copy it into a directory on your `MANPATH`:

```sh
# system-wide (may need sudo)
install -m 0644 govenv.1 /usr/local/share/man/man1/govenv.1

# or per-user
mkdir -p ~/.local/share/man/man1
install -m 0644 govenv.1 ~/.local/share/man/man1/govenv.1   # ensure ~/.local/share/man is on MANPATH
```

Then `man govenv`. You can also read it straight from the repo without installing: `man ./govenv.1`.

## Usage

Run from a Go module directory (one containing `go.mod`):

```sh
govenv .venv
source .venv/activate
```

That builds the module's binary (named after the `module` directive in `go.mod`) into `.venv/`, then `source .venv/activate` puts `.venv/` at the front of your `PATH` and defines a `govenv-deactivate` command in your shell.

```sh
govenv-deactivate    # restore PATH and unset govenv vars
```

`govenv-deactivate` is a shell function that `source .venv/activate` defines in your current shell (just like Python venv's `deactivate`). It restores `PATH`, unsets the govenv variables, and removes itself. It's named `govenv-deactivate` rather than `deactivate` on purpose, so it never collides with Python venv's `deactivate` when both environments are active in the same shell.

> Deactivation has to be a shell function (not a `govenv` subcommand or flag) because a binary runs in a child process and cannot modify the shell that launched it — that's also why activation uses `source`. So `govenv-deactivate` lives in your shell, defined by `activate`, and exists only while an env is active.

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

While the env is active, `go`, `gofmt`, and friends resolve to the chosen toolchain. `govenv-deactivate` restores the original `PATH`.

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
   - Defines a `govenv-deactivate` shell function that restores `PATH`, unsets the `GOVENV*` variables, and removes itself. The unique name avoids colliding with Python venv's `deactivate`.

There is no daemon, no cache, no global state — just a directory with a binary and a shell script.

## Releasing

`govenv` is distributed via `go install`, so a release is just a git tag. `go install github.com/kenk667/govenv@latest` resolves to the **highest semver tag — not the newest commit on `main`**. Merging a PR therefore ships nothing to users on its own; you must cut a new tag.

```sh
git tag -a vX.Y.Z <commit> -m "vX.Y.Z — summary"
git push origin vX.Y.Z
gh release create vX.Y.Z --title vX.Y.Z --notes "release notes"
```

Bump per [semver](https://semver.org/): patch (`v0.1.x`) for fixes, minor (`v0.2.0`) for new flags/features, major (`v1.0.0`) once the CLI is stable.

The public Go module proxy caches its `@latest` answer for a short while (minutes to ~30 min) after a new tag is pushed, so `@latest` may briefly still resolve to the previous version. To bypass the lag, install the exact tag (`go install github.com/kenk667/govenv@vX.Y.Z`) or skip the proxy (`GOPROXY=direct go install github.com/kenk667/govenv@latest`).
