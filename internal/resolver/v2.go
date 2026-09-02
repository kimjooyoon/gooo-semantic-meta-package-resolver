package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	V2RepositoryIdentity = "github.com/kimjooyoon/gooo-semantic-meta-package-resolver"
	V2ReleaseID          = "380317048"
	V2TagObject          = "c250309bd20574b011e4ab9cf53a646e6fe0bf3d"
	V2TagTarget          = "16db5f69d7b1a8ba6a0d9bb0d7e5fdb72e5ca5e1"
	V2AssetID            = "539217648"
	V2AssetDigest        = "sha256:4d01025815de1155458195359fd586729310475292f53911ae2103d332be86ee"
	V2ToolchainVersion   = "go1.27.x"
	V2MergeStrategy      = "FIXED_POINT"
)

var (
	V2ProofChoices     = []string{"FOUNDATION", "COHERENCE", "REGRESSION"}
	V2IndicatorClasses = []string{"DRIVER", "OUTCOME", "GUARDRAIL"}
)

type v2ResolveState struct {
	root        Source
	contract    Source
	catalog     []V2CatalogEntry
	catalogRoot string
	selected    map[string]v2SelectedPackage
	resolving   map[string]bool
	visited     map[string]bool
	edges       []V2LinkEdge
	decisions   []Claim
}

type v2SelectedPackage struct {
	entry  V2CatalogEntry
	source Source
}

// V2Resolver extends the original resolver at the semantic-import boundary.
// It accepts only pinned catalog manifests and never asks a registry for
// branch, latest, or semver text.
type V2Resolver struct {
	Contract    Source
	Catalog     V2Catalog
	CatalogLock V2CatalogLock
	CatalogRoot string
}

func LoadV2Catalog(path, lockPath string) (V2Catalog, V2CatalogLock, error) {
	var catalog V2Catalog
	if err := LoadJSON(path, &catalog); err != nil {
		return V2Catalog{}, V2CatalogLock{}, err
	}
	if catalog.Schema != V2CatalogSchema {
		return V2Catalog{}, V2CatalogLock{}, fmt.Errorf("unexpected v2 catalog schema %q", catalog.Schema)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return V2Catalog{}, V2CatalogLock{}, err
	}
	var lock V2CatalogLock
	if err := LoadJSON(lockPath, &lock); err != nil {
		return V2Catalog{}, V2CatalogLock{}, err
	}
	if lock.Schema != V2CatalogSchema+"-lock" || lock.CatalogDigest == "" {
		return V2Catalog{}, V2CatalogLock{}, fmt.Errorf("invalid v2 catalog lock")
	}
	if digestBytes(data) != lock.CatalogDigest {
		return V2Catalog{}, V2CatalogLock{}, fmt.Errorf("v2 catalog immutable digest mismatch")
	}
	return catalog, lock, nil
}

func ValidateV2Contract(contract Source) error {
	if contract.Schema != "gooo/contract/source/v2" || contract.Kind != "contract" || contract.Name != "semantic_meta_package_resolver" || contract.LanguageVersion != "v2" {
		return fmt.Errorf("invalid v2 resolver contract identity")
	}
	if contract.FixedDenominator != 12 || contract.FixedCells != 12 || contract.FixedMetaActivities != 12 || contract.ProofVector != "4/4/4" || contract.IndicatorVector != "4/4/4" {
		return fmt.Errorf("v2 denominator must be exactly 12 cells, 12 meta activities, proof 4/4/4, indicators 4/4/4")
	}
	if len(contract.Cells) != 12 || len(contract.MetaActivities) != 12 || !sameStrings(contract.Precedence, Precedence) || !sameStrings(contract.UnknownFields, []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"}) {
		return fmt.Errorf("v2 contract cells, precedence, or UNKNOWN fields are incomplete")
	}
	if err := validateV2CellBindings(contract); err != nil {
		return err
	}
	if contract.MergeStrategy != V2MergeStrategy || contract.CrossProjectRequiredGates != 0 {
		return fmt.Errorf("v2 contract requires explicit FIXED_POINT and zero cross-project gates")
	}
	if contract.Identity.RepositoryIdentity != V2RepositoryIdentity || contract.Identity.ImmutableReleaseID != V2ReleaseID || contract.Identity.AnnotatedTagObject != V2TagObject || contract.Identity.AnnotatedTagTarget != V2TagTarget || contract.Identity.AssetID != V2AssetID || contract.Identity.AssetDigest != V2AssetDigest {
		return fmt.Errorf("v2 contract release identity does not match the preserved immutable v0.1.2 release")
	}
	if contract.Identity.GoToolchainDigest != toolchainDigest() {
		return fmt.Errorf("v2 contract Go toolchain digest does not match go1.27.x")
	}
	if len(contract.Grammar) < 5 || len(contract.Rules) < 10 || len(contract.Types) < 8 || len(contract.Capabilities) < 2 || len(contract.Effects) < 2 || len(contract.OriginRules) < 2 {
		return fmt.Errorf("v2 contract is missing an authoritative declaration")
	}
	return nil
}

