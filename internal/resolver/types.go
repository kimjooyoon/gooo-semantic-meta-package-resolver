package resolver

import "sort"

const (
	ContractSchema     = "gooo/semantic-meta-package-resolver/contract/v1"
	CatalogSchema      = "gooo/semantic-meta-package-resolver/catalog/v1"
	CatalogLockSchema  = "gooo/semantic-meta-package-resolver/catalog-lock/v1"
	LockSchema         = "gooo/semantic-meta-package-resolver/lock/v1"
	ResolutionSchema   = "gooo/semantic-meta-package-resolver/resolution/v1"
	IRSchema           = "gooo/semantic-meta-package-resolver/ir/v1"
	ArtifactSchema     = "gooo/semantic-meta-package-resolver/artifact/v1"
	ExecutionSchema    = "gooo/semantic-meta-package-resolver/execution/v1"
	CaseSchema         = "gooo/semantic-meta-package-resolver/cases/v1"
	EvidenceSchema     = "gooo/semantic-meta-package-resolver/ci-evidence/v1"
	V2ContractSchema   = "gooo/semantic-meta-package-resolver/contract/v2"
	V2CatalogSchema    = "gooo/semantic-meta-package-resolver/catalog/v2"
	V2CaseSchema       = "gooo/semantic-meta-package-resolver/cases/v2"
	V2ResolutionSchema = "gooo/semantic-meta-package-resolver/resolution/v2"
	V2IRSchema         = "gooo/semantic-meta-package-resolver/ir/v2"
	V2ArtifactSchema   = "gooo/semantic-meta-package-resolver/artifact/v2"
)

var Precedence = []string{"REFUTED", "UNKNOWN", "CLOSED"}

type Status string

const (
	Closed  Status = "CLOSED"
	Unknown Status = "UNKNOWN"
	Refuted Status = "REFUTED"
)

