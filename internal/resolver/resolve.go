package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Resolver struct {
	Contract    Source
	Catalog     Catalog
	CatalogLock CatalogLock
	CatalogRoot string
}

type resolveState struct {
	root        Source
	provided    *Lockfile
	catalog     []CatalogEntry
	catalogRoot string
	selected    map[string]SelectedPackage
	resolving   map[string]bool
	visited     map[string]bool
	edges       []LinkEdge
	decisions   []Claim
	lockRefs    map[string]string
}

func ValidateContract(contract Source) error {
	if !strings.HasPrefix(contract.Schema, "gooo/contract/semantic_meta_package_resolver/") || !strings.HasSuffix(contract.Schema, "/v1") || contract.Kind != "contract" || contract.Name != "semantic_meta_package_resolver" || contract.LanguageVersion != "v1" {
		return fmt.Errorf("invalid resolver contract identity: schema=%q kind=%q name=%q language=%q", contract.Schema, contract.Kind, contract.Name, contract.LanguageVersion)
	}
	if contract.FixedDenominator != 7 || !sameStrings(contract.Precedence, Precedence) || !sameStrings(contract.UnknownFields, []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"}) {
		return fmt.Errorf("fixed denominator or resolution precedence does not match")
	}
	if len(contract.Grammar) < 5 || len(contract.Rules) < 5 || len(contract.Types) < 4 || len(contract.Capabilities) < 2 || len(contract.Effects) < 2 || len(contract.OriginRules) < 2 {
		return fmt.Errorf("contract is missing an authoritative grammar/rule/type/capability/effect/origin declaration")
	}
	return nil
}

func ValidateRoot(root Source) error {
	if root.Kind != "consumer" || root.Name == "" || root.Version != "1.0.0" {
		return fmt.Errorf("root must be a v1 consumer with version 1.0.0")
	}
	if root.Selection.Strategy != "semantic-contract-first" || root.Selection.Registry != "fixture-catalog" || root.Selection.VersionOrder != "lowest-compatible" || !root.Selection.NoLatest {
		return fmt.Errorf("root selection policy must forbid registry latestness")
	}
	if len(root.Imports) == 0 {
		return fmt.Errorf("root has no imports")
	}
	seen := map[string]bool{}
	for _, lock := range root.Locks {
		if lock.Package == "" || lock.Digest == "" || seen[lock.Package] {
			return fmt.Errorf("invalid duplicate or empty source lock")
		}
		seen[lock.Package] = true
	}
	return nil
}

func (r Resolver) Resolve(root Source, provided *Lockfile) (Resolution, error) {
	if err := ValidateContract(r.Contract); err != nil {
		return Resolution{}, err
	}
	if err := ValidateRoot(root); err != nil {
		return Resolution{}, err
	}
	state := &resolveState{
		root: root, provided: provided, catalog: r.Catalog.Entries, catalogRoot: r.CatalogRoot, selected: map[string]SelectedPackage{}, resolving: map[string]bool{}, visited: map[string]bool{}, lockRefs: map[string]string{},
	}
	for _, lock := range root.Locks {
		state.lockRefs[lock.Package] = lock.Digest
	}
	if provided != nil {
		if err := validateProvidedLock(*provided, root, r.Contract, r.CatalogLock); err != nil {
			return Resolution{}, err
		}
		for _, lock := range provided.Packages {
			if existing, ok := state.lockRefs[lock.Name]; ok && existing != lock.Digest {
				state.decisions = append(state.decisions, refutedClaim("LOCK", "CHECK_LOCK_DIGEST", "LOCK_DIGEST_CONFLICT", lock.Name))
			} else {
				state.lockRefs[lock.Name] = lock.Digest
			}
		}
	}
	imports := append([]ImportDecl(nil), root.Imports...)
	sort.Slice(imports, func(i, j int) bool { return importKey(imports[i]) < importKey(imports[j]) })
	for _, imp := range imports {
		state.resolveImport(root.Name, imp, nil)
	}
	packages := make([]PackageLock, 0, len(state.selected))
	for _, selected := range state.selected {
		packages = append(packages, selected.Signature)
	}
	sort.Slice(packages, func(i, j int) bool { return packages[i].Name < packages[j].Name })
	sort.Slice(state.edges, func(i, j int) bool {
		left, right := edgeKey(state.edges[i]), edgeKey(state.edges[j])
		return left < right
	})
	status, claim := finalClaim(state.decisions)
	lock := Lockfile{Schema: LockSchema, SourceDigest: root.SourceDigest, ContractDigest: r.Contract.SourceDigest, CatalogDigest: r.CatalogLock.CatalogDigest, Selection: root.Selection, Packages: packages}
	lock.LockDigest = lockDigest(lock)
	return Resolution{Schema: ResolutionSchema, Root: root.Name, Status: status, Claim: claim, ContractDigest: r.Contract.SourceDigest, CatalogDigest: r.CatalogLock.CatalogDigest, SourceDigest: root.SourceDigest, Packages: packages, Edges: state.edges, Decisions: state.decisions, Lock: lock}, nil
}