func validateV2CellBindings(contract Source) error {
	if len(contract.CellBindings) != 12 {
		return fmt.Errorf("v2 contract must declare exactly 12 proof and indicator cell bindings")
	}
	seenCells := make(map[string]bool, len(contract.CellBindings))
	seenCases := make(map[string]bool, len(contract.CellBindings))
	proofCounts := make(map[string]int, len(V2ProofChoices))
	indicatorCounts := make(map[string]int, len(V2IndicatorClasses))
	for _, cell := range contract.CellBindings {
		if cell.CellID == "" || cell.CaseID == "" || cell.Activity == "" || cell.EvidenceRef == "" || !contains(V2ProofChoices, cell.ProofChoice) || !contains(V2IndicatorClasses, cell.IndicatorClass) {
			return fmt.Errorf("v2 cell binding %q is incomplete", cell.CellID)
		}
		if seenCells[cell.CellID] || seenCases[cell.CaseID] {
			return fmt.Errorf("v2 cell or case binding is duplicated: %q", cell.CellID)
		}
		seenCells[cell.CellID] = true
		seenCases[cell.CaseID] = true
		proofCounts[cell.ProofChoice]++
		indicatorCounts[cell.IndicatorClass]++
	}
	for _, choice := range V2ProofChoices {
		if proofCounts[choice] != 4 {
			return fmt.Errorf("v2 proof choice %s must occur exactly 4 times", choice)
		}
	}
	for _, class := range V2IndicatorClasses {
		if indicatorCounts[class] != 4 {
			return fmt.Errorf("v2 indicator class %s must occur exactly 4 times", class)
		}
	}
	return nil
}

func v2CellForRoot(contract Source, rootName string) (V2CellBinding, bool) {
	for _, cell := range contract.CellBindings {
		if cell.CellID == rootName {
			return cell, true
		}
	}
	return V2CellBinding{}, false
}

func ValidateV2Root(root Source, contract Source) error {
	if root.Kind != "consumer" || root.Name == "" || root.Version != "2.0.0" || root.LanguageVersion != "v2" {
		return fmt.Errorf("root must be a v2 consumer with version 2.0.0")
	}
	if root.Selection.Strategy != "semantic-contract-first" || root.Selection.Registry != "fixture-catalog" || root.Selection.VersionOrder != "lowest-compatible" || !root.Selection.NoLatest {
		return fmt.Errorf("v2 root selection policy must forbid registry latestness")
	}
	if root.MergeStrategy != V2MergeStrategy || len(root.Imports) == 0 {
		return fmt.Errorf("v2 root requires explicit FIXED_POINT merge and at least one import")
	}
	if _, ok := v2CellForRoot(contract, root.Name); !ok {
		return fmt.Errorf("v2 root %q has no authority cell binding", root.Name)
	}
	for _, imp := range root.Imports {
		if err := validateImportIdentity(imp, contract); err != nil {
			return err
		}
	}
	return nil
}

