# NeetoInvoice CLI

Manage clients, invoices, projects and time entries from the terminal.

<!-- neeto-cli-commons:installation:start -->
## Installation

### macOS / Linux

**Homebrew (recommended on macOS):**

```bash
brew install neetozone/tap/neetoinvoice
```

**Shell script:**

```bash
curl -fsSL https://neeto-downloads.s3.amazonaws.com/cli/NeetoInvoice/latest/install.sh | sh
```

This verifies the download's SHA-256 checksum against the published `SHA256SUMS`,
then installs to `/usr/local/bin` (may prompt for sudo). Set `NEETOINVOICE_INSTALL_DIR`
to a directory you own to install without sudo.

### Windows

**PowerShell:**

```powershell
irm https://neeto-downloads.s3.amazonaws.com/cli/NeetoInvoice/latest/install.ps1 | iex
```

**Command Prompt (CMD):**

```cmd
curl -fsSL https://neeto-downloads.s3.amazonaws.com/cli/NeetoInvoice/latest/install.cmd -o install.cmd && install.cmd
```

Both verify the download's SHA-256 checksum before installing to
`%LOCALAPPDATA%\Programs\neetoinvoice` and adding it to your user PATH. Set
`NEETOINVOICE_INSTALL_DIR` to install somewhere else.
<!-- neeto-cli-commons:installation:end -->

<!-- neeto-cli-commons:verify-installation:start -->
### Verify installation

```bash
neetoinvoice --help
```
<!-- neeto-cli-commons:verify-installation:end -->

<!-- neeto-cli-commons:ai-coding-assistants:start -->
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
run these commands from the root of the project the assistant works in. Re-run
them after every upgrade: `setup cursor` and `setup windsurf` overwrite their rule
file, while `setup copilot`, `setup gemini` and `setup codex` keep the existing
content of their file and replace only the NeetoInvoice section instead of adding a
duplicate.
<!-- neeto-cli-commons:ai-coding-assistants:end -->

<!-- neeto-cli-commons:prerequisites:start -->
## Prerequisites (development)

- [Go](https://go.dev/dl/) 1.26.1+
- Access to a NeetoInvoice organization
<!-- neeto-cli-commons:prerequisites:end -->

## Development

```bash
git clone https://github.com/neetozone/neeto-invoice-cli.git
cd neeto-invoice-cli
bin/setup
```

This installs Go dependencies, golangci-lint, configures git hooks, and builds the binary.

<!-- neeto-cli-commons:make-targets:start -->
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
<!-- neeto-cli-commons:make-targets:end -->

### Pointing to a local or staging server

Set `NEETOINVOICE_BASE_URL` to override the default `https://<subdomain>.neetoinvoice.com`:

```bash
export NEETOINVOICE_BASE_URL=http://acme.lvh.me:8980
neetoinvoice login --subdomain acme
```

<!-- neeto-cli-commons:global-flags:start -->
## Global flags

Every command accepts:

| Flag | Description |
|---|---|
| `--subdomain <name>` | Which logged-in subdomain to use (required when multiple are logged in). |
| `--json` | Force JSON envelope output. |
| `--quiet` | Emit raw data only. Action commands print just the identifier; `delete` prints `success`. |
| `--toon` | TOON (Token-Optimized Output Notation) — compact format for LLMs. |
| `--verbose` | Expand every field of a record instead of a table. |
<!-- neeto-cli-commons:global-flags:end -->

## Adding product-specific commands

See [`docs/adding-commands.md`](docs/adding-commands.md) for the step-by-step
workflow for adding new resource commands that use the built-in auth, HTTP
client, and output helpers.

Quick API wrapper reference: [`docs/api-wrapper-reference.md`](docs/api-wrapper-reference.md).

<!-- neeto-cli-commons:release:start -->
## Release

Releases are cut by the CI pipeline defined in `.neetoci/release.yml`.

Merging a PR with a `major`, `minor`, or `patch` label to `main` triggers the shared release script from `neeto-cli-commons`. The script:

* Bumps and tags `VERSION`
* Runs GoReleaser
* Uploads artifacts to `s3://neeto-downloads/cli/NeetoInvoice/`
* Updates the Homebrew tap (`neetozone/tap`)
* Pushes the version bump commit to `main`
<!-- neeto-cli-commons:release:end -->
