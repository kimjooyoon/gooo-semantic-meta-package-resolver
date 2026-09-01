package resolver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type ConformanceResult struct {
	Schema       string            `json:"schema"`
	Status       Status            `json:"status"`
	Cases        int               `json:"cases"`
	Closed       int               `json:"closed"`
	Unknown      int               `json:"unknown"`
	Refuted      int               `json:"refuted"`
	Observations []CaseObservation `json:"observations"`
}

type CaseObservation struct {
	ID            string        `json:"id"`
	Expected      Status        `json:"expected"`
	Observed      Status        `json:"observed"`
	Claim         Claim         `json:"claim"`
	Selected      []PackageLock `json:"selected"`
	ReplayDigest  string        `json:"replay_digest"`
	Deterministic bool          `json:"deterministic"`
}

func RunConformance(repoRoot, fixturesPath, outPath string) (ConformanceResult, error) {
	if !filepath.IsAbs(outPath) {
		return ConformanceResult{}, fmt.Errorf("conformance output must be absolute")
	}
	contract, err := ParseSource(filepath.Join(repoRoot, "contracts", "resolver.gooo"))
	if err != nil {
		return ConformanceResult{}, err
	}
	catalog, catalogLock, err := LoadCatalog(filepath.Join(fixturesPath, "catalog.json"), filepath.Join(fixturesPath, "catalog.lock.json"))
	if err != nil {
		return ConformanceResult{}, err
	}
	var index CaseIndex
	if err := LoadJSON(filepath.Join(fixturesPath, "cases", "index.json"), &index); err != nil {
		return ConformanceResult{}, err
	}
	if index.Schema != CaseSchema || index.FixedDenominator != 7 || len(index.Cases) != 7 {
		return ConformanceResult{}, fmt.Errorf("case denominator must contain exactly seven cases")
	}
	observations := make([]CaseObservation, 0, len(index.Cases))
	for _, spec := range index.Cases {
		root, err := ParseSource(filepath.Join(repoRoot, spec.Source))
		if err != nil {
			return ConformanceResult{}, err
		}
		engine := Resolver{Contract: contract, Catalog: catalog, CatalogLock: catalogLock, CatalogRoot: filepath.Dir(filepath.Join(fixturesPath, "catalog.json"))}
		resolution, err := engine.Resolve(root, nil)
		if err != nil {
			return ConformanceResult{}, fmt.Errorf("case %s: %w", spec.ID, err)
		}
		if resolution.Status != spec.Expected {
			return ConformanceResult{}, fmt.Errorf("case %s: expected %s, got %s", spec.ID, spec.Expected, resolution.Status)
		}
		if !resolution.Claim.Valid() {
			return ConformanceResult{}, fmt.Errorf("case %s: invalid claim tuple", spec.ID)
		}
		artifacts, err := GenerateArtifacts(resolution)
		if err != nil {
			return ConformanceResult{}, err
		}
		replay, err := GenerateArtifacts(resolution)
		if err != nil {
			return ConformanceResult{}, err
		}
		replayDigest := digestBytes(artifacts.IR)
		deterministic := bytes.Equal(artifacts.LinkedGooo, replay.LinkedGooo) && bytes.Equal(artifacts.IR, replay.IR) && bytes.Equal(artifacts.Go, replay.Go) && bytes.Equal(artifacts.Lock, replay.Lock)
		if !deterministic {
			return ConformanceResult{}, fmt.Errorf("case %s: generated replay changed", spec.ID)
		}
		if spec.Assertion == "compatible_extension" && !hasPackageVersion(resolution.Packages, "core", "1.1.0") {
			return ConformanceResult{}, fmt.Errorf("case %s: compatible extension did not select v1.1.0", spec.ID)
		}
		if spec.Assertion == "exact_lock" && !hasPackageVersion(resolution.Packages, "core", "1.0.0") {
			return ConformanceResult{}, fmt.Errorf("case %s: exact lock did not select v1.0.0", spec.ID)
		}
		if spec.Assertion == "effect_refuted" && resolution.Claim.Reason != "EFFECT_SIGNATURE_INCOMPATIBLE" {
			return ConformanceResult{}, fmt.Errorf("case %s: wrong effect refutation", spec.ID)
		}
		if spec.Assertion == "origin_refuted" && resolution.Claim.Reason != "ORIGIN_OWNERSHIP_CONFLICT" {
			return ConformanceResult{}, fmt.Errorf("case %s: wrong origin refutation", spec.ID)
		}
		if spec.Assertion == "missing_proof" && (resolution.Claim.Reason != "IMMUTABLE_PROOF_UNAVAILABLE" || !resolution.Claim.Valid()) {
			return ConformanceResult{}, fmt.Errorf("case %s: missing proof did not preserve UNKNOWN tuple", spec.ID)
		}
		if spec.Assertion == "diamond" {
			reversed := append([]CatalogEntry(nil), catalog.Entries...)
			for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
				reversed[left], reversed[right] = reversed[right], reversed[left]
			}
			reversedResolution, err := (Resolver{Contract: contract, Catalog: Catalog{Schema: catalog.Schema, Entries: reversed}, CatalogLock: catalogLock, CatalogRoot: filepath.Dir(filepath.Join(fixturesPath, "catalog.json"))}).Resolve(root, nil)
			if err != nil {
				return ConformanceResult{}, err
			}
			reversedArtifacts, err := GenerateArtifacts(reversedResolution)
			if err != nil || !bytes.Equal(artifacts.IR, reversedArtifacts.IR) {
				return ConformanceResult{}, fmt.Errorf("case %s: catalog order changed linked IR", spec.ID)
			}
		}
		observations = append(observations, CaseObservation{ID: spec.ID, Expected: spec.Expected, Observed: resolution.Status, Claim: resolution.Claim, Selected: resolution.Packages, ReplayDigest: replayDigest, Deterministic: deterministic})
	}
	sort.Slice(observations, func(i, j int) bool { return observations[i].ID < observations[j].ID })
	result := ConformanceResult{Schema: CaseSchema, Status: Closed, Cases: len(observations), Observations: observations}
	for _, observation := range observations {
		switch observation.Observed {
		case Closed:
			result.Closed++
		case Unknown:
			result.Unknown++
		case Refuted:
			result.Refuted++
		}
	}
	if result.Closed != 4 || result.Unknown != 1 || result.Refuted != 2 {
		return ConformanceResult{}, fmt.Errorf("unexpected fixed case summary")
	}
	if err := os.MkdirAll(outPath, 0o755); err != nil {
		return ConformanceResult{}, err
	}
	data, err := marshalJSON(result)
	if err != nil {
		return ConformanceResult{}, err
	}
	if err := os.WriteFile(filepath.Join(outPath, "conformance.json"), data, 0o644); err != nil {
		return ConformanceResult{}, err
	}
	replay, err := marshalJSON(struct {
		Schema string `json:"schema"`
		Digest string `json:"digest"`
		Cases  int    `json:"cases"`
	}{"gooo/semantic-meta-package-resolver/replay/v1", digestBytes(data), len(observations)})
	if err != nil {
		return ConformanceResult{}, err
	}
	if err := os.WriteFile(filepath.Join(outPath, "replay.json"), replay, 0o644); err != nil {
		return ConformanceResult{}, err
	}
	return result, nil
}

func hasPackageVersion(packages []PackageLock, name, version string) bool {
	for _, pkg := range packages {
		if pkg.Name == name && pkg.Version == version {
			return true
		}
	}
	return false
}

func DecodeCaseIndex(data []byte) (CaseIndex, error) {
	var index CaseIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return CaseIndex{}, err
	}
	return index, nil
}
