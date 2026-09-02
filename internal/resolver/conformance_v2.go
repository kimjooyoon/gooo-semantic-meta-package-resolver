package resolver

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func RunV2Conformance(repoRoot, fixturesPath, outPath string) (V2ConformanceResult, error) {
	if !filepath.IsAbs(outPath) {
		return V2ConformanceResult{}, fmt.Errorf("v2 conformance output must be absolute")
	}
	contract, err := ParseSource(filepath.Join(repoRoot, "contracts", "resolver-v2.gooo"))
	if err != nil {
		return V2ConformanceResult{}, err
	}
	catalog, catalogLock, err := LoadV2Catalog(filepath.Join(fixturesPath, "v2", "catalog.json"), filepath.Join(fixturesPath, "v2", "catalog.lock.json"))
	if err != nil {
		return V2ConformanceResult{}, err
	}
	var index V2CaseIndex
	if err := LoadJSON(filepath.Join(fixturesPath, "v2", "cases", "index.json"), &index); err != nil {
		return V2ConformanceResult{}, err
	}
	if index.Schema != V2CaseSchema || index.FixedDenominator != 12 || index.Cells != 12 || index.MetaActivities != 12 || index.ProofClosed != 4 || index.ProofUnknown != 4 || index.ProofRefuted != 4 || index.IndicatorClosed != 4 || index.IndicatorUnknown != 4 || index.IndicatorRefuted != 4 || len(index.Cases) != 12 {
		return V2ConformanceResult{}, fmt.Errorf("v2 case denominator must contain exactly 12 cells and 4/4/4 proof and indicator vectors")
	}
	if err := validateV2CaseIndex(contract, index); err != nil {
		return V2ConformanceResult{}, err
	}
	observations := make([]V2CaseObservation, 0, len(index.Cases))
	for _, spec := range index.Cases {
		root, err := ParseSource(filepath.Join(repoRoot, spec.Source))
		if err != nil {
			return V2ConformanceResult{}, err
		}
		engine := V2Resolver{Contract: contract, Catalog: catalog, CatalogLock: catalogLock, CatalogRoot: filepath.Dir(filepath.Join(fixturesPath, "v2", "catalog.json"))}
		resolution, err := engine.Resolve(root)
		if err != nil {
			return V2ConformanceResult{}, fmt.Errorf("case %s: %w", spec.ID, err)
		}
		if resolution.Status != spec.Expected {
			return V2ConformanceResult{}, fmt.Errorf("case %s: expected %s, got %s", spec.ID, spec.Expected, resolution.Status)
		}
		if resolution.Cell.CellID != spec.CellID || resolution.Cell.CaseID != spec.ID || resolution.Cell.ProofChoice != spec.ProofChoice || resolution.Cell.IndicatorClass != spec.IndicatorClass {
			return V2ConformanceResult{}, fmt.Errorf("case %s: resolution cell evidence does not match authority mapping", spec.ID)
		}
		if !resolution.Claim.Valid() {
			return V2ConformanceResult{}, fmt.Errorf("case %s: invalid claim tuple", spec.ID)
		}
		if err := assertV2Case(spec, resolution); err != nil {
			return V2ConformanceResult{}, err
		}
		artifacts, err := GenerateV2Artifacts(resolution)
		if err != nil {
			return V2ConformanceResult{}, err
		}
		replay, err := GenerateV2Artifacts(resolution)
		if err != nil {
			return V2ConformanceResult{}, err
		}
		deterministic := bytes.Equal(artifacts.LinkedGooo, replay.LinkedGooo) && bytes.Equal(artifacts.PackageManifests, replay.PackageManifests) && bytes.Equal(artifacts.IR, replay.IR) && bytes.Equal(artifacts.Go, replay.Go) && bytes.Equal(artifacts.Resolution, replay.Resolution) && bytes.Equal(artifacts.MachineDossier, replay.MachineDossier) && bytes.Equal(artifacts.HumanDossier, replay.HumanDossier) && bytes.Equal(artifacts.Manifest, replay.Manifest)
		if !deterministic {
			return V2ConformanceResult{}, fmt.Errorf("case %s: generated v2 replay changed", spec.ID)
		}
		caseDir := filepath.Join(outPath, "cases", spec.ID)
		if filepath.Base(caseDir) != spec.ID {
			return V2ConformanceResult{}, fmt.Errorf("case %s: unsafe evidence output path", spec.ID)
		}
		if err := WriteV2Artifacts(caseDir, artifacts); err != nil {
			return V2ConformanceResult{}, fmt.Errorf("case %s: write evidence artifacts: %w", spec.ID, err)
		}
		if spec.Assertion == "diamond" {
			reversed := append([]V2CatalogEntry(nil), catalog.Entries...)
			for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
				reversed[left], reversed[right] = reversed[right], reversed[left]
			}
			reversedResolution, err := (V2Resolver{Contract: contract, Catalog: V2Catalog{Schema: catalog.Schema, Entries: reversed}, CatalogLock: catalogLock, CatalogRoot: filepath.Dir(filepath.Join(fixturesPath, "v2", "catalog.json"))}).Resolve(root)
			if err != nil {
				return V2ConformanceResult{}, err
			}
			reversedArtifacts, err := GenerateV2Artifacts(reversedResolution)
			if err != nil || !bytes.Equal(artifacts.IR, reversedArtifacts.IR) {
				return V2ConformanceResult{}, fmt.Errorf("case %s: catalog order changed canonical linked IR", spec.ID)
			}
		}
		observations = append(observations, V2CaseObservation{ID: spec.ID, Expected: spec.Expected, Observed: resolution.Status, Assertion: spec.Assertion, CellID: resolution.Cell.CellID, Activity: resolution.Cell.Activity, ProofChoice: resolution.Cell.ProofChoice, IndicatorClass: resolution.Cell.IndicatorClass, EvidenceRef: resolution.Cell.EvidenceRef, Claim: resolution.Claim, IdentityDigest: resolution.IdentityDigest, LinkedIRDigest: resolution.LinkedIRDigest, SemanticIRDigest: digestBytes(artifacts.IR), MachineDossierDigest: digestBytes(artifacts.MachineDossier), HumanDossierDigest: digestBytes(artifacts.HumanDossier), ArtifactManifestDigest: digestBytes(artifacts.Manifest), Deterministic: deterministic})
	}
	sort.Slice(observations, func(i, j int) bool { return observations[i].ID < observations[j].ID })
	proofCounts := map[string]int{}
	indicatorCounts := map[string]int{}
	result := V2ConformanceResult{Schema: V2CaseSchema, Status: Closed, Cases: len(observations), Cells: index.Cells, MetaActivities: index.MetaActivities, Proof: map[string]int{"CLOSED": 4, "UNKNOWN": 4, "REFUTED": 4}, Indicators: map[string]int{"CLOSED": 4, "UNKNOWN": 4, "REFUTED": 4}, Observations: observations}
	for _, observation := range observations {
		proofCounts[observation.ProofChoice]++
		indicatorCounts[observation.IndicatorClass]++
		switch observation.Observed {
		case Closed:
			result.Closed++
		case Unknown:
			result.Unknown++
		case Refuted:
			result.Refuted++
		}
	}
	if result.Closed != 4 || result.Unknown != 4 || result.Refuted != 4 || proofCounts["FOUNDATION"] != 4 || proofCounts["COHERENCE"] != 4 || proofCounts["REGRESSION"] != 4 || indicatorCounts["DRIVER"] != 4 || indicatorCounts["OUTCOME"] != 4 || indicatorCounts["GUARDRAIL"] != 4 {
		return V2ConformanceResult{}, fmt.Errorf("unexpected v2 fixed case summary")
	}
	if err := os.MkdirAll(outPath, 0o755); err != nil {
		return V2ConformanceResult{}, err
	}
	data, err := marshalJSON(result)
	if err != nil {
		return V2ConformanceResult{}, err
	}
	if err := os.WriteFile(filepath.Join(outPath, "conformance-v2.json"), data, 0o644); err != nil {
		return V2ConformanceResult{}, err
	}
	return result, nil
}

