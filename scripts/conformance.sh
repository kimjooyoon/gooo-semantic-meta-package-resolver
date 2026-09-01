#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
out=${1:-"${RUNNER_TEMP:-/tmp}/gooo-semantic-meta-package-resolver-conformance"}
mkdir -p "$out"
go run ./cmd/gooo-semantic-meta-package-resolver conformance \
  --root "$root" \
  --fixtures "$root/fixtures" \
  --out "$out" >/dev/null
jq -e '
  .schema == "gooo/semantic-meta-package-resolver/cases/v1" and
  .status == "CLOSED" and .cases == 7 and .closed == 4 and .unknown == 1 and .refuted == 2 and
  ([.observations[] | select(.observed == "CLOSED")]|length) == 4 and
  ([.observations[] | select(.observed == "UNKNOWN" and .claim.stage != "" and .claim.step != "" and .claim.reason != "" and .claim.unknown_class != "" and .claim.next_operation != "" and (.claim.blocked_by|type)=="array")]|length) == 1 and
  ([.observations[] | select(.observed == "REFUTED")]|length) == 2 and
  ([.observations[] | select(.deterministic == true)]|length) == 7
' "$out/conformance.json" >/dev/null
echo "conformance: PASS (cases=7 CLOSED=4 UNKNOWN=1 REFUTED=2)"
