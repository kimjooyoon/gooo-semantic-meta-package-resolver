# Immutable release procedure

The release workflow follows GitHub's immutable-release sequence: create a
draft, upload all assets, publish the draft, then query the release API and
verify `immutable=true` plus every asset digest. Existing releases and tags are
never overwritten or deleted.

Before a release, the maintainer uses the authenticated user API to verify:

```text
GET /repos/kimjooyoon/gooo-semantic-meta-package-resolver/immutable-releases
→ {"enabled":true,...}
```

The workflow only calls that repository capability endpoint with its regular
contents token; it never queries an admin settings endpoint with
`GITHUB_TOKEN`. Enabling the repository setting is a maintainer action outside
the workflow.

The release asset is a tarball containing the exact CI evidence, generated
artifacts, release metadata, and SHA-256 sidecar. The release job validates the
annotated tag object and target commit before creating a draft. A public tag or
release collision fails closed.