type Claim struct {
	State         Status   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

func closedClaim(reason string) Claim {
	return Claim{State: Closed, Stage: "RESOLUTION", Step: "RESOLVE_TYPED_META_GRAPH", Reason: reason, NextOperation: "EXECUTE_LINKED_ARTIFACT"}
}

func refutedClaim(stage, step, reason string, blockedBy ...string) Claim {
	return Claim{State: Refuted, Stage: stage, Step: step, Reason: reason, BlockedBy: blockedBy}
}

func unknownClaim(stage, step, reason, class, next string, blockedBy ...string) Claim {
	sort.Strings(blockedBy)
	return Claim{State: Unknown, Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: blockedBy}
}

func (c Claim) Valid() bool {
	if c.State == Unknown {
		return c.Stage != "" && c.Step != "" && c.Reason != "" && c.UnknownClass != "" && c.NextOperation != "" && c.BlockedBy != nil
	}
	return c.State == Closed || c.State == Refuted
}

type VersionContract struct {
	Major         int    `json:"major"`
	Minor         int    `json:"minor"`
	Patch         int    `json:"patch"`
	Compatibility string `json:"compatibility"`
}

type Origin struct {
	Owner string `json:"owner"`
	Kind  string `json:"kind"`
}

type ImmutableProof struct {
	Digest  string `json:"digest"`
	Proof   string `json:"proof"`
	Adapter string `json:"adapter"`
}

type ExportDecl struct {
	ID           string   `json:"id"`
	SemanticID   string   `json:"semantic_id"`
	Type         string   `json:"type"`
	Stage        string   `json:"stage"`
	Capabilities []string `json:"capabilities"`
	Effects      []string `json:"effects"`
	Optional     bool     `json:"optional"`
}

type ImportDecl struct {
	Package       string         `json:"package"`
	Constraint    string         `json:"constraint"`
	PackageDigest string         `json:"package_digest,omitempty"`
	SemanticID    string         `json:"semantic_id"`
	Type          string         `json:"type"`
	Stage         string         `json:"stage"`
	Capabilities  []string       `json:"capabilities"`
	Effects       []string       `json:"effects"`
	Owner         string         `json:"owner"`
	Identity      ImportIdentity `json:"identity,omitempty"`
}

// ImportIdentity is the immutable identity boundary for a v2 semantic import.
type ImportIdentity struct {
	RepositoryIdentity      string `json:"repository_identity"`
	ImmutableReleaseID      string `json:"immutable_release_id"`
	AnnotatedTagObject      string `json:"annotated_tag_object"`
	AnnotatedTagTarget      string `json:"annotated_tag_target"`
	AssetID                 string `json:"asset_id"`
	AssetDigest             string `json:"asset_digest"`
	PackageSemanticID       string `json:"package_semantic_id"`
	ExportedSymbolSetDigest string `json:"exported_symbol_set_digest"`
	ContractDigest          string `json:"contract_digest"`
	GoToolchainDigest       string `json:"go_toolchain_digest"`
}

type LockRef struct {
	Package string `json:"package"`
	Digest  string `json:"digest"`
}

type SelectionPolicy struct {
	Strategy     string `json:"strategy"`
	VersionOrder string `json:"version_order"`
	Registry     string `json:"registry"`
	NoLatest     bool   `json:"no_latest"`
}

type Source struct {
	Schema                    string          `json:"schema"`
	Kind                      string          `json:"kind"`
	Name                      string          `json:"name"`
	LanguageVersion           string          `json:"language_version"`
	Version                   string          `json:"version"`
	SemanticID                string          `json:"semantic_id"`
	Stage                     string          `json:"stage"`
	Origin                    Origin          `json:"origin"`
	VersionContract           VersionContract `json:"version_contract"`
	Exports                   []ExportDecl    `json:"exports"`
	Imports                   []ImportDecl    `json:"imports"`
	Immutable                 ImmutableProof  `json:"immutable"`
	Selection                 SelectionPolicy `json:"selection"`
	Locks                     []LockRef       `json:"locks"`
	Grammar                   []string        `json:"grammar"`
	Rules                     []string        `json:"rules"`
	Types                     []string        `json:"types"`
	Capabilities              []string        `json:"capabilities"`
	Effects                   []string        `json:"effects"`
	OriginRules               []string        `json:"origin_rules"`
	FixedDenominator          int             `json:"fixed_denominator"`
	Precedence                []string        `json:"precedence"`
	UnknownFields             []string        `json:"unknown_fields"`
	Cells                     []string        `json:"cells"`
	MetaActivities            []string        `json:"meta_activities"`
	CellBindings              []V2CellBinding `json:"cell_bindings"`
	FixedCells                int             `json:"fixed_cells"`
	FixedMetaActivities       int             `json:"fixed_meta_activities"`
	FixedClosed               int             `json:"fixed_closed"`
	FixedUnknown              int             `json:"fixed_unknown"`
	FixedRefuted              int             `json:"fixed_refuted"`
	ProofVector               string          `json:"proof_vector"`
	IndicatorVector           string          `json:"indicator_vector"`
	MergeStrategy             string          `json:"merge_strategy"`
	CrossProjectRequiredGates int             `json:"cross_project_required_gates"`
	Identity                  ImportIdentity  `json:"identity"`
	SymbolSetDigest           string          `json:"exported_symbol_set_digest"`
	SourceDigest              string          `json:"source_digest"`
}

type CatalogEntry struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	Digest         string `json:"digest"`
	Source         string `json:"source"`
	ImmutableProof string `json:"immutable_proof"`
	ReleaseAdapter string `json:"release_adapter"`
	Owner          string `json:"owner"`
}

type Catalog struct {
	Schema  string         `json:"schema"`
	Entries []CatalogEntry `json:"entries"`
}

type CatalogLock struct {
	Schema        string `json:"schema"`
	CatalogDigest string `json:"catalog_digest"`
}

type V2CatalogEntry struct {
	Name                    string `json:"name"`
	Version                 string `json:"version"`
	Digest                  string `json:"digest"`
	Source                  string `json:"source"`
	RepositoryIdentity      string `json:"repository_identity"`
	ImmutableReleaseID      string `json:"immutable_release_id"`
	AnnotatedTagObject      string `json:"annotated_tag_object"`
	AnnotatedTagTarget      string `json:"annotated_tag_target"`
	AssetID                 string `json:"asset_id"`
	AssetDigest             string `json:"asset_digest"`
	PackageSemanticID       string `json:"package_semantic_id"`
	ExportedSymbolSetDigest string `json:"exported_symbol_set_digest"`
	ContractDigest          string `json:"contract_digest"`
	GoToolchainDigest       string `json:"go_toolchain_digest"`
}

type V2Catalog struct {
	Schema  string           `json:"schema"`
	Entries []V2CatalogEntry `json:"entries"`
}

type V2CatalogLock struct {
	Schema        string `json:"schema"`
	CatalogDigest string `json:"catalog_digest"`
}