func validateImportIdentity(imp ImportDecl, contract Source) error {
	i := imp.Identity
	if imp.Package == "" || imp.Constraint == "" || imp.SemanticID == "" || imp.Type == "" || imp.Stage == "" {
		return fmt.Errorf("v2 import has incomplete semantic signature")
	}
	if i.RepositoryIdentity == "" || i.ImmutableReleaseID == "" || i.AnnotatedTagObject == "" || i.AnnotatedTagTarget == "" || i.AssetID == "" || i.AssetDigest == "" || i.PackageSemanticID == "" || i.ExportedSymbolSetDigest == "" || i.ContractDigest == "" || i.GoToolchainDigest == "" {
		return fmt.Errorf("v2 import %q is missing required identity fields", imp.Package)
	}
	if i.PackageSemanticID != imp.SemanticID || i.ContractDigest != contract.SourceDigest || i.RepositoryIdentity != contract.Identity.RepositoryIdentity || i.ImmutableReleaseID != contract.Identity.ImmutableReleaseID || i.AnnotatedTagObject != contract.Identity.AnnotatedTagObject || i.AnnotatedTagTarget != contract.Identity.AnnotatedTagTarget || i.AssetID != contract.Identity.AssetID || i.AssetDigest != contract.Identity.AssetDigest || i.GoToolchainDigest != contract.Identity.GoToolchainDigest {
		return fmt.Errorf("v2 import %q is not bound to the authoritative release or contract identity", imp.Package)
	}
	if !strings.HasPrefix(i.AssetDigest, "sha256:") || (imp.PackageDigest != "" && !strings.HasPrefix(imp.PackageDigest, "sha256:")) {
		return fmt.Errorf("v2 import %q has an invalid digest", imp.Package)
	}
	return nil
}

func (r V2Resolver) Resolve(root Source) (V2Resolution, error) {
	if err := ValidateV2Contract(r.Contract); err != nil {
		return V2Resolution{}, err
	}
	if err := ValidateV2Root(root, r.Contract); err != nil {
		return V2Resolution{}, err
	}
	cell, _ := v2CellForRoot(r.Contract, root.Name)
	state := &v2ResolveState{root: root, contract: r.Contract, catalog: append([]V2CatalogEntry(nil), r.Catalog.Entries...), catalogRoot: r.CatalogRoot, selected: map[string]v2SelectedPackage{}, resolving: map[string]bool{}, visited: map[string]bool{}}
	imports := append([]ImportDecl(nil), root.Imports...)
	sort.SliceStable(imports, func(i, j int) bool { return v2ImportKey(imports[i]) < v2ImportKey(imports[j]) })
	for _, imp := range imports {
		state.resolveImport(root.Name, imp, nil)
	}
	packages := make([]V2PackageManifest, 0, len(state.selected))
	for _, selected := range state.selected {
		packages = append(packages, v2PackageManifest(selected.entry, selected.source, r.Contract.SourceDigest))
	}
	sort.Slice(packages, func(i, j int) bool { return v2ManifestKey(packages[i]) < v2ManifestKey(packages[j]) })
	merged := mergeV2Exports(state.selected)
	sort.Slice(state.edges, func(i, j int) bool { return v2EdgeKey(state.edges[i]) < v2EdgeKey(state.edges[j]) })
	identityDigest, err := digestValue(struct {
		ContractDigest  string              `json:"contract_digest"`
		ToolchainDigest string              `json:"toolchain_digest"`
		Cell            V2CellBinding       `json:"cell"`
		Packages        []V2PackageManifest `json:"packages"`
	}{r.Contract.SourceDigest, r.Contract.Identity.GoToolchainDigest, cell, packages})
	if err != nil {
		return V2Resolution{}, err
	}
	status, claim := finalClaim(state.decisions)
	resolution := V2Resolution{Schema: V2ResolutionSchema, Root: root.Name, Status: status, Claim: claim, ContractDigest: r.Contract.SourceDigest, ToolchainDigest: r.Contract.Identity.GoToolchainDigest, SourceDigest: root.SourceDigest, MergeStrategy: root.MergeStrategy, Cell: cell, Packages: packages, Exports: merged, Edges: state.edges, Decisions: state.decisions, IdentityDigest: identityDigest}
	resolution.LinkedIRDigest, err = v2LinkedIRDigest(resolution)
	if err != nil {
		return V2Resolution{}, err
	}
	return resolution, nil
}

