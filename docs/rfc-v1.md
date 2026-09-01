# Semantic meta package resolver v1

## Scope

The resolver links typed meta-operation packages. A package declares its
exports, imports, semantic IDs, stage, version contract, capability/effect
rows, origin ownership, and immutable proof in `.gooo`. The fixture catalog
adds a byte digest and a source path. The catalog lock fixes the catalog bytes
as a whole.

The root consumer also declares a selection policy and may declare exact
package digest locks. A lock entry is an identity assertion, not a hint.

## Candidate proof order

For an import, the resolver performs these checks in order:

1. The version constraint limits candidates; it never selects by registry
   latestness.
2. The catalog entry has a `sha256:` digest and a non-empty immutable proof.
   Missing proof yields an UNKNOWN tuple.
3. The source bytes equal the catalog digest and the source proof equals the
   catalog proof. A mismatch is REFUTED.
4. The candidate's origin owner equals both catalog ownership and the import's
   requested owner when one is declared.
5. A matching export must have the requested semantic ID, type, and stage;
   capability rows are subset-compatible and effect rows must match exactly.
6. Candidates are ordered by semantic-checkable version ascending and then
   digest. This is a deterministic tie-breaker, not a latest-version query.

The generated lock records every selected package's digest and signature.
The linked IR records every dependency edge. The generated Go artifact emits a
deterministic JSON execution payload, and `execute` compiles/runs that exact
artifact in the caller's output directory.

## Resolution states

`CLOSED` means all required identity and semantic predicates were observed.
`UNKNOWN` means the graph cannot close because evidence is absent or
unsupported. `REFUTED` means an observed predicate contradicts the import or
lock. The fixed precedence is `REFUTED > UNKNOWN > CLOSED`, including when a
graph contains several independent problems.

UNKNOWN requires all six fields:

```text
stage, step, reason, unknown_class, next_operation, blocked_by
```

## Cases and limits

The fixed denominator contains exactly seven cases. The diamond case resolves
the shared `common` package once and compares output with a reversed catalog
order. The replay case regenerates every linked artifact twice and compares
bytes. No case reads a network registry.

The resolver is not a replacement for Go module version selection. The Go
module reference describes MVS as selecting the highest required module
version and using checksums for content verification. This resolver adds a
typed semantic boundary on top of that idea, but intentionally does not claim
to replace Go's build graph, toolchain compatibility, or behavior-level proof.
