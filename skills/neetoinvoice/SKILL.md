---
name: neetoinvoice
description: >
  Manage NeetoInvoice from the command line.
  Use when the user asks about operations exposed by the NeetoInvoice CLI.
---

## Prerequisites

Run `neetoinvoice doctor` to check authentication and connectivity.
If not authenticated, run `neetoinvoice login`.

## Authentication & multi-subdomain

Credentials for every logged-in subdomain are stored together in
`~/.config/neetoinvoice/auth.json`. A command that talks to the API picks which
subdomain to use by these rules:

- 0 subdomains logged in → every credential-using command errors with
  "not logged in. Run 'neetoinvoice login' to authenticate".
- 1 subdomain logged in → that one is the implicit default; `--subdomain`
  may be omitted.
- 2+ subdomains logged in → **`--subdomain <name>` is required** on every
  credential-using command, including `doctor`. The error lists every
  logged-in subdomain so the agent can offer a choice.

`login` / `logout` / `whoami` have dedicated behavior:

| Command | Behavior |
|---|---|
| `neetoinvoice login --subdomain <name>` | Adds or refreshes the entry for `<name>`. No flag → prompts for the subdomain. |
| `neetoinvoice logout --subdomain <name>` | Removes that one entry. |
| `neetoinvoice logout --all` | Removes every entry. |
| `neetoinvoice logout` (no flag) | Removes the only entry if exactly one is logged in; errors if multiple. |
| `neetoinvoice whoami` | Lists every logged-in account. Marks the entry `(default)` when exactly one. |
| `neetoinvoice whoami --subdomain <name>` | Shows just that one. |

## Global flags (persistent on every command)

| Flag | Purpose |
|---|---|
| `--subdomain <name>` | Select which logged-in subdomain the command targets. Required when multiple are logged in. |
| `--json` | Force JSON envelope output even on a TTY. |
| `--quiet` | Emit only the raw payload — no envelope, no breadcrumbs. For action commands (create/update), emits just the resource identifier; `delete` emits `success`. Designed for scripting. |
| `--toon` | Emit TOON (Token Optimized Output Notation). Preferred for feeding list/show output back to an LLM; ~30–60% fewer tokens than JSON. |

Precedence if multiple are set: `--toon` > `--quiet` > `--json` > pretty.

## Output modes & response envelope

**Pretty (default on a TTY)** — tables for arrays, key-value for objects,
breadcrumbs appended. Not intended for machine consumption.

**JSON envelope** (non-TTY, or `--json`):
```json
{
  "data": <resource body>,
  "breadcrumbs": [{ "label": "List", "command": "neetoinvoice <resource> list" }],
  "pagination": {
    "current_page_number": 1,
    "total_pages": 10,
    "total_records": 250
  }
}
```
`breadcrumbs` is omitted when empty. `pagination` is present only for list
commands.

**Quiet** (`--quiet`) — `data` contents only, no envelope. For action
commands `PrintQuiet` unwraps a single-key wrapper and prints the first of
`sid` / `id` / `name`. For `delete` it prints `success`.

**TOON** (`--toon`) — same data as JSON, re-encoded into TOON. Shape is
equivalent but whitespace/keys are compressed. Parse by re-reading keys as
you would JSON.

### Pagination

List commands accept `--page` (1-indexed) and `--page-size` (max 100).
The envelope's `pagination` field always exposes:
`current_page_number`, `total_pages`, `total_records`. Agents should loop
by incrementing `--page` until `current_page_number == total_pages`.

## Discovery

The full, always-accurate command tree (including any flags added after
this skill was built) is available as JSON:

```bash
neetoinvoice commands
```

Each catalog entry has `command`, `description`, optional `flags` (with
`name`, `type`, `default`, `description`, `required`), and `subcommands`.
Use this whenever a user asks about a flag or command not covered below.

## Diagnostics & IDE setup

