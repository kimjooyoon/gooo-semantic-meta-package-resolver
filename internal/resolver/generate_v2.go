package resolver

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type v2LinkedIR struct {
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
	IRDigest        string              `json:"ir_digest"`
}

func makeV2LinkedIR(resolution V2Resolution) (v2LinkedIR, error) {
	ir := v2LinkedIR{Schema: V2IRSchema, Version: "v2", Root: resolution.Root, Status: resolution.Status, Claim: resolution.Claim, ContractDigest: resolution.ContractDigest, ToolchainDigest: resolution.ToolchainDigest, MergeStrategy: resolution.MergeStrategy, Cell: resolution.Cell, Packages: resolution.Packages, Exports: resolution.Exports, Edges: resolution.Edges, IdentityDigest: resolution.IdentityDigest}
	digest, err := digestValue(struct {
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
	}{ir.Schema, ir.Version, ir.Root, ir.Status, ir.Claim, ir.ContractDigest, ir.ToolchainDigest, ir.MergeStrategy, ir.Cell, ir.Packages, ir.Exports, ir.Edges, ir.IdentityDigest})
	if err != nil {
		return v2LinkedIR{}, err
	}
	ir.IRDigest = digest
	return ir, nil
}

func GenerateV2Artifacts(resolution V2Resolution) (V2Artifacts, error) {
	if err := VerifyV2Resolution(resolution); err != nil {
		return V2Artifacts{}, err
	}
	ir, err := makeV2LinkedIR(resolution)
	if err != nil {
		return V2Artifacts{}, err
	}
	if ir.IRDigest != resolution.LinkedIRDigest {
		return V2Artifacts{}, fmt.Errorf("v2 linked IR digest changed after resolution")
	}
	irBytes, err := marshalJSON(ir)
	if err != nil {
		return V2Artifacts{}, err
	}
	linked := []string{
		"gooo linked semantic_meta_package_resolver v2",
		"root name=" + resolution.Root,
		"status state=" + string(resolution.Status),
		"merge strategy=" + resolution.MergeStrategy,
		"contract digest=" + resolution.ContractDigest,
		"toolchain digest=" + resolution.ToolchainDigest,
		"identity digest=" + resolution.IdentityDigest,
		fmt.Sprintf("cell id=%s case_id=%s activity=%s proof_choice=%s indicator_class=%s evidence_ref=%s", resolution.Cell.CellID, resolution.Cell.CaseID, resolution.Cell.Activity, resolution.Cell.ProofChoice, resolution.Cell.IndicatorClass, resolution.Cell.EvidenceRef),
	}
	for _, pkg := range resolution.Packages {
		linked = append(linked, fmt.Sprintf("package-manifest name=%s version=%s digest=%s repository=%s release_id=%s tag_object=%s tag_target=%s asset_id=%s asset_digest=%s package_semantic_id=%s symbols_digest=%s contract_digest=%s toolchain_digest=%s", pkg.Name, pkg.Version, pkg.Digest, pkg.RepositoryIdentity, pkg.ImmutableReleaseID, pkg.AnnotatedTagObject, pkg.AnnotatedTagTarget, pkg.AssetID, pkg.AssetDigest, pkg.PackageSemanticID, pkg.ExportedSymbolSetDigest, pkg.ContractDigest, pkg.GoToolchainDigest))
	}
	for _, export := range resolution.Exports {
		linked = append(linked, fmt.Sprintf("export-merge package=%s id=%s semantic_id=%s type=%s stage=%s source_digest=%s export_digest=%s", export.Package, export.ID, export.SemanticID, export.Type, export.Stage, export.SourceDigest, export.ExportDigest))
	}
	for _, edge := range resolution.Edges {
		linked = append(linked, fmt.Sprintf("import-edge from=%s to=%s constraint=%s package_digest=%s package_semantic_id=%s symbols_digest=%s", edge.From, edge.To, edge.Constraint, edge.PackageDigest, edge.PackageSemanticID, edge.ExportedSymbolSetDigest))
	}
	linkedGooo := []byte(strings.Join(linked, "\n") + "\n")
	manifests, err := marshalJSON(struct {
		Schema   string              `json:"schema"`
		Packages []V2PackageManifest `json:"packages"`
	}{V2CatalogSchema + "-pinned-manifests", resolution.Packages})
	if err != nil {
		return V2Artifacts{}, err
	}
	resolutionBytes, err := marshalJSON(resolution)
	if err != nil {
		return V2Artifacts{}, err
	}
	goBytes := []byte(generateV2GoArtifact(resolution))
	machine, err := marshalJSON(v2MachineDossier(resolution))
	if err != nil {
		return V2Artifacts{}, err
	}
	human := []byte(generateV2HumanDossier(resolution))
	files := []struct {
		Path string
		Data []byte
	}{
		{"linked.gooo", linkedGooo},
		{"pinned-package-manifests.json", manifests},
		{"linked-ir.json", irBytes},
		{"linked-artifact.go", goBytes},
		{"resolution.json", resolutionBytes},
		{"machine-dossier.json", machine},
		{"human-dossier.md", human},
	}
	manifest := struct {
		Schema         string `json:"schema"`
		Status         Status `json:"status"`
		IdentityDigest string `json:"identity_digest"`
		IRDigest       string `json:"ir_digest"`
		ArtifactDigest string `json:"artifact_digest"`
		Files          []struct {
			Path   string `json:"path"`
			Digest string `json:"digest"`
			Bytes  int    `json:"bytes"`
		} `json:"files"`
	}{Schema: V2ArtifactSchema, Status: resolution.Status, IdentityDigest: resolution.IdentityDigest, IRDigest: resolution.LinkedIRDigest}
	for _, file := range files {
		manifest.Files = append(manifest.Files, struct {
			Path   string `json:"path"`
			Digest string `json:"digest"`
			Bytes  int    `json:"bytes"`
		}{file.Path, digestBytes(file.Data), len(file.Data)})
	}
	manifest.ArtifactDigest = digestBytes(goBytes)
	manifestBytes, err := marshalJSON(manifest)
	if err != nil {
		return V2Artifacts{}, err
	}
	return V2Artifacts{LinkedGooo: linkedGooo, PackageManifests: manifests, IR: irBytes, Go: goBytes, Resolution: resolutionBytes, MachineDossier: machine, HumanDossier: human, Manifest: manifestBytes, ArtifactDigest: manifest.ArtifactDigest, IRDigest: resolution.LinkedIRDigest}, nil
}