func (s *resolveState) resolveImport(from string, imp ImportDecl, stack []string) {
	if s.resolving[imp.Package] {
		return
	}
	if selected, ok := s.selected[imp.Package]; ok {
		if claim, compatible := importMatches(selected.Source, imp); !compatible {
			s.decisions = append(s.decisions, claim)
			return
		}
		s.addEdge(from, selected, imp)
		if !s.visited[imp.Package] {
			s.walkDependencies(selected, append(stack, imp.Package))
		}
		return
	}
	candidates := s.candidates(imp)
	if len(candidates) == 0 {
		s.decisions = append(s.decisions, refutedClaim("VERSION", "CHECK_VERSION_CONTRACT", "VERSION_CONSTRAINT_UNSATISFIED", imp.Package))
		return
	}
	lockedDigest := s.lockRefs[imp.Package]
	if lockedDigest != "" {
		filtered := make([]CatalogEntry, 0, 1)
		for _, candidate := range candidates {
			if candidate.Digest == lockedDigest {
				filtered = append(filtered, candidate)
			}
		}
		if len(filtered) == 0 {
			s.decisions = append(s.decisions, refutedClaim("LOCK", "CHECK_LOCK_DIGEST", "LOCK_DIGEST_NOT_IN_CATALOG", imp.Package))
			return
		}
		candidates = filtered
	}
	var refutations, unknowns []Claim
	for _, candidate := range candidates {
		selected, claim, ok := s.inspectCandidate(candidate, imp)
		if ok {
			s.selected[imp.Package] = selected
			s.addEdge(from, selected, imp)
			s.walkDependencies(selected, append(stack, imp.Package))
			return
		}
		if claim.State == Refuted {
			refutations = append(refutations, claim)
		}
		if claim.State == Unknown {
			unknowns = append(unknowns, claim)
		}
	}
	if len(refutations) > 0 {
		s.decisions = append(s.decisions, refutations[0])
	} else if len(unknowns) > 0 {
		s.decisions = append(s.decisions, unknowns[0])
	} else {
		s.decisions = append(s.decisions, unknownClaim("RESOLUTION", "CHECK_CANDIDATE_SET", "NO_VERIFIABLE_CANDIDATE", "NO_VERIFIABLE_CANDIDATE", "PROVIDE_A_TYPED_CATALOG_ENTRY", imp.Package))
	}
}

func (s *resolveState) walkDependencies(selected SelectedPackage, stack []string) {
	if s.visited[selected.Entry.Name] || contains(stack[:len(stack)-1], selected.Entry.Name) {
		return
	}
	s.resolving[selected.Entry.Name] = true
	imports := append([]ImportDecl(nil), selected.Source.Imports...)
	sort.Slice(imports, func(i, j int) bool { return importKey(imports[i]) < importKey(imports[j]) })
	for _, imp := range imports {
		s.resolveImport(selected.Entry.Name, imp, stack)
	}
	delete(s.resolving, selected.Entry.Name)
	s.visited[selected.Entry.Name] = true
}

func (s *resolveState) candidates(imp ImportDecl) []CatalogEntry {
	result := make([]CatalogEntry, 0)
	for _, candidate := range s.catalogEntries(imp.Package) {
		if satisfies(candidate.Version, imp.Constraint) {
			result = append(result, candidate)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		left, leftErr := parseSemver(result[i].Version)
		right, rightErr := parseSemver(result[j].Version)
		if leftErr == nil && rightErr == nil && compareSemver(left, right) != 0 {
			return compareSemver(left, right) < 0
		}
		return result[i].Digest < result[j].Digest
	})
	return result
}