func (s *v2ResolveState) resolveImport(from string, imp ImportDecl, stack []string) {
	if contains(stack, imp.Package) {
		s.decisions = append(s.decisions, refutedClaim("GRAPH", "CHECK_DEPENDENCY_CYCLE", "DEPENDENCY_CYCLE", imp.Package))
		return
	}
	if selected, ok := s.selected[imp.Package]; ok {
		if imp.PackageDigest != "" && imp.PackageDigest != selected.entry.Digest {
			s.decisions = append(s.decisions, refutedClaim("IDENTITY", "CHECK_PACKAGE_DIGEST_CONVERGENCE", "PACKAGE_DIGEST_CONFLICT", imp.Package))
			return
		}
		if claim, compatible := v2ImportMatches(selected.source, imp); !compatible {
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
		s.decisions = append(s.decisions, refutedClaim("IDENTITY", "CHECK_PACKAGE_DIGEST_CONVERGENCE", "PACKAGE_DIGEST_NOT_FOUND", imp.Package))
		return
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
		} else if claim.State == Unknown {
			unknowns = append(unknowns, claim)
		}
	}
	if len(refutations) > 0 {
		s.decisions = append(s.decisions, refutations[0])
	} else if len(unknowns) > 0 {
		s.decisions = append(s.decisions, unknowns[0])
	} else {
		s.decisions = append(s.decisions, unknownClaim("RESOLUTION", "CHECK_PINNED_MANIFEST_SET", "NO_VERIFIABLE_PINNED_MANIFEST", "MISSING_PINNED_MANIFEST", "PROVIDE_PINNED_PACKAGE_MANIFEST", imp.Package))
	}
}

func (s *v2ResolveState) walkDependencies(selected v2SelectedPackage, stack []string) {
	if s.visited[selected.entry.Name] {
		return
	}
	s.resolving[selected.entry.Name] = true
	imports := append([]ImportDecl(nil), selected.source.Imports...)
	sort.SliceStable(imports, func(i, j int) bool { return v2ImportKey(imports[i]) < v2ImportKey(imports[j]) })
	for _, imp := range imports {
		s.resolveImport(selected.entry.Name, imp, stack)
	}
	delete(s.resolving, selected.entry.Name)
	s.visited[selected.entry.Name] = true
}

func (s *v2ResolveState) candidates(imp ImportDecl) []V2CatalogEntry {
	result := make([]V2CatalogEntry, 0)
	for _, candidate := range s.catalog {
		if candidate.Name == imp.Package && satisfies(candidate.Version, imp.Constraint) && (imp.PackageDigest == "" || candidate.Digest == imp.PackageDigest) {
			result = append(result, candidate)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		left, leftErr := parseSemver(result[i].Version)
		right, rightErr := parseSemver(result[j].Version)
		if leftErr == nil && rightErr == nil && compareSemver(left, right) != 0 {
			return compareSemver(left, right) < 0
		}
		return result[i].Digest < result[j].Digest
	})
	return result
}

