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

- 0 subdomains authenticated → every credential-using command errors with
  "Not authenticated. Run 'neetoinvoice login' to authenticate.".
- 1 subdomain authenticated → that one is the implicit default; `--subdomain`
  may be omitted.
- 2+ subdomains authenticated → **`--subdomain <name>` is required** on every
  credential-using command, including `doctor`. The error lists every
  authenticated subdomain so the agent can offer a choice.

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
| `setup cursor` / `windsurf` / `copilot` / `gemini` / `codex` | Write NeetoInvoice rule files into the current project directory; re-run after an upgrade to refresh them. |

## Environment variable override

Set `NEETOINVOICE_BASE_URL` to point the CLI at a staging or local server:

```bash
export NEETOINVOICE_BASE_URL=http://acme.lvh.me:8980
neetoinvoice login --subdomain acme
```

## Error surface

Every command exits non-zero on failure and writes a single-line message to
stderr. Common errors the agent should expect:

- `Not authenticated. Run 'neetoinvoice login' to authenticate.` — empty credential store.
- `Multiple subdomains authenticated (acme, beta); specify --subdomain.` — pick one.
- `Not authenticated for "foo". Authenticated subdomains: acme, beta.` — bad `--subdomain`.
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
| `invoices list` | List invoices, newest first; filter by `--client`, `--status`, `--start-date`/`--end-date` (issue dates, YYYY-MM-DD) or `--overdue`; rows include grand_total, amount_paid, amount_due and days_overdue. |
| `invoices send <invoice>` | Email an invoice to the client's recipients; all current recipients unless repeatable `--recipient-id` narrows it. Requires `--client`; optional `--subject`, `--message`, `--attach-pdf`, `--send-me-a-copy`, and `--user-email` (only with an API key). `<invoice>` is the number, identifier or ID. |
| `invoices create` | Generate an invoice for a client. Requires `--client`, `--user-email` and `--number`; line items (`invoice_time_entries`/`invoice_services`) are passed via `--json-file <file.json>`, which can also supply `--number`. Dates are YYYY-MM-DD. Taxes (`tax_details`) and expenses (`invoice_expenses` with a `category` name) go in --json-file; --time-entry-id bills unbilled entries. With `--send-email` and no `invoice_details.recipient_ids`, every current recipient is emailed. |
| `projects list/create/show/update` | Manage projects. `list` filters by `--client-id`, `--user-email`, `--status` or `--name`; `create` requires `--name`, `--client-id` (the client record ID), at least one `--task`, and `--user-email`. Billing methods: `hourly_project_rate`, `hourly_person_rate`, `hourly_task_rate`, `fixed_price_project`. |
| `project-users list/create/update/delete` | Manage users on a project. All take `--project <project-identifier>`; roles are `regular_user`/`project_manager`. `create` takes exactly one of `--user-id` or `--user-email`. |
| `time-entries list/create/update/delete` | List unbilled time entries for a project, or log, edit and remove time. `list`/`create` take `--client` and `--project`; `list` can additionally filter to a single task with `--task-id`; `create` also needs `--task-id`, `--user-email`, `--recorded-on` (YYYY-MM-DD), and `--hours`. `update`/`delete` take the time-entry ID and require `--user-email`, the user the entry belongs to. Billed entries cannot be changed or deleted. `update` also takes `--task-id` to move the entry to another task on the same project. Pass `--is-override` to write to a date autolock has closed. |
| `team-members list/show/create/update/delete` | Manage organization members. `create` takes repeatable `--email` plus `--role`; `show`/`update`/`delete` take the team-member ID. |
| `monthly-ptos list/update-earned` | Read the monthly PTO report, or set PTO earned. Both need `--month` and `--year`; `update-earned` takes repeatable `--email` and reports `updated`/`skipped` per address. |
| `reports payroll-summary/timesheet-summary/missing-entries` | Read-only reports over logged time. `payroll-summary` needs `--month`/`--year`; the other two need `--start-date`/`--end-date`, and `missing-entries` takes an optional `--user-email`; omit it for the whole team. |

`clients list`, `projects list`, `project-users list`, `invoices list`,
`team-members list` and `time-entries list` are paginated (`--page`,
`--page-size`, max 100) and carry a `pagination` block in the envelope.
`monthly-ptos list` and the reports return the full set. PTO figures are hours.

ID conventions: clients and projects accept the identifier or the ID
everywhere; invoices accept the number, identifier or ID. `clients create` and
project `--client-id` use the record `id` (UUID). Recipient, project-user,
task, team-member, and time-entry IDs are UUIDs.
