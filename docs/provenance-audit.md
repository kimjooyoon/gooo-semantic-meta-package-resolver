# Provenance audit record

This record preserves the release-line exception and is not part of semantic
resolution. The audit range is inclusive and was inspected from the GitHub
commit and pull-request APIs.

```text
range=78bed26bb4a6d3392aa3e4511feb0773418aa1be..cd11ca65135fc9eefe74263a70abbd4bb4c90f2e
commit_count=2
commit.1=78bed26bb4a6d3392aa3e4511feb0773418aa1be|Implement semantic meta package resolver (#1)|PR#1
commit.2=cd11ca65135fc9eefe74263a70abbd4bb4c90f2e|ci: use maintainer token for immutable capability check|PR:none
operator_process=REFUTED
operator_process_commit=cd11ca65135fc9eefe74263a70abbd4bb4c90f2e
operator_process_reason=direct post-merge push to main
v0.1.0=tag-preserved|release-absent|failed-workflow-33484531613
v0.1.1=tag-preserved|immutable-release-published
v0.1.2=tag-preserved|release-id-380317048|immutable=true|asset-id-539217648|asset-digest=sha256:4d01025815de1155458195359fd586729310475292f53911ae2103d332be86ee|target=16db5f69d7b1a8ba6a0d9bb0d7e5fdb72e5ca5e1|tag-object=c250309bd20574b011e4ab9cf53a646e6fe0bf3d
v2_gate=cross_project_required_gates=0|pr-actions-authoritative|release-job-draft-first-only
```

Neither public tag was overwritten or deleted. This audit file does not alter
the fixed seven-case vector, fixture catalog digest, or package source digests.
