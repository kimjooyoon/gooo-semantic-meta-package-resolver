#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
out=${1:-"${RUNNER_TEMP:-/tmp}/gooo-semantic-meta-package-resolver-conformance-v2"}
mkdir -p "$out"
go run ./cmd/gooo-semantic-meta-package-resolver conformance-v2 \
  --root "$root" \
  --fixtures "$root/fixtures" \
  --out "$out"
jq -e '
  .schema == "gooo/semantic-meta-package-resolver/cases/v2" and
  .status == "CLOSED" and .cases == 12 and .closed == 4 and .unknown == 4 and .refuted == 4 and
  .cells == 12 and .meta_activities == 12 and
  .proof == {CLOSED:4,UNKNOWN:4,REFUTED:4} and
  .indicators == {CLOSED:4,UNKNOWN:4,REFUTED:4} and
  ([.observations[] | {id,cell_id,proof_choice,indicator_class}] | sort_by(.id)) ==
  ([
    {id:"ambiguous-export-unknown",cell_id:"ambiguous_export",proof_choice:"FOUNDATION",indicator_class:"GUARDRAIL"},
    {id:"dependency-cycle-refuted",cell_id:"dependency_cycle",proof_choice:"FOUNDATION",indicator_class:"DRIVER"},
    {id:"deterministic-diamond-convergence-closed",cell_id:"diamond_convergence",proof_choice:"COHERENCE",indicator_class:"OUTCOME"},
    {id:"exact-single-import-closed",cell_id:"exact_single",proof_choice:"FOUNDATION",indicator_class:"DRIVER"},
    {id:"explicit-fixed-point-merge-closed",cell_id:"fixed_point_merge",proof_choice:"FOUNDATION",indicator_class:"OUTCOME"},
    {id:"forged-package-identity-refuted",cell_id:"forged_package_identity",proof_choice:"REGRESSION",indicator_class:"GUARDRAIL"},
    {id:"missing-asset-digest-unknown",cell_id:"missing_asset_digest",proof_choice:"REGRESSION",indicator_class:"GUARDRAIL"},
    {id:"missing-immutable-release-unknown",cell_id:"missing_immutable_release",proof_choice:"COHERENCE",indicator_class:"DRIVER"},
    {id:"same-name-conflicting-digest-refuted",cell_id:"conflicting_digest",proof_choice:"REGRESSION",indicator_class:"DRIVER"},
    {id:"same-name-same-digest-dedupe-closed",cell_id:"same_name_same_digest",proof_choice:"REGRESSION",indicator_class:"GUARDRAIL"},
    {id:"semantic-export-conflict-refuted",cell_id:"semantic_export_conflict",proof_choice:"COHERENCE",indicator_class:"OUTCOME"},
    {id:"toolchain-mismatch-unknown",cell_id:"toolchain_mismatch",proof_choice:"COHERENCE",indicator_class:"OUTCOME"}
  ] | sort_by(.id)) and
  ([.observations[] | select((.semantic_ir_digest|test("^sha256:[0-9a-f]{64}$")) and (.machine_dossier_digest|test("^sha256:[0-9a-f]{64}$")) and (.human_dossier_digest|test("^sha256:[0-9a-f]{64}$")) and (.artifact_manifest_digest|test("^sha256:[0-9a-f]{64}$")))]|length) == 12 and
  ([.observations[] | select(.observed == "UNKNOWN" and .claim.stage != "" and .claim.step != "" and .claim.reason != "" and .claim.unknown_class != "" and .claim.next_operation != "" and (.claim.blocked_by|type)=="array")]|length) == 4 and
  ([.observations[] | select(.observed == "CLOSED" or .observed == "UNKNOWN" or .observed == "REFUTED")]|length) == 12 and
  ([.observations[] | select(.deterministic == true)]|length) == 12
' "$out/conformance-v2.json" >/dev/null
echo "conformance-v2: PASS (cells=12 meta_activities=12 cases=12 CLOSED=4 UNKNOWN=4 REFUTED=4)"
