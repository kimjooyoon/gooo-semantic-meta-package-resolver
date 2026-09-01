package resolver

import "sort"

const (
	ContractSchema    = "gooo/semantic-meta-package-resolver/contract/v1"
	CatalogSchema     = "gooo/semantic-meta-package-resolver/catalog/v1"
	CatalogLockSchema = "gooo/semantic-meta-package-resolver/catalog-lock/v1"
	LockSchema        = "gooo/semantic-meta-package-resolver/lock/v1"
	ResolutionSchema  = "gooo/semantic-meta-package-resolver/resolution/v1"
	IRSchema          = "gooo/semantic-meta-package-resolver/ir/v1"
	ArtifactSchema    = "gooo/semantic-meta-package-resolver/artifact/v1"
	ExecutionSchema   = "gooo/semantic-meta-package-resolver/execution/v1"
	CaseSchema        = "gooo/semantic-meta-package-resolver/cases/v1"
	EvidenceSchema    = "gooo/semantic-meta-package-resolver/ci-evidence/v1"
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
	Package      string   `json:"package"`
	Constraint   string   `json:"constraint"`
	SemanticID   string   `json:"semantic_id"`
	Type         string   `json:"type"`
	Stage        string   `json:"stage"`
	Capabilities []string `json:"capabilities"`
	Effects      []string `json:"effects"`
	Owner        string   `json:"owner"`
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
	Schema           string          `json:"schema"`
	Kind             string          `json:"kind"`
	Name             string          `json:"name"`
	LanguageVersion  string          `json:"language_version"`
	Version          string          `json:"version"`
	SemanticID       string          `json:"semantic_id"`
	Stage            string          `json:"stage"`
	Origin           Origin          `json:"origin"`
	VersionContract  VersionContract `json:"version_contract"`
	Exports          []ExportDecl    `json:"exports"`
	Imports          []ImportDecl    `json:"imports"`
	Immutable        ImmutableProof  `json:"immutable"`
	Selection        SelectionPolicy `json:"selection"`
	Locks            []LockRef       `json:"locks"`
	Grammar          []string        `json:"grammar"`
	Rules            []string        `json:"rules"`
	Types            []string        `json:"types"`
	Capabilities     []string        `json:"capabilities"`
	Effects          []string        `json:"effects"`
	OriginRules      []string        `json:"origin_rules"`
	FixedDenominator int             `json:"fixed_denominator"`
	Precedence       []string        `json:"precedence"`
	UnknownFields    []string        `json:"unknown_fields"`
	SourceDigest     string          `json:"source_digest"`
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