func v2MachineDossier(resolution V2Resolution) any {
	counts := map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0}
	counts[string(resolution.Status)] = 1
	return struct {
		Schema          string              `json:"schema"`
		Status          Status              `json:"status"`
		Claim           Claim               `json:"claim"`
		IdentityDigest  string              `json:"identity_digest"`
		ContractDigest  string              `json:"contract_digest"`
		ToolchainDigest string              `json:"toolchain_digest"`
		MergeStrategy   string              `json:"merge_strategy"`
		Cell            V2CellBinding       `json:"cell"`
		StateCounts     map[string]int      `json:"state_counts"`
		Packages        []V2PackageManifest `json:"packages"`
		Exports         []V2MergedExport    `json:"exports"`
		Edges           []V2LinkEdge        `json:"edges"`
		}{V2ArtifactSchema + "/machine-dossier", resolution.Status, resolution.Claim, resolution.IdentityDigest, resolution.ContractDigest, resolution.ToolchainDigest, resolution.MergeStrategy, resolution.Cell, counts, resolution.Packages, resolution.Exports, resolution.Edges}
}

func generateV2HumanDossier(resolution V2Resolution) string {
	var b strings.Builder
	b.WriteString("# Gooo semantic import dossier v2\n\n")
	fmt.Fprintf(&b, "status: %s\nroot: %s\nmerge_strategy: %s\ncontract_digest: %s\ntoolchain_digest: %s\nidentity_digest: %s\n\n", resolution.Status, resolution.Root, resolution.MergeStrategy, resolution.ContractDigest, resolution.ToolchainDigest, resolution.IdentityDigest)
	b.WriteString("## Authority cell\n\n")
	b.WriteString("| cell | case | activity | proof_choice | indicator_class | evidence_ref |\n|---|---|---|---|---|---|\n")
	fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n", resolution.Cell.CellID, resolution.Cell.CaseID, resolution.Cell.Activity, resolution.Cell.ProofChoice, resolution.Cell.IndicatorClass, resolution.Cell.EvidenceRef)
	b.WriteString("## Pinned package manifests\n\n")
	b.WriteString("| package | version | package digest | release ID | asset digest | semantic ID | symbol set digest |\n|---|---|---|---|---|---|---|\n")
	for _, pkg := range resolution.Packages {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s |\n", pkg.Name, pkg.Version, pkg.Digest, pkg.ImmutableReleaseID, pkg.AssetDigest, pkg.PackageSemanticID, pkg.ExportedSymbolSetDigest)
	}
	b.WriteString("\n## Export merge\n\n")
	b.WriteString("| package | symbol | semantic ID | type | stage | export digest |\n|---|---|---|---|---|---|\n")
	for _, export := range resolution.Exports {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n", export.Package, export.ID, export.SemanticID, export.Type, export.Stage, export.ExportDigest)
	}
	b.WriteString("\n## Claim\n\n")
	fmt.Fprintf(&b, "- state: %s\n- stage: %s\n- step: %s\n- reason: %s\n", resolution.Claim.State, resolution.Claim.Stage, resolution.Claim.Step, resolution.Claim.Reason)
	if resolution.Status == Unknown {
		fmt.Fprintf(&b, "- unknown_class: %s\n- next_operation: %s\n- blocked_by: %s\n", resolution.Claim.UnknownClass, resolution.Claim.NextOperation, strings.Join(resolution.Claim.BlockedBy, ","))
	}
	return b.String()
}