func (s *resolveState) catalogEntries(name string) []CatalogEntry {
	result := make([]CatalogEntry, 0)
	for _, entry := range s.catalogForAll() {
		if entry.Name == name {
			result = append(result, entry)
		}
	}
	return result
}

func (s *resolveState) catalogForAll() []CatalogEntry {
	// Keep catalog order irrelevant: candidate sorting is the only selection order.
	return append([]CatalogEntry(nil), s.catalog...)
}

func (s *resolveState) inspectCandidate(entry CatalogEntry, imp ImportDecl) (SelectedPackage, Claim, bool) {
	path := filepath.Join(s.rootCatalogRoot(), entry.Source)
	data, err := os.ReadFile(path)
	if err != nil {
		return SelectedPackage{}, unknownClaim("CATALOG", "READ_PACKAGE_SOURCE", "PACKAGE_SOURCE_UNAVAILABLE", "SOURCE_UNAVAILABLE", "PROVIDE_CATALOG_SOURCE", entry.Name), false
	}
	source, err := ParseSource(path)
	if err != nil {
		return SelectedPackage{}, refutedClaim("PARSE", "PARSE_PACKAGE_SOURCE", "PACKAGE_SOURCE_INVALID", entry.Name), false
	}
	actualDigest := digestBytes(data)
	if entry.Digest == "" || !strings.HasPrefix(entry.Digest, "sha256:") || entry.ImmutableProof == "" || source.Immutable.Proof == "" {
		return SelectedPackage{}, unknownClaim("IMMUTABLE_PROOF", "VERIFY_CATALOG_PROOF", "IMMUTABLE_PROOF_UNAVAILABLE", "MISSING_IMMUTABLE_PROOF", "OBTAIN_IMMUTABLE_RELEASE_PROOF", entry.Name), false
	}
	if actualDigest != entry.Digest || source.Immutable.Digest != "" && source.Immutable.Digest != entry.Digest {
		return SelectedPackage{}, refutedClaim("IMMUTABLE_PROOF", "VERIFY_DIGEST_BINDING", "PACKAGE_SOURCE_DIGEST_MISMATCH", entry.Name), false
	}
	if source.Immutable.Proof != entry.ImmutableProof {
		return SelectedPackage{}, refutedClaim("IMMUTABLE_PROOF", "VERIFY_PROOF_MATCH", "IMMUTABLE_PROOF_MISMATCH", entry.Name), false
	}
	if entry.Owner != "" && source.Origin.Owner != entry.Owner {
		return SelectedPackage{}, refutedClaim("ORIGIN", "CHECK_CATALOG_OWNERSHIP", "CATALOG_ORIGIN_OWNERSHIP_MISMATCH", entry.Name), false
	}
	claim, compatible := importMatches(source, imp)
	if !compatible {
		return SelectedPackage{}, claim, false
	}
	return SelectedPackage{Entry: entry, Source: source, SourceDigest: actualDigest, Signature: packageLock(entry, source, actualDigest)}, Claim{}, true
}

func (s *resolveState) rootCatalogRoot() string {
	return s.catalogRoot
}

func importMatches(source Source, imp ImportDecl) (Claim, bool) {
	if imp.Owner != "" && source.Origin.Owner != imp.Owner {
		return refutedClaim("ORIGIN", "CHECK_ORIGIN_OWNERSHIP", "ORIGIN_OWNERSHIP_CONFLICT", source.Name), false
	}
	for _, export := range source.Exports {
		if export.SemanticID != imp.SemanticID || export.Type != imp.Type || export.Stage != imp.Stage {
			continue
		}
		if !containsAll(export.Capabilities, imp.Capabilities) {
			return refutedClaim("CAPABILITY", "CHECK_CAPABILITY_ROW", "CAPABILITY_SIGNATURE_INCOMPATIBLE", source.Name), false
		}
		if !sameStrings(sortStrings(export.Effects), sortStrings(imp.Effects)) {
			return refutedClaim("EFFECT", "CHECK_EFFECT_ROW", "EFFECT_SIGNATURE_INCOMPATIBLE", source.Name), false
		}
		return Claim{}, true
	}
	return refutedClaim("SEMANTIC", "CHECK_SEMANTIC_SIGNATURE", "SEMANTIC_SIGNATURE_INCOMPATIBLE", source.Name), false
}