type PackageLock struct {
	Name            string          `json:"name"`
	Version         string          `json:"version"`
	Digest          string          `json:"digest"`
	SourceDigest    string          `json:"source_digest"`
	SemanticID      string          `json:"semantic_id"`
	Stage           string          `json:"stage"`
	TypeRows        []string        `json:"type_rows"`
	OriginOwner     string          `json:"origin_owner"`
	CapabilityRows  []string        `json:"capability_rows"`
	EffectRows      []string        `json:"effect_rows"`
	VersionContract VersionContract `json:"version_contract"`
}

type Lockfile struct {
	Schema         string          `json:"schema"`
	SourceDigest   string          `json:"source_digest"`
	ContractDigest string          `json:"contract_digest"`
	CatalogDigest  string          `json:"catalog_digest"`
	Selection      SelectionPolicy `json:"selection"`
	Packages       []PackageLock   `json:"packages"`
	LockDigest     string          `json:"lock_digest"`
}

type SelectedPackage struct {
	Entry        CatalogEntry
	Source       Source
	SourceDigest string
	Signature    PackageLock
}

type LinkEdge struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Constraint string `json:"constraint"`
	SemanticID string `json:"semantic_id"`
	Type       string `json:"type"`
	Stage      string `json:"stage"`
}

type LinkedIR struct {
	Schema         string        `json:"schema"`
	Version        string        `json:"version"`
	Root           string        `json:"root"`
	Status         Status        `json:"status"`
	Claim          Claim         `json:"claim"`
	ContractDigest string        `json:"contract_digest"`
	CatalogDigest  string        `json:"catalog_digest"`
	LockDigest     string        `json:"lock_digest"`
	Packages       []PackageLock `json:"packages"`
	Edges          []LinkEdge    `json:"edges"`
	IRDigest       string        `json:"ir_digest"`
}

type Resolution struct {
	Schema         string        `json:"schema"`
	Root           string        `json:"root"`
	Status         Status        `json:"status"`
	Claim          Claim         `json:"claim"`
	ContractDigest string        `json:"contract_digest"`
	CatalogDigest  string        `json:"catalog_digest"`
	SourceDigest   string        `json:"source_digest"`
	Packages       []PackageLock `json:"packages"`
	Edges          []LinkEdge    `json:"edges"`
	Decisions      []Claim       `json:"decisions"`
	Lock           Lockfile      `json:"lock"`
}

type Execution struct {
	Schema         string `json:"schema"`
	Status         Status `json:"status"`
	ArtifactDigest string `json:"artifact_digest"`
	OutputDigest   string `json:"output_digest"`
	Output         string `json:"output"`
}

type Artifacts struct {
	LinkedGooo     []byte
	IR             []byte
	Go             []byte
	Lock           []byte
	Resolution     []byte
	Manifest       []byte
	ArtifactDigest string
	IRDigest       string
	LockDigest     string
}

type V2PackageManifest struct {
	Name                    string          `json:"name"`
	Version                 string          `json:"version"`
	Digest                  string          `json:"digest"`
	SourceDigest            string          `json:"source_digest"`
	RepositoryIdentity      string          `json:"repository_identity"`
	ImmutableReleaseID      string          `json:"immutable_release_id"`
	AnnotatedTagObject      string          `json:"annotated_tag_object"`
	AnnotatedTagTarget      string          `json:"annotated_tag_target"`
	AssetID                 string          `json:"asset_id"`
	AssetDigest             string          `json:"asset_digest"`
	PackageSemanticID       string          `json:"package_semantic_id"`
	ExportedSymbolSetDigest string          `json:"exported_symbol_set_digest"`
	ContractDigest          string          `json:"contract_digest"`
	GoToolchainDigest       string          `json:"go_toolchain_digest"`
	Stage                   string          `json:"stage"`
	OriginOwner             string          `json:"origin_owner"`
	VersionContract         VersionContract `json:"version_contract"`
}

type V2MergedExport struct {
	Package      string   `json:"package"`
	ID           string   `json:"id"`
	SemanticID   string   `json:"semantic_id"`
	Type         string   `json:"type"`
	Stage        string   `json:"stage"`
	Capabilities []string `json:"capabilities"`
	Effects      []string `json:"effects"`
	Optional     bool     `json:"optional"`
	SourceDigest string   `json:"source_digest"`
	ExportDigest string   `json:"export_digest"`
}

type V2LinkEdge struct {
	From                    string `json:"from"`
	To                      string `json:"to"`
	Constraint              string `json:"constraint"`
	PackageDigest           string `json:"package_digest"`
	PackageSemanticID       string `json:"package_semantic_id"`
	ExportedSymbolSetDigest string `json:"exported_symbol_set_digest"`
}