func generateV2GoArtifact(resolution V2Resolution) string {
	payload := struct {
		Schema         string `json:"schema"`
		Status         Status `json:"status"`
		Root           string `json:"root"`
		IdentityDigest string `json:"identity_digest"`
		LinkedIRDigest string `json:"linked_ir_digest"`
	}{V2ArtifactSchema, resolution.Status, resolution.Root, resolution.IdentityDigest, resolution.LinkedIRDigest}
	data, _ := json.Marshal(payload)
	return fmt.Sprintf("// Code generated by gooo semantic import resolver; DO NOT EDIT.\npackage main\n\nimport \"fmt\"\n\nconst LinkedArtifactSchema = %q\nconst LinkedArtifactDigestBasis = %q\n\nfunc main() {\n\tfmt.Println(%q)\n}\n", V2ArtifactSchema, resolution.LinkedIRDigest, string(data))
}

func WriteV2Artifacts(out string, artifacts V2Artifacts) error {
	if !filepath.IsAbs(out) {
		return fmt.Errorf("output directory must be an absolute caller-owned path")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	files := map[string][]byte{
		"linked.gooo":                   artifacts.LinkedGooo,
		"pinned-package-manifests.json": artifacts.PackageManifests,
		"linked-ir.json":                artifacts.IR,
		"linked-artifact.go":            artifacts.Go,
		"resolution.json":               artifacts.Resolution,
		"machine-dossier.json":          artifacts.MachineDossier,
		"human-dossier.md":              artifacts.HumanDossier,
		"manifest.json":                 artifacts.Manifest,
	}
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := os.WriteFile(filepath.Join(out, path), files[path], 0o644); err != nil {
			return err
		}
	}
	return nil
}

func ExecuteV2Artifact(path string) (Execution, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Execution{}, err
	}
	artifactDigest := digestBytes(data)
	if !strings.Contains(string(data), "const LinkedArtifactSchema") || !strings.Contains(string(data), V2ArtifactSchema) {
		return Execution{}, fmt.Errorf("generated v2 artifact schema marker missing")
	}
	cmd := exec.Command("go", "run", path)
	cmd.Dir = filepath.Dir(path)
	cmd.Env = append(os.Environ(), "GO111MODULE=off")
	output, err := cmd.Output()
	if err != nil {
		return Execution{}, fmt.Errorf("execute generated v2 artifact: %w", err)
	}
	output = []byte(strings.TrimSpace(string(output)))
	var payload map[string]any
	if err := json.Unmarshal(output, &payload); err != nil {
		return Execution{}, fmt.Errorf("generated v2 artifact output is not JSON: %w", err)
	}
	if payload["schema"] != V2ArtifactSchema {
		return Execution{}, fmt.Errorf("generated v2 artifact returned an unexpected schema")
	}
	return Execution{Schema: V2ArtifactSchema + "/execution", Status: Status(payload["status"].(string)), ArtifactDigest: artifactDigest, OutputDigest: digestBytes(output), Output: string(output)}, nil
}
