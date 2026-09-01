# Gooo Semantic Meta Package Resolver

This repository implements a typed meta-operation resolver for `.gooo`
packages. It resolves a graph by immutable content digest plus a semantic
signature: semantic ID, stage, type rows, capability rows, effect rows, origin
ownership, and version-contract compatibility. It is intentionally not a
name/latest-version package-manager clone.

The authoritative contract is the actual
[`contracts/resolver.gooo`](contracts/resolver.gooo) source. Its grammar,
rules, types, capability/effect rows, origin rules, fixed seven-case
denominator, and precedence are parsed at runtime. Go only parses, lowers,
generates, executes, and verifies those declarations.

## Resolution model

```text
root .gooo + resolver.gooo + digest-locked fixture catalog
        ↓ parse and lower
typed semantic graph
        ↓ resolve immutable proof → semantic signature → lowest compatible version
gooo.lock.json + linked.gooo + linked-ir.json + linked-artifact.go
        ↓ execute and verify
deterministic execution receipt
```

The selection policy is declared in each consumer `.gooo` source:
`semantic-contract-first`, `fixture-catalog`, `lowest-compatible`, and
`no_latest=true`. A catalog is accepted only when its own immutable digest
lock matches and every selected source's bytes match its catalog digest. A
missing immutable proof is `UNKNOWN`; effect or origin contradictions are
`REFUTED`. Resolution precedence is fixed as `REFUTED > UNKNOWN > CLOSED`.

An `UNKNOWN` claim always carries `stage`, `step`, `reason`, `unknown_class`,
`next_operation`, and `blocked_by`. The resolver never turns uncertainty into
a successful selection and never consults registry freshness or network
state.

## Fixed cases

The seven `.gooo` cases under [`fixtures/cases`](fixtures/cases) are the
conformance denominator:

| Case | Expected result |
| --- | --- |
| exact immutable lock | CLOSED |
| compatible semantic extension | CLOSED |
| deterministic diamond | CLOSED |
| version-only but effect-incompatible | REFUTED |
| origin ownership conflict | REFUTED |
| missing immutable proof | UNKNOWN |
| byte-identical replay | CLOSED |

Run the resolver with a caller-owned absolute output directory:

```sh
go run ./cmd/gooo-semantic-meta-package-resolver run \
  --source fixtures/cases/exact-lock.gooo \
  --contract contracts/resolver.gooo \
  --catalog fixtures/catalog.json \
  --catalog-lock fixtures/catalog.lock.json \
  --out /tmp/gooo-semantic-meta-package-resolver-run \
  --execute
```

The output contains the lockfile, linked `.gooo`, linked IR, generated Go
artifact, resolution, manifest, and execution receipt. Generated files are
written only below the caller-provided output path; the implementation's
repository write authority is fixed at `0`.

CI is the validation authority. It uses Go 1.27 and records exact integer
wall-time/RSS measurements for compile, build, test, conformance, and
integration. Local test/build/vet/conformance/integration execution counts are
fixed at zero. No automatic commit, push, merge, or release authority is
embedded in the resolver.

## Why semantic signatures are additional

Go's [Minimal Version Selection](https://go.dev/ref/mod#minimal-version-selection)
selects the highest required version in a module graph. It is deliberately
simple and reproducible when paired with module checksums and committed module
metadata. The resolver here addresses a different boundary: two versions can
be version-compatible while differing in semantic identity, stage, capability,
effect, or origin ownership. Conversely, a later version can be a compatible
semantic extension when the required export signature remains valid.

This resolver does not replace Go module MVS or the Go toolchain. It does not
prove behavioral equivalence, ABI compatibility, publisher intent, or
platform-specific build success. Its effect rows are a declared contract, not
a whole-program effect analysis; its immutable proof is a digest/proof
boundary, not a signature-verification service. An optional public-release
adapter may enrich a candidate, but adapter failure remains `UNKNOWN` and
cannot close a graph.

See [`docs/rfc-v1.md`](docs/rfc-v1.md), [`contracts/ci-evidence-v1.schema.json`](contracts/ci-evidence-v1.schema.json), and
[`docs/release.md`](docs/release.md) for the wire contracts and release
procedure.