func validateV2CaseIndex(contract Source, index V2CaseIndex) error {
	seen := make(map[string]bool, len(index.Cases))
	for _, spec := range index.Cases {
		if spec.ID == "" || spec.CellID == "" || spec.ProofChoice == "" || spec.IndicatorClass == "" {
			return fmt.Errorf("case %s: proof and indicator mapping is incomplete", spec.ID)
		}
		if seen[spec.ID] {
			return fmt.Errorf("case %s: duplicate case mapping", spec.ID)
		}
		seen[spec.ID] = true
		var binding *V2CellBinding
		for index := range contract.CellBindings {
			if contract.CellBindings[index].CaseID == spec.ID {
				binding = &contract.CellBindings[index]
				break
			}
		}
		if binding == nil || binding.CellID != spec.CellID || binding.ProofChoice != spec.ProofChoice || binding.IndicatorClass != spec.IndicatorClass {
			return fmt.Errorf("case %s: index mapping does not match .gooo authority", spec.ID)
		}
	}
	if len(seen) != 12 {
		return fmt.Errorf("v2 case index must map all 12 authority cells")
	}
	for _, binding := range contract.CellBindings {
		found := false
		for _, spec := range index.Cases {
			if spec.CellID == binding.CellID && spec.ID == binding.CaseID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("authority cell %s has no case-index evidence binding", binding.CellID)
		}
	}
	return nil
}