func (s *v2ResolveState) inspectCandidate(entry V2CatalogEntry, imp ImportDecl) (v2SelectedPackage, Claim, bool) {
	path := filepath.Join(s.catalogRoot, entry.Source)
	data, err := os.ReadFile(path)
	if err != nil {
		return v2SelectedPackage{}, unknownClaim("MANIFEST", "READ_PINNED_PACKAGE_MANIFEST", "PACKAGE_SOURCE_UNAVAILABLE", "SOURCE_UNAVAILABLE", "PROVIDE_PINNED_PACKAGE_MANIFEST", entry.Name), false
	}
	source, err := ParseSource(path)
	if err != nil {
		return v2SelectedPackage{}, refutedClaim("MANIFEST", "PARSE_PINNED_PACKAGE_MANIFEST", "PACKAGE_MANIFEST_INVALID", entry.Name), false
	}
	if entry.ImmutableReleaseID == "" || source.Identity.ImmutableReleaseID == "" {
		return v2SelectedPackage{}, unknownClaim("IDENTITY", "VERIFY_IMMUTABLE_RELEASE", "IMMUTABLE_RELEASE_UNAVAILABLE", "MISSING_IMMUTABLE_RELEASE", "OBTAIN_IMMUTABLE_RELEASE_ID", entry.Name), false
	}
	if entry.AssetDigest == "" || source.Identity.AssetDigest == "" {
		return v2SelectedPackage{}, unknownClaim("IDENTITY", "VERIFY_RELEASE_ASSET_DIGEST", "ASSET_DIGEST_UNAVAILABLE", "MISSING_ASSET_DIGEST", "OBTAIN_RELEASE_ASSET_DIGEST", entry.Name), false
	}
	if digestBytes(data) != entry.Digest || source.SourceDigest != entry.Digest {
		return v2SelectedPackage{}, refutedClaim("IDENTITY", "VERIFY_PACKAGE_SOURCE_DIGEST", "PACKAGE_SOURCE_DIGEST_MISMATCH", entry.Name), false
	}
	if !v2EntryIdentityMatchesSource(entry, source) || entry.PackageSemanticID != source.SemanticID || entry.PackageSemanticID == "" {
		return v2SelectedPackage{}, refutedClaim("IDENTITY", "VERIFY_PACKAGE_IDENTITY", "FORGED_PACKAGE_IDENTITY", entry.Name), false
	}
	if entry.GoToolchainDigest == "" || source.Identity.GoToolchainDigest == "" || entry.GoToolchainDigest != s.contract.Identity.GoToolchainDigest || source.Identity.GoToolchainDigest != s.contract.Identity.GoToolchainDigest {
		return v2SelectedPackage{}, unknownClaim("IDENTITY", "VERIFY_GO_TOOLCHAIN_DIGEST", "TOOLCHAIN_MISMATCH", "TOOLCHAIN_UNSUPPORTED", "PROVIDE_GO1_27_TOOLCHAIN_DIGEST", entry.Name), false
	}
	if entry.ContractDigest == "" || source.Identity.ContractDigest == "" || entry.ContractDigest != s.contract.SourceDigest || source.Identity.ContractDigest != s.contract.SourceDigest {
		return v2SelectedPackage{}, unknownClaim("IDENTITY", "VERIFY_CONTRACT_DIGEST", "CONTRACT_DIGEST_UNAVAILABLE", "CONTRACT_EVIDENCE_UNAVAILABLE", "PROVIDE_MATCHING_CONTRACT_DIGEST", entry.Name), false
	}
	if source.SymbolSetDigest == "" || entry.ExportedSymbolSetDigest == "" || source.SymbolSetDigest != entry.ExportedSymbolSetDigest || exportedSymbolSetDigest(source) != entry.ExportedSymbolSetDigest {
		return v2SelectedPackage{}, refutedClaim("EXPORT", "VERIFY_EXPORTED_SYMBOL_SET_DIGEST", "EXPORTED_SYMBOL_SET_DIGEST_MISMATCH", entry.Name), false
	}
	if !identityMatchesImport(entry, imp) {
		return v2SelectedPackage{}, refutedClaim("IDENTITY", "CHECK_IMPORT_IDENTITY", "IMPORT_IDENTITY_CONFLICT", entry.Name), false
	}
	claim, compatible := v2ImportMatches(source, imp)
	if !compatible {
		return v2SelectedPackage{}, claim, false
	}
	return v2SelectedPackage{entry: entry, source: source}, Claim{}, true
}

func v2EntryIdentityMatchesSource(entry V2CatalogEntry, source Source) bool {
	return entry.Name == source.Name && entry.RepositoryIdentity == source.Identity.RepositoryIdentity && entry.ImmutableReleaseID == source.Identity.ImmutableReleaseID && entry.AnnotatedTagObject == source.Identity.AnnotatedTagObject && entry.AnnotatedTagTarget == source.Identity.AnnotatedTagTarget && entry.AssetID == source.Identity.AssetID && entry.AssetDigest == source.Identity.AssetDigest && entry.PackageSemanticID == source.Identity.PackageSemanticID && entry.ExportedSymbolSetDigest == source.Identity.ExportedSymbolSetDigest && entry.ContractDigest == source.Identity.ContractDigest && entry.GoToolchainDigest == source.Identity.GoToolchainDigest
}

func identityMatchesImport(entry V2CatalogEntry, imp ImportDecl) bool {
	i := imp.Identity
	return entry.RepositoryIdentity == i.RepositoryIdentity && entry.ImmutableReleaseID == i.ImmutableReleaseID && entry.AnnotatedTagObject == i.AnnotatedTagObject && entry.AnnotatedTagTarget == i.AnnotatedTagTarget && entry.AssetID == i.AssetID && entry.AssetDigest == i.AssetDigest && entry.PackageSemanticID == i.PackageSemanticID && entry.ExportedSymbolSetDigest == i.ExportedSymbolSetDigest && entry.ContractDigest == i.ContractDigest && entry.GoToolchainDigest == i.GoToolchainDigest
}

