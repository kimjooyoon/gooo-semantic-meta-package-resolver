#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
out=${1:-"${RUNNER_TEMP:-/tmp}/gooo-semantic-meta-package-resolver-conformance-v2"}
mkdir -p "$out"
go run ./cmd/gooo-semantic-meta-package-resolver conformance-v2 \
  --root "$root" \
  --fixtures "$root/fixtures" \
  --out "$out" >/dev/null
jq -e '
  .schema == "gooo/semantic-meta-package-resolver/cases/v2" and
  .status == "CLOSED" and .cases == 12 and .closed == 4 and .unknown == 4 and .refuted == 4 and
  .cells == 12 and .meta_activities == 12 and
  .proof == {CLOSED:4,UNKNOWN:4,REFUTED:4} and
  .indicators == {CLOSED:4,UNKNOWN:4,REFUTED:4} and
  ([.observations[] | select(.observed == "UNKNOWN" and .claim.stage != "" and .claim.step != "" and .claim.reason != "" and .claim.unknown_class != "" and .claim.next_operation != "" and (.claim.blocked_by|type)=="array")]|length) == 4 and
  ([.observations[] | select(.observed == "CLOSED" or .observed == "UNKNOWN" or .observed == "REFUTED")]|length) == 12 and
  ([.observations[] | select(.deterministic == true)]|length) == 12
' "$out/conformance-v2.json" >/dev/null
echo "conformance-v2: PASS (cells=12 meta_activities=12 cases=12 CLOSED=4 UNKNOWN=4 REFUTED=4)"