| Command | Purpose |
|---|---|
| `doctor` | Auth check + API reachability + version. Uses `--subdomain` when multiple are logged in. |
| `version` | Print CLI version / commit / build date. |
| `commands` | Emit the full command/flag catalog as JSON. |
| `setup claude` | Install NeetoInvoice plugin into Claude Code (`plugin.json`, hooks, this SKILL.md). |
| `setup cursor` / `windsurf` / `copilot` / `gemini` / `codex` | Write IDE-specific NeetoInvoice rule files. |

## Environment variable override

Set `NEETOINVOICE_BASE_URL` to point the CLI at a staging or local server:

```bash
export NEETOINVOICE_BASE_URL=http://acme.lvh.me:8980
neetoinvoice login --subdomain acme
```

## Error surface

Every command exits non-zero on failure and writes a single-line message to
stderr. Common errors the agent should expect:

- `not logged in. Run 'neetoinvoice login' to authenticate` — empty credential store.
- `multiple subdomains logged in (acme, beta); specify --subdomain` — pick one.
- `not logged in to "foo". Logged in subdomains: acme, beta` — bad `--subdomain`.
- `required flag(s) "xxx" not set` (from cobra) — missing required flag.
- API errors come through with the server's message body; inspect the
  JSON envelope (or the `--quiet` payload) for `error` / `errors` / `notice`
  keys and any suggestions the API returns.

## Product-specific commands

Run `neetoinvoice commands` for the full machine-readable catalog with every
flag. The resource commands are:

| Command | Description |
| --- | --- |
| `clients list/create/show/update` | Manage clients. `list` finds an identifier by `--name` or `--status`; `show`/`update` take the client identifier; `create` requires `--name`. |
| `recipients create/update/delete` | Manage a client's invoice recipients. All take `--client <client-identifier>`; `create` also needs `--name`, `--email`, and `--user-email` (the acting organization user). |
| `invoices create` | Generate an invoice for a client. Requires `--client` and `--user-email`; needs an invoice `--number` plus line items (`invoice_time_entries`/`invoice_services`) passed via `--data <file.json>`. Dates are MM/DD/YYYY. |
| `projects list/create/show/update` | Manage projects. `list` filters by `--client-id`, `--user-email` or `--status`; `create` requires `--name`, `--client-id` (the client record ID), at least one `--task`, and `--user-email`. Billing methods: `hourly_project_rate`, `hourly_person_rate`, `hourly_task_rate`, `fixed_price_project`. |
| `project-users list/create/update/delete` | Manage users on a project. All take `--project <project-identifier>`; roles are `regular_user`/`project_manager`. |
| `time-entries list/create/update/delete` | List unbilled time entries for a project, or log, edit and remove time. `list`/`create` take `--client` and `--project`; `create` also needs `--task-id`, `--user-email`, `--recorded-on` (YYYY-MM-DD), and `--hours`. `update`/`delete` take the time-entry ID. Billed entries cannot be changed or deleted. Pass `--is-override` to write to a date autolock has closed. |
| `team-members list/show/create/update/delete` | Manage organization members. `create` takes repeatable `--email` plus `--role`; `show`/`update`/`delete` take the team-member ID. |
| `forced-ptos list/create` | View and create Forced PTO entries. `list` needs `--start-date` and `--end-date`; `create` needs `--user-email` and `--date`, with `--hours` defaulting to 8. Requires a Forced PTO task on the workspace's HR project. |
| `monthly-ptos list/update-earned` | Read the monthly PTO report, or set PTO earned. Both need `--month` and `--year`; `update-earned` takes repeatable `--email` and reports `updated`/`skipped` per address. |
| `reports payroll-summary/timesheet-summary/missing-entries` | Read-only reports over logged time. `payroll-summary` needs `--month`/`--year`; the other two need `--start-date`/`--end-date`, and `missing-entries` also needs `--user-email`. |

`time-entries list` and `team-members list` paginate (`--page`, `--page-size`)
and carry a `pagination` block in the envelope; the other list commands return
the full set.

ID conventions: `clients`/`projects` are addressed by their `identifier`
(short hex string returned in `show` responses); `clients create` and
project `--client-id` use the record `id` (UUID). Recipient, project-user,
task, team-member, and time-entry IDs are UUIDs.