func v2ImportMatches(source Source, imp ImportDecl) (Claim, bool) {
	if imp.Owner != "" && source.Origin.Owner != imp.Owner {
		return refutedClaim("SEMANTIC", "CHECK_ORIGIN_OWNERSHIP", "ORIGIN_OWNERSHIP_CONFLICT", source.Name), false
	}
	matches := make([]ExportDecl, 0)
	for _, export := range source.Exports {
		if export.SemanticID == imp.SemanticID && export.Type == imp.Type && export.Stage == imp.Stage {
			matches = append(matches, export)
		}
	}
	if len(matches) > 1 {
		return unknownClaim("EXPORT", "SELECT_UNAMBIGUOUS_EXPORT", "AMBIGUOUS_EXPORT", "AMBIGUOUS_EXPORT_SET", "PROVIDE_ONE_EXPORTED_SYMBOL", source.Name), false
	}
	if len(matches) == 0 {
		return refutedClaim("EXPORT", "CHECK_EXPORTED_SYMBOL", "SEMANTIC_EXPORT_INCOMPATIBLE", source.Name), false
	}
	export := matches[0]
	if !containsAll(export.Capabilities, imp.Capabilities) {
		return refutedClaim("EXPORT", "CHECK_CAPABILITY_EXPORT", "CAPABILITY_EXPORT_INCOMPATIBLE", source.Name), false
	}
	if !sameStrings(sortStrings(export.Effects), sortStrings(imp.Effects)) {
		return refutedClaim("EXPORT", "CHECK_EFFECT_EXPORT", "EFFECT_EXPORT_INCOMPATIBLE", source.Name), false
	}
	return Claim{}, true
}

func (s *v2ResolveState) addEdge(from string, selected v2SelectedPackage, imp ImportDecl) {
	edge := V2LinkEdge{From: from, To: selected.entry.Name, Constraint: imp.Constraint, PackageDigest: selected.entry.Digest, PackageSemanticID: selected.entry.PackageSemanticID, ExportedSymbolSetDigest: selected.entry.ExportedSymbolSetDigest}
	for _, existing := range s.edges {
		if v2EdgeKey(existing) == v2EdgeKey(edge) {
			return
		}
	}
	s.edges = append(s.edges, edge)
}

func v2PackageManifest(entry V2CatalogEntry, source Source, contractDigest string) V2PackageManifest {
	return V2PackageManifest{Name: entry.Name, Version: entry.Version, Digest: entry.Digest, SourceDigest: source.SourceDigest, RepositoryIdentity: entry.RepositoryIdentity, ImmutableReleaseID: entry.ImmutableReleaseID, AnnotatedTagObject: entry.AnnotatedTagObject, AnnotatedTagTarget: entry.AnnotatedTagTarget, AssetID: entry.AssetID, AssetDigest: entry.AssetDigest, PackageSemanticID: entry.PackageSemanticID, ExportedSymbolSetDigest: entry.ExportedSymbolSetDigest, ContractDigest: contractDigest, GoToolchainDigest: entry.GoToolchainDigest, Stage: source.Stage, OriginOwner: source.Origin.Owner, VersionContract: source.VersionContract}
}

func mergeV2Exports(selected map[string]v2SelectedPackage) []V2MergedExport {
	merged := make([]V2MergedExport, 0)
	seen := map[string]bool{}
	for _, pkg := range selected {
		for _, export := range pkg.source.Exports {
			key := pkg.entry.Name + "|" + export.ID + "|" + pkg.entry.Digest
			if seen[key] {
				continue
			}
			seen[key] = true
			digest, _ := digestValue(struct {
				Package string     `json:"package"`
				Export  ExportDecl `json:"export"`
			}{pkg.entry.Name, export})
			merged = append(merged, V2MergedExport{Package: pkg.entry.Name, ID: export.ID, SemanticID: export.SemanticID, Type: export.Type, Stage: export.Stage, Capabilities: sortStrings(export.Capabilities), Effects: sortStrings(export.Effects), Optional: export.Optional, SourceDigest: pkg.entry.Digest, ExportDigest: digest})
		}
	}
	sort.Slice(merged, func(i, j int) bool { return v2ExportKey(merged[i]) < v2ExportKey(merged[j]) })
	return merged
}

