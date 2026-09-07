#!/bin/sh
# NeetoInvoice CLI — session-start hook for Claude Code
# Lightweight auth liveness check. Always exits 0 (informational).

if ! command -v neetoinvoice >/dev/null 2>&1; then
  echo "NeetoInvoice CLI is not installed or not on PATH."
  exit 0
fi

if neetoinvoice whoami >/dev/null 2>&1; then
  echo "NeetoInvoice plugin active."
else
  echo "NeetoInvoice CLI installed but not authenticated. Run 'neetoinvoice login' to authenticate."
fi

exit 0
