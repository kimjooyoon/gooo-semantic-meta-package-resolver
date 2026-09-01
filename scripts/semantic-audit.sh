#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
contract="$root/contracts/resolver.gooo"
test -s "$contract"
rg -q '^grammar source_authority=' "$contract"
rg -q '^rule name=immutable_selection' "$contract"
rg -q '^rule name=no_registry_latestness' "$contract"
rg -q '^rule name=origin_ownership' "$contract"
rg -q '^rule name=effect_safety' "$contract"
rg -q '^fixed_denominator .*cases=7' "$contract"
rg -q '^precedence states=REFUTED>UNKNOWN>CLOSED' "$contract"
rg -q '^unknown_fields fields=stage,step,reason,unknown_class,next_operation,blocked_by' "$contract"
test "$(jq -r '.schema' "$root/fixtures/catalog.json")" = "gooo/semantic-meta-package-resolver/catalog/v1"
test "$(jq -r '.schema' "$root/fixtures/catalog.lock.json")" = "gooo/semantic-meta-package-resolver/catalog-lock/v1"
if rg -n 'go list -m -u|@latest|proxy.golang.org' "$root" -g '*.go' -g '*.gooo'; then
  echo "registry latestness is forbidden in resolver source" >&2
  exit 1
fi
echo "semantic audit: PASS (authority is .gooo; registry latestness absent)"