func exportedSymbolSetDigest(source Source) string {
	canonical := make([]string, 0, len(source.Exports))
	for _, export := range source.Exports {
		canonical = append(canonical, fmt.Sprintf("%s|%s|%s|%s|%s|%s|%t", export.ID, export.SemanticID, export.Type, export.Stage, strings.Join(sortStrings(export.Capabilities), ","), strings.Join(sortStrings(export.Effects), ","), export.Optional))
	}
	sort.Strings(canonical)
	return digestBytes([]byte(strings.Join(canonical, "\n") + "\n"))
}

func toolchainDigest() string { return digestBytes([]byte(V2ToolchainVersion)) }

func v2ImportKey(imp ImportDecl) string {
	return imp.Package + "|" + imp.PackageDigest + "|" + imp.Constraint + "|" + imp.SemanticID + "|" + imp.Type + "|" + imp.Stage + "|" + imp.Identity.ExportedSymbolSetDigest
}

func v2ManifestKey(manifest V2PackageManifest) string { return manifest.Name + "|" + manifest.Digest }
func v2EdgeKey(edge V2LinkEdge) string {
	return edge.From + "|" + edge.To + "|" + edge.Constraint + "|" + edge.PackageDigest + "|" + edge.PackageSemanticID
}
func v2ExportKey(export V2MergedExport) string {
	return export.Package + "|" + export.ID + "|" + export.SourceDigest
}

func v2LinkedIRDigest(resolution V2Resolution) (string, error) {
	ir := struct {
		Schema          string              `json:"schema"`
		Version         string              `json:"version"`
		Root            string              `json:"root"`
		Status          Status              `json:"status"`
		Claim           Claim               `json:"claim"`
		ContractDigest  string              `json:"contract_digest"`
		ToolchainDigest string              `json:"toolchain_digest"`
		MergeStrategy   string              `json:"merge_strategy"`
		Cell            V2CellBinding       `json:"cell"`
		Packages        []V2PackageManifest `json:"packages"`
		Exports         []V2MergedExport    `json:"exports"`
		Edges           []V2LinkEdge        `json:"edges"`
		IdentityDigest  string              `json:"identity_digest"`
	}{V2IRSchema, "v2", resolution.Root, resolution.Status, resolution.Claim, resolution.ContractDigest, resolution.ToolchainDigest, resolution.MergeStrategy, resolution.Cell, resolution.Packages, resolution.Exports, resolution.Edges, resolution.IdentityDigest}
	return digestValue(ir)
}

func VerifyV2Resolution(resolution V2Resolution) error {
	if resolution.Schema != V2ResolutionSchema || resolution.Root == "" || resolution.ContractDigest == "" || resolution.ToolchainDigest == "" || resolution.MergeStrategy != V2MergeStrategy {
		return fmt.Errorf("v2 resolution identity is incomplete")
	}
	if resolution.Cell.CellID == "" || resolution.Cell.CaseID == "" || resolution.Cell.Activity == "" || !contains(V2ProofChoices, resolution.Cell.ProofChoice) || !contains(V2IndicatorClasses, resolution.Cell.IndicatorClass) || resolution.Cell.EvidenceRef == "" {
		return fmt.Errorf("v2 resolution cell evidence binding is incomplete")
	}
	if !resolution.Claim.Valid() || resolution.Status != resolution.Claim.State {
		return fmt.Errorf("v2 resolution claim is invalid or does not match status")
	}
	if resolution.LinkedIRDigest == "" || resolution.IdentityDigest == "" {
		return fmt.Errorf("v2 resolution linked identity is incomplete")
	}
	for _, pkg := range resolution.Packages {
		if pkg.Name == "" || pkg.Version == "" || !strings.HasPrefix(pkg.Digest, "sha256:") || pkg.SourceDigest != pkg.Digest || pkg.RepositoryIdentity == "" || pkg.ImmutableReleaseID == "" || pkg.AnnotatedTagObject == "" || pkg.AnnotatedTagTarget == "" || pkg.AssetID == "" || !strings.HasPrefix(pkg.AssetDigest, "sha256:") || pkg.PackageSemanticID == "" || !strings.HasPrefix(pkg.ExportedSymbolSetDigest, "sha256:") || pkg.ContractDigest != resolution.ContractDigest || pkg.GoToolchainDigest != resolution.ToolchainDigest {
			return fmt.Errorf("v2 selected package %q has an invalid pinned manifest", pkg.Name)
		}
	}
	return nil
}