func packageLock(entry CatalogEntry, source Source, sourceDigest string) PackageLock {
	typeRows, capabilityRows, effectRows := []string{}, []string{}, []string{}
	for _, export := range source.Exports {
		typeRows = append(typeRows, export.ID+":"+export.SemanticID+":"+export.Type+":"+export.Stage)
		capabilityRows = append(capabilityRows, export.ID+":"+strings.Join(sortStrings(export.Capabilities), ","))
		effectRows = append(effectRows, export.ID+":"+strings.Join(sortStrings(export.Effects), ","))
	}
	sort.Strings(typeRows)
	sort.Strings(capabilityRows)
	sort.Strings(effectRows)
	return PackageLock{Name: entry.Name, Version: entry.Version, Digest: entry.Digest, SourceDigest: sourceDigest, SemanticID: source.SemanticID, Stage: source.Stage, TypeRows: typeRows, OriginOwner: source.Origin.Owner, CapabilityRows: capabilityRows, EffectRows: effectRows, VersionContract: source.VersionContract}
}

func validateProvidedLock(lock Lockfile, root, contract Source, catalogLock CatalogLock) error {
	if lock.Schema != LockSchema || lock.SourceDigest != root.SourceDigest || lock.ContractDigest != contract.SourceDigest || lock.CatalogDigest != catalogLock.CatalogDigest || lock.LockDigest != lockDigest(lock) {
		return fmt.Errorf("provided lockfile failed immutable identity verification")
	}
	return nil
}

func lockDigest(lock Lockfile) string {
	copyOf := lock
	copyOf.LockDigest = ""
	digest, _ := digestValue(copyOf)
	return digest
}

func (s *resolveState) addEdge(from string, selected SelectedPackage, imp ImportDecl) {
	edge := LinkEdge{From: from, To: selected.Entry.Name, Constraint: imp.Constraint, SemanticID: imp.SemanticID, Type: imp.Type, Stage: imp.Stage}
	for _, existing := range s.edges {
		if edgeKey(existing) == edgeKey(edge) {
			return
		}
	}
	s.edges = append(s.edges, edge)
}

func finalClaim(decisions []Claim) (Status, Claim) {
	if len(decisions) == 0 {
		return Closed, closedClaim("SEMANTIC_SIGNATURES_VERIFIED")
	}
	ordered := append([]Claim(nil), decisions...)
	sort.SliceStable(ordered, func(i, j int) bool {
		rank := func(state Status) int {
			if state == Refuted {
				return 0
			}
			if state == Unknown {
				return 1
			}
			return 2
		}
		if rank(ordered[i].State) != rank(ordered[j].State) {
			return rank(ordered[i].State) < rank(ordered[j].State)
		}
		return ordered[i].Stage+"/"+ordered[i].Step+"/"+ordered[i].Reason < ordered[j].Stage+"/"+ordered[j].Step+"/"+ordered[j].Reason
	})
	return ordered[0].State, ordered[0]
}

func VerifyResolution(resolution Resolution) error {
	if resolution.Schema != ResolutionSchema || resolution.Root == "" || resolution.ContractDigest == "" || resolution.CatalogDigest == "" {
		return fmt.Errorf("resolution identity is incomplete")
	}
	if !resolution.Claim.Valid() || resolution.Status != resolution.Claim.State {
		return fmt.Errorf("resolution claim is invalid or does not match status")
	}
	if resolution.Lock.Schema != LockSchema || resolution.Lock.LockDigest != lockDigest(resolution.Lock) || resolution.Lock.CatalogDigest != resolution.CatalogDigest {
		return fmt.Errorf("resolution lock identity is invalid")
	}
	for _, pkg := range resolution.Packages {
		if pkg.Name == "" || pkg.Version == "" || !strings.HasPrefix(pkg.Digest, "sha256:") || pkg.SourceDigest != pkg.Digest {
			return fmt.Errorf("selected package %q has an invalid immutable identity", pkg.Name)
		}
	}
	return nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func containsAll(have, required []string) bool {
	for _, want := range required {
		found := false
		for _, got := range have {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func importKey(imp ImportDecl) string {
	return imp.Package + "|" + imp.Constraint + "|" + imp.SemanticID + "|" + imp.Type + "|" + imp.Stage
}
func edgeKey(edge LinkEdge) string {
	return edge.From + "|" + edge.To + "|" + edge.Constraint + "|" + edge.SemanticID + "|" + edge.Type + "|" + edge.Stage
}
