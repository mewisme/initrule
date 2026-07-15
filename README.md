# initrule

CLI to install Cursor rule files (`.cursor/rules/*.mdc`) into a project. Interactive Bubble Tea multi-select, or fast non-interactive install.

Bundled rules: **codegraph**, **ponytail** (more later).

## Install

**From release** (see [Releases](https://github.com/mewisme/initrule/releases)):

```bash
# example: Linux amd64
curl -fsSL -o initrule.tar.gz https://github.com/mewisme/initrule/releases/latest/download/initrule_Linux_x86_64.tar.gz
tar -xzf initrule.tar.gz
sudo mv initrule /usr/local/bin/
```

**From source:**

```bash
go install github.com/mewisme/initrule@latest
```

## Usage

Run in the project you want to configure:

```bash
initrule                          # interactive multi-select
initrule install codegraph        # fast path
initrule i ponytail               # short alias for install
initrule install codegraph ponytail
initrule install --all            # all registered rules
initrule i -a
initrule i codegraph -u           # force reinstall codegraph binary, then hooks
initrule i -a --update
```

Keys in the TUI: `↑`/`↓` move, `space` toggle, `a` all, `enter` confirm, `q` quit.

### What each rule does

| Rule | Hooks |
|------|--------|
| `codegraph` | **preinstall:** ensure `codegraph` binary (install if missing, or with `-u`); **install:** `codegraph init -i`; **postinstall:** write embed `.cursor/rules/codegraph.mdc` |
| `ponytail` | **install:** write `.cursor/rules/ponytail.mdc` |

## Develop

```bash
go test ./...
go build -o initrule .
```

Release builds use [GoReleaser](https://goreleaser.com/). Tag and push:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## License

[MIT](LICENSE)
