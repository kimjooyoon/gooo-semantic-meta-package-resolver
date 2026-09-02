#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
contract="$root/contracts/resolver.gooo"
contract_v2="$root/contracts/resolver-v2.gooo"
search() {
  local pattern=$1 file=$2
  if command -v rg >/dev/null 2>&1; then rg -q "$pattern" "$file"; else grep -Eq "$pattern" "$file"; fi
}
test -s "$contract"
test -s "$contract_v2"
test -s "$root/contracts/ci-evidence-v2.schema.json"
search '^grammar source_authority=' "$contract"
search '^rule name=immutable_selection' "$contract"
search '^rule name=no_registry_latestness' "$contract"
search '^rule name=origin_ownership' "$contract"
search '^rule name=effect_safety' "$contract"
search '^fixed_denominator .*cases=7' "$contract"
search '^precedence states=REFUTED>UNKNOWN>CLOSED' "$contract"
search '^unknown_fields fields=stage,step,reason,unknown_class,next_operation,blocked_by' "$contract"
search '^fixed_denominator id=semantic-import-v2 cells=12 meta_activities=12 proof=4/4/4 indicator=4/4/4 cases=12 closed=4 unknown=4 refuted=4' "$contract_v2"
search '^merge strategy=FIXED_POINT' "$contract_v2"
search '^gate name=cross_project_required_gates value=0' "$contract_v2"
search '^rule name=cell_evidence_binding' "$contract_v2"
test "$(rg -c '^cell .* proof_choice=.* indicator_class=.* evidence_ref=' "$contract_v2")" = "12"
search '^repository identity=github.com/kimjooyoon/gooo-semantic-meta-package-resolver' "$contract_v2"
search '^release id=380317048' "$contract_v2"
search '^tag object=c250309bd20574b011e4ab9cf53a646e6fe0bf3d target=16db5f69d7b1a8ba6a0d9bb0d7e5fdb72e5ca5e1' "$contract_v2"
search '^asset id=539217648 digest=sha256:4d01025815de1155458195359fd586729310475292f53911ae2103d332be86ee' "$contract_v2"
search '^toolchain digest=sha256:6c0f056a5d26dd8157e8d6f60404f35cf765a675e8bb37e56843418b6382adde' "$contract_v2"
test "$(jq -r '.schema' "$root/fixtures/catalog.json")" = "gooo/semantic-meta-package-resolver/catalog/v1"
test "$(jq -r '.schema' "$root/fixtures/catalog.lock.json")" = "gooo/semantic-meta-package-resolver/catalog-lock/v1"
test "$(jq -r '.schema' "$root/fixtures/v2/catalog.json")" = "gooo/semantic-meta-package-resolver/catalog/v2"
test "$(jq -r '.schema' "$root/fixtures/v2/catalog.lock.json")" = "gooo/semantic-meta-package-resolver/catalog/v2-lock"
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
