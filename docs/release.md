# Immutable release procedure

The release workflow follows GitHub's immutable-release sequence: create a
draft, upload all assets, publish the draft, then query the release API and
verify `immutable=true` plus every asset digest. Existing releases and tags are
never overwritten or deleted.

Before a release, the maintainer may use the authenticated user API to verify:

```text
GET /repos/kimjooyoon/gooo-semantic-meta-package-resolver/immutable-releases
→ {"enabled":true,...}
```

The release workflow does not call this capability or any admin settings
endpoint. Release creation, asset upload, publish, and the post-publish
`immutable=true`/digest verification use only the standard `GITHUB_TOKEN`
contents permission. Enabling the repository setting is a maintainer action
outside the workflow.

The release asset is a tarball containing the exact CI evidence, generated
artifacts, release metadata, and SHA-256 sidecar. The release job validates the
annotated tag object and target commit before creating a draft. A public tag or
release collision fails closed.

The v2 CI evidence is fetched from the successful `main` CI run for the exact
tag target before packaging. `ci-evidence-v2.json` contains all 12 case/cell
bindings and their semantic-IR, machine-dossier, human-dossier, and artifact
manifest digests. It also contains integer `build_ms`, `test_ms`, `wall_ms`,
and `peak_rss_kib` measurements. Cache hit/miss fields remain `null` and the
metrics claim remains `UNKNOWN` when the runner cannot observe cache events.
Improvement is `UNKNOWN` with a null value until an exact before/after identity
pair is available.

The v0.1.2 release is preserved as the v2 identity anchor:

```text
release_id=380317048
tag=v0.1.2
tag_object=c250309bd20574b011e4ab9cf53a646e6fe0bf3d
target=16db5f69d7b1a8ba6a0d9bb0d7e5fdb72e5ca5e1
asset_id=539217648
asset_digest=sha256:4d01025815de1155458195359fd586729310475292f53911ae2103d332be86ee
```

The next release job is PR-first: required PR Actions are the authority for
the subject, while the tag workflow only packages already-green evidence and
publishes a new draft-first immutable release. The import graph has zero
cross-project required gates; it consumes immutable release and digest
identity only.