type V2Resolution struct {
	Schema          string              `json:"schema"`
	Root            string              `json:"root"`
	Status          Status              `json:"status"`
	Claim           Claim               `json:"claim"`
	ContractDigest  string              `json:"contract_digest"`
	ToolchainDigest string              `json:"toolchain_digest"`
	SourceDigest    string              `json:"source_digest"`
	MergeStrategy   string              `json:"merge_strategy"`
	Cell            V2CellBinding       `json:"cell"`
	Packages        []V2PackageManifest `json:"packages"`
	Exports         []V2MergedExport    `json:"exports"`
	Edges           []V2LinkEdge        `json:"edges"`
	Decisions       []Claim             `json:"decisions"`
	IdentityDigest  string              `json:"identity_digest"`
	LinkedIRDigest  string              `json:"linked_ir_digest"`
}

// V2CellBinding is the authority-owned identity for one fixed conformance
// cell. The same binding is carried into the semantic IR, both dossiers, and
// the CI observation so a state vector cannot be detached from its proof and
// indicator coordinates.
type V2CellBinding struct {
	CellID         string `json:"cell_id"`
	CaseID         string `json:"case_id"`
	Activity       string `json:"activity"`
	ProofChoice    string `json:"proof_choice"`
	IndicatorClass string `json:"indicator_class"`
	EvidenceRef    string `json:"evidence_ref"`
}

type V2Artifacts struct {
	LinkedGooo       []byte
	PackageManifests []byte
	IR               []byte
	Go               []byte
	Resolution       []byte
	MachineDossier   []byte
	HumanDossier     []byte
	Manifest         []byte
	ArtifactDigest   string
	IRDigest         string
}

type V2CaseSpec struct {
	ID             string `json:"id"`
	Source         string `json:"source"`
	Expected       Status `json:"expected"`
	Assertion      string `json:"assertion"`
	CellID         string `json:"cell_id"`
	ProofChoice    string `json:"proof_choice"`
	IndicatorClass string `json:"indicator_class"`
}

type V2CaseIndex struct {
	Schema           string       `json:"schema"`
	FixedDenominator int          `json:"fixed_denominator"`
	Cells            int          `json:"cells"`
	MetaActivities   int          `json:"meta_activities"`
	ProofClosed      int          `json:"proof_closed"`
	ProofUnknown     int          `json:"proof_unknown"`
	ProofRefuted     int          `json:"proof_refuted"`
	IndicatorClosed  int          `json:"indicator_closed"`
	IndicatorUnknown int          `json:"indicator_unknown"`
	IndicatorRefuted int          `json:"indicator_refuted"`
	Cases            []V2CaseSpec `json:"cases"`
}

type V2ConformanceResult struct {
	Schema         string              `json:"schema"`
	Status         Status              `json:"status"`
	Cases          int                 `json:"cases"`
	Closed         int                 `json:"closed"`
	Unknown        int                 `json:"unknown"`
	Refuted        int                 `json:"refuted"`
	Cells          int                 `json:"cells"`
	MetaActivities int                 `json:"meta_activities"`
	Proof          map[string]int      `json:"proof"`
	Indicators     map[string]int      `json:"indicators"`
	Observations   []V2CaseObservation `json:"observations"`
}

type V2CaseObservation struct {
	ID                     string `json:"id"`
	Expected               Status `json:"expected"`
	Observed               Status `json:"observed"`
	Assertion              string `json:"assertion"`
	CellID                 string `json:"cell_id"`
	Activity               string `json:"activity"`
	ProofChoice            string `json:"proof_choice"`
	IndicatorClass         string `json:"indicator_class"`
	EvidenceRef            string `json:"evidence_ref"`
	Claim                  Claim  `json:"claim"`
	IdentityDigest         string `json:"identity_digest"`
	LinkedIRDigest         string `json:"linked_ir_digest"`
	SemanticIRDigest       string `json:"semantic_ir_digest"`
	MachineDossierDigest   string `json:"machine_dossier_digest"`
	HumanDossierDigest     string `json:"human_dossier_digest"`
	ArtifactManifestDigest string `json:"artifact_manifest_digest"`
	Deterministic          bool   `json:"deterministic"`
}

type CaseSpec struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Expected  Status `json:"expected"`
	Assertion string `json:"assertion"`
}

type CaseIndex struct {
	Schema           string     `json:"schema"`
	FixedDenominator int        `json:"fixed_denominator"`
	Cases            []CaseSpec `json:"cases"`
}

func sortStrings(values []string) []string {
	copyOf := append([]string(nil), values...)
	sort.Strings(copyOf)
	return copyOf
}
