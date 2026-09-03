# NeetoInvoice CLI

A command-line interface for NeetoInvoice.

## Installation

### macOS / Linux

**Homebrew (recommended on macOS):**

```bash
brew trust neetozone/tap
brew install neetozone/homebrew-tap/neetoinvoice
```

**Shell script:**

```bash
curl -fsSL https://neetoinvoice.com/cli/install.sh | sh
```

### Windows

**PowerShell:**

```powershell
irm https://neetoinvoice.com/cli/install.ps1 | iex
```

**Command Prompt (CMD):**

```cmd
curl -fsSL https://neetoinvoice.com/cli/install.cmd -o install.cmd && install.cmd
```

### Verify installation

```bash
neetoinvoice --help
```

## Prerequisites (development)

- [Go](https://go.dev/dl/) 1.26.1+
- Access to a NeetoInvoice organization

## Development

```bash
git clone https://github.com/neetozone/neeto-invoice-cli.git
cd neeto-invoice-cli
bin/setup
```

This installs Go dependencies, golangci-lint, configures git hooks, and builds the binary.

### Make targets

```bash
make build          # Builds ./neetoinvoice
make test           # Run tests
make lint           # golangci-lint
make fmt            # gofmt -w
make vet            # go vet
make check          # fmt + vet + test
make install        # Installs to /usr/local/bin
make clean          # Remove built binary
```

### Pointing to a local or staging server

Set `NEETOINVOICE_BASE_URL` to override the default `https://<subdomain>.neetoinvoice.com`:

```bash
export NEETOINVOICE_BASE_URL=http://acme.lvh.me:8980
neetoinvoice login --subdomain acme
```

## Global flags

Every command accepts:

| Flag | Description |
|---|---|
| `--subdomain <name>` | Which logged-in subdomain to use (required when multiple are logged in). |
| `--json` | Force JSON envelope output. |
| `--quiet` | Emit raw data only. Action commands print just the identifier; `delete` prints `success`. |
| `--toon` | TOON (Token-Optimized Output Notation) — compact format for LLMs. |

## Adding product-specific commands

See [`docs/adding-commands.md`](docs/adding-commands.md) for the step-by-step
workflow for adding new resource commands that use the built-in auth, HTTP
client, and output helpers.

Quick API wrapper reference: [`docs/api-wrapper-reference.md`](docs/api-wrapper-reference.md).

## Release

Releases are cut by BigBinary's CI pipeline defined in
`.neetoci/release.yml`. Merging a PR with a `major` / `minor` / `patch`
label to `main` triggers `.scripts/release.sh`, which tags the current
VERSION, runs GoReleaser, uploads artifacts to
`s3://neeto-downloads/cli/NeetoInvoice/`, updates the Homebrew tap
(`neetozone/homebrew-tap`), and opens the next-version bump PR.

## AI coding assistants

```bash
neetoinvoice setup claude      # Register plugin with Claude Code
neetoinvoice setup cursor      # Write .cursor/rules/neetoinvoice.mdc
neetoinvoice setup windsurf    # Write .windsurf/rules/neetoinvoice.md
neetoinvoice setup copilot     # Add a NeetoInvoice section to .github/copilot-instructions.md
neetoinvoice setup gemini      # Add a NeetoInvoice section to GEMINI.md
neetoinvoice setup codex       # Add a NeetoInvoice section to AGENTS.md
```

Every command except `setup claude` writes into the current project directory, so
run it from the root of the project the assistant works in. `setup copilot`,
`setup gemini` and `setup codex` keep the existing content of their file and,
when re-run after an upgrade, replace the NeetoInvoice section instead of
adding a duplicate.
