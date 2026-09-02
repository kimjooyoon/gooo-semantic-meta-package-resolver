#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 19 ]; then
  echo "usage: write-ci-evidence.sh subject compile_wall compile_rss build_wall build_rss test_wall test_rss conformance_wall conformance_rss integration_wall integration_rss tests_total tests_selected tests_executed tests_reused tests_failed tests_unknown generated_dir output" >&2
  exit 64
fi

subject=$1
compile_wall=$2; compile_rss=$3
build_wall=$4; build_rss=$5
test_wall=$6; test_rss=$7
conformance_wall=$8; conformance_rss=$9
integration_wall=${10}; integration_rss=${11}
tests_total=${12}; tests_selected=${13}; tests_executed=${14}; tests_reused=${15}; tests_failed=${16}; tests_unknown=${17}
generated_dir=${18}; output=${19}

integer() { case "$1" in (''|*[!0-9]*) echo "not an integer: $1" >&2; exit 1;; esac; }
for value in "$compile_wall" "$compile_rss" "$build_wall" "$build_rss" "$test_wall" "$test_rss" "$conformance_wall" "$conformance_rss" "$integration_wall" "$integration_rss" "$tests_total" "$tests_selected" "$tests_executed" "$tests_reused" "$tests_failed" "$tests_unknown"; do integer "$value"; done

count_files() { git ls-files -z -- "$1" | tr -cd '\0' | wc -c | tr -d ' '; }
physical_lines() {
  local total=0 file lines
  while IFS= read -r -d '' file; do
    lines=$(wc -l < "$file" | tr -d ' ')
    total=$((total + lines))
  done < <(git ls-files -z -- "$1")
  echo "$total"
}
descendant_dirs() {
  git ls-files -- "$1" | awk 'BEGIN{FS="/"} {if (NF==1) print "."; else {out=$1; for (i=2;i<NF;i++) out=out "/" $i; print out}}' | sort -u | wc -l | tr -d ' '
}
generated_files=$(find "$generated_dir" -type f -print | wc -l | tr -d ' ')
generated_bytes=$(find "$generated_dir" -type f -exec wc -c {} + | awk 'END{print $1+0}')
go_files=$(count_files '*.go'); go_lines=$(physical_lines '*.go'); go_dirs=$(descendant_dirs '*.go')
gooo_files=$(count_files '*.gooo'); gooo_lines=$(physical_lines '*.gooo'); gooo_dirs=$(descendant_dirs '*.gooo')
source_digest=$(sha256sum fixtures/cases/index.json | awk '{print $1}')
contract_digest=$(sha256sum contracts/resolver.gooo | awk '{print $1}')
runner_digest=$(sha256sum scripts/conformance.sh scripts/semantic-audit.sh | sha256sum | awk '{print $1}')
mkdir -p "$(dirname "$output")"
jq -S -n \
  --arg schema "gooo/semantic-meta-package-resolver/ci-evidence/v1" \
  --arg subject_sha "$subject" \
  --arg source_digest "sha256:$source_digest" \
  --arg contract_digest "sha256:$contract_digest" \
  --arg runner_digest "sha256:$runner_digest" \
  --argjson compile_wall "$compile_wall" --argjson compile_rss "$compile_rss" \
  --argjson build_wall "$build_wall" --argjson build_rss "$build_rss" \
  --argjson test_wall "$test_wall" --argjson test_rss "$test_rss" \
  --argjson conformance_wall "$conformance_wall" --argjson conformance_rss "$conformance_rss" \
  --argjson integration_wall "$integration_wall" --argjson integration_rss "$integration_rss" \
  --argjson tests_total "$tests_total" --argjson tests_selected "$tests_selected" --argjson tests_executed "$tests_executed" --argjson tests_reused "$tests_reused" --argjson tests_failed "$tests_failed" --argjson tests_unknown "$tests_unknown" \
  --argjson go_files "$go_files" --argjson go_lines "$go_lines" --argjson go_dirs "$go_dirs" \
  --argjson gooo_files "$gooo_files" --argjson gooo_lines "$gooo_lines" --argjson gooo_dirs "$gooo_dirs" \
  --argjson generated_files "$generated_files" --argjson generated_bytes "$generated_bytes" \
  '{schema:$schema,subject_sha:$subject_sha,
    inventory:{scope:"tracked files excluding root README.md",
      go:{physical_lines:$go_lines,descendant_dirs:$go_dirs,regular_files:$go_files,generated_files:$generated_files,generated_bytes:$generated_bytes},
      gooo:{physical_lines:$gooo_lines,descendant_dirs:$gooo_dirs,regular_files:$gooo_files,generated_files:$generated_files,generated_bytes:$generated_bytes}},
    stages:{compile:{wall_ms:$compile_wall,peak_rss_kib:$compile_rss},build:{wall_ms:$build_wall,peak_rss_kib:$build_rss},test:{wall_ms:$test_wall,peak_rss_kib:$test_rss},conformance:{wall_ms:$conformance_wall,peak_rss_kib:$conformance_rss},integration:{wall_ms:$integration_wall,peak_rss_kib:$integration_rss}},
    tests:{total:$tests_total,selected:$tests_selected,executed:$tests_executed,reused:$tests_reused,failed:$tests_failed,unknown:$tests_unknown},
    local_executions:{test:0,build:0,vet:0,conformance:0,integration:0},repository_writes:0,
    improvement:{state:"UNKNOWN",scenario:"fixed-seven-case-resolution-replay",source:$source_digest,contract:$contract_digest,toolchain:"go1.27.x",runner:$runner_digest,before:null,after:null,reason:"EXACT_BEFORE_AFTER_IDENTITY_PAIR_NOT_AVAILABLE"}}' > "$output"
