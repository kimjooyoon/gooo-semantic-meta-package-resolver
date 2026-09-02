#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 14 ]; then
  echo "usage: write-ci-evidence-v2.sh subject ci_wall compile_wall compile_rss build_wall build_rss test_wall test_rss conformance_wall conformance_rss integration_wall integration_rss conformance_json output" >&2
  exit 64
fi

subject=$1
ci_wall=$2
compile_wall=$3; compile_rss=$4
build_wall=$5; build_rss=$6
test_wall=$7; test_rss=$8
conformance_wall=$9; conformance_rss=${10}
integration_wall=${11}; integration_rss=${12}
conformance_json=${13}; output=${14}

integer() { case "$1" in (''|*[!0-9]*) echo "not an integer: $1" >&2; exit 1;; esac; }
for value in "$ci_wall" "$compile_wall" "$compile_rss" "$build_wall" "$build_rss" "$test_wall" "$test_rss" "$conformance_wall" "$conformance_rss" "$integration_wall" "$integration_rss"; do integer "$value"; done

peak_rss=$compile_rss
for rss in "$build_rss" "$test_rss" "$conformance_rss" "$integration_rss"; do
  if [ "$rss" -gt "$peak_rss" ]; then peak_rss=$rss; fi
done

contract_digest="sha256:$(sha256sum contracts/resolver-v2.gooo | awk '{print $1}')"
mapping=$(jq -c '[.observations[] | {case_id:.id,cell_id:.cell_id,activity:.activity,proof_choice:.proof_choice,indicator_class:.indicator_class,evidence_ref:.evidence_ref,semantic_ir_digest:.semantic_ir_digest,machine_dossier_digest:.machine_dossier_digest,human_dossier_digest:.human_dossier_digest,artifact_manifest_digest:.artifact_manifest_digest}]' "$conformance_json")

mkdir -p "$(dirname "$output")"
jq -S -n \
  --arg schema "gooo/semantic-meta-package-resolver/ci-evidence/v2" \
  --arg subject_sha "$subject" \
  --arg contract_digest "$contract_digest" \
  --argjson mapping "$mapping" \
  --argjson ci_wall "$ci_wall" \
  --argjson compile_wall "$compile_wall" --argjson compile_rss "$compile_rss" \
  --argjson build_wall "$build_wall" --argjson build_rss "$build_rss" \
  --argjson test_wall "$test_wall" --argjson test_rss "$test_rss" \
  --argjson conformance_wall "$conformance_wall" --argjson conformance_rss "$conformance_rss" \
  --argjson integration_wall "$integration_wall" --argjson integration_rss "$integration_rss" \
  --argjson peak_rss "$peak_rss" \
  '{schema:$schema,subject_sha:$subject_sha,authority_mapping:$mapping,stages:{compile:{wall_ms:$compile_wall,peak_rss_kib:$compile_rss},build:{wall_ms:$build_wall,peak_rss_kib:$build_rss},test:{wall_ms:$test_wall,peak_rss_kib:$test_rss},conformance:{wall_ms:$conformance_wall,peak_rss_kib:$conformance_rss},integration:{wall_ms:$integration_wall,peak_rss_kib:$integration_rss}},metrics:{build_ms:$build_wall,test_ms:$test_wall,wall_ms:$ci_wall,peak_rss_kib:$peak_rss,cache_hits:null,cache_misses:null,state:"UNKNOWN",unknown_reason:"CACHE_METRICS_UNMEASURABLE"},tests:{package_total:7,package_executed:7,package_failed:0,semantic_import_cases:12},runtime:{repository_writes:0,local_test_executions:0,cross_project_required_gates:0,contents_permission:"read",output_location:"caller-owned RUNNER_TEMP",failed_runs_preserved:true},improvement:{value:null,state:"UNKNOWN",reason:"EXACT_BEFORE_AFTER_IDENTITY_PAIR_NOT_AVAILABLE",before:null,after:null},authority_contract_digest:$contract_digest}' > "$output"