func assertV2Case(spec V2CaseSpec, resolution V2Resolution) error {
	if spec.Expected == Unknown && !resolution.Claim.Valid() {
		return fmt.Errorf("case %s: UNKNOWN claim is incomplete", spec.ID)
	}
	switch spec.Assertion {
	case "exact_single":
		if len(resolution.Packages) != 1 || resolution.Packages[0].Name != "core" {
			return fmt.Errorf("case %s: exact single import did not pin core once", spec.ID)
		}
	case "diamond":
		if len(resolution.Packages) != 3 || len(resolution.Exports) != 3 {
			return fmt.Errorf("case %s: diamond did not converge to three manifests and exports", spec.ID)
		}
	case "dedupe":
		if len(resolution.Packages) != 1 || len(resolution.Edges) != 1 {
			return fmt.Errorf("case %s: same-name same-digest was not deduplicated", spec.ID)
		}
	case "fixed_point":
		if resolution.MergeStrategy != V2MergeStrategy || len(resolution.Exports) != 1 {
			return fmt.Errorf("case %s: explicit FIXED_POINT merge was not closed", spec.ID)
		}
	case "missing_release":
		if resolution.Claim.Reason != "IMMUTABLE_RELEASE_UNAVAILABLE" {
			return fmt.Errorf("case %s: wrong missing-release reason", spec.ID)
		}
	case "missing_asset":
		if resolution.Claim.Reason != "ASSET_DIGEST_UNAVAILABLE" {
			return fmt.Errorf("case %s: wrong missing-asset reason", spec.ID)
		}
	case "ambiguous_export":
		if resolution.Claim.Reason != "AMBIGUOUS_EXPORT" {
			return fmt.Errorf("case %s: wrong ambiguous-export reason", spec.ID)
		}
	case "toolchain_mismatch":
		if resolution.Claim.Reason != "TOOLCHAIN_MISMATCH" {
			return fmt.Errorf("case %s: wrong toolchain-mismatch reason", spec.ID)
		}
	case "conflicting_digest":
		if resolution.Claim.Reason != "PACKAGE_DIGEST_CONFLICT" {
			return fmt.Errorf("case %s: wrong conflicting-digest reason", spec.ID)
		}
	case "dependency_cycle":
		if resolution.Claim.Reason != "DEPENDENCY_CYCLE" {
			return fmt.Errorf("case %s: wrong dependency-cycle reason", spec.ID)
		}
	case "forged_identity":
		if resolution.Claim.Reason != "FORGED_PACKAGE_IDENTITY" {
			return fmt.Errorf("case %s: wrong forged-identity reason", spec.ID)
		}
	case "semantic_export":
		if resolution.Claim.Reason != "SEMANTIC_EXPORT_INCOMPATIBLE" {
			return fmt.Errorf("case %s: wrong semantic-export reason", spec.ID)
		}
	}
	return nil
}
