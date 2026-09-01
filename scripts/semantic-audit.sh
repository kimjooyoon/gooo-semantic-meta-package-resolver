#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
contract="$root/contracts/resolver.gooo"
search() {
  local pattern=$1 file=$2
  if command -v rg >/dev/null 2>&1; then rg -q "$pattern" "$file"; else grep -Eq "$pattern" "$file"; fi
}
test -s "$contract"
search '^grammar source_authority=' "$contract"
search '^rule name=immutable_selection' "$contract"
search '^rule name=no_registry_latestness' "$contract"
search '^rule name=origin_ownership' "$contract"
search '^rule name=effect_safety' "$contract"
search '^fixed_denominator .*cases=7' "$contract"
search '^precedence states=REFUTED>UNKNOWN>CLOSED' "$contract"
search '^unknown_fields fields=stage,step,reason,unknown_class,next_operation,blocked_by' "$contract"
test "$(jq -r '.schema' "$root/fixtures/catalog.json")" = "gooo/semantic-meta-package-resolver/catalog/v1"
test "$(jq -r '.schema' "$root/fixtures/catalog.lock.json")" = "gooo/semantic-meta-package-resolver/catalog-lock/v1"
forbidden=0
if command -v rg >/dev/null 2>&1; then
  rg -n 'go list -m -u|@latest|proxy.golang.org' "$root" -g '*.go' -g '*.gooo' && forbidden=1 || true
else
  grep -REn 'go list -m -u|@latest|proxy.golang.org' "$root" --include='*.go' --include='*.gooo' && forbidden=1 || true
fi
if [ "$forbidden" -eq 1 ]; then
  echo "registry latestness is forbidden in resolver source" >&2
  exit 1
fi
echo "semantic audit: PASS (authority is .gooo; registry latestness absent)"
