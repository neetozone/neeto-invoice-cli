# Contributing to NeetoInvoice CLI

## Development setup

```bash
git clone https://github.com/neetozone/neeto-invoice-cli.git
cd neeto-invoice-cli
bin/setup
```

## Workflow

1. Branch naming: use `<issue-number>-<kebab-case-description>` or an
   organization-appropriate convention.
2. Keep commits focused. Commit messages: single-line, past tense, no
   `feat:` / `fix:` / `chore:` prefixes.
3. Run `make check` before pushing (`fmt` + `vet` + `test`).
4. Open a PR against `main`. Apply a `major` / `minor` / `patch` label
   for the next release.

## Adding commands

See [`docs/adding-commands.md`](adding-commands.md).

## Running against a local server

```bash
export NEETOINVOICE_BASE_URL=http://acme.lvh.me:8980
./neetoinvoice login --subdomain acme
```
