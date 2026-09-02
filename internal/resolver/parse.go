package resolver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ParseSource(path string) (Source, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Source{}, err
	}
	source := Source{SourceDigest: digestBytes(data), LanguageVersion: "v1"}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		values, valueErr := parseKeyValues(fields[1:])
		switch fields[0] {
		case "gooo":
			if len(fields) != 4 || (fields[3] != "v1" && fields[3] != "v2") {
				return Source{}, fmt.Errorf("line %d: invalid gooo header", lineNo)
			}
			source.Kind, source.Name, source.LanguageVersion = fields[1], fields[2], fields[3]
			source.Schema = "gooo/" + fields[1] + "/source/" + fields[3]
		case "grammar":
			if valueErr != nil || len(values) != 1 {
				return Source{}, fmt.Errorf("line %d: invalid grammar: %w", lineNo, valueErr)
			}
			for _, value := range values {
				if value == "" {
					return Source{}, fmt.Errorf("line %d: invalid grammar declaration", lineNo)
				}
				source.Grammar = append(source.Grammar, value)
			}
		case "rule":
			if valueErr != nil || values["name"] == "" || values["definition"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid rule: %w", lineNo, valueErr)
			}
			source.Rules = append(source.Rules, values["name"]+"="+values["definition"])
		case "type":
			if valueErr != nil || values["name"] == "" || values["definition"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid type: %w", lineNo, valueErr)
			}
			source.Types = append(source.Types, values["name"]+"="+values["definition"])
		case "capability":
			if valueErr != nil || values["name"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid capability: %w", lineNo, valueErr)
			}
			source.Capabilities = append(source.Capabilities, values["name"]+"="+values["row"])
		case "effect":
			if valueErr != nil || values["name"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid effect: %w", lineNo, valueErr)
			}
			source.Effects = append(source.Effects, values["name"]+"="+values["row"])
		case "origin_rule":
			if valueErr != nil || values["name"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid origin rule: %w", lineNo, valueErr)
			}
			source.OriginRules = append(source.OriginRules, values["name"]+"="+values["definition"])
		case "fixed_denominator":
			if valueErr != nil {
				return Source{}, fmt.Errorf("line %d: invalid fixed denominator: %w", lineNo, valueErr)
			}
			source.FixedDenominator, err = parseInt(values, "cases")
			if err != nil {
				return Source{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			if values["cells"] != "" {
				source.FixedCells, err = strconv.Atoi(values["cells"])
				if err != nil {
					return Source{}, fmt.Errorf("line %d: invalid cells", lineNo)
				}
			}
			if values["meta_activities"] != "" {
				source.FixedMetaActivities, err = strconv.Atoi(values["meta_activities"])
				if err != nil {
					return Source{}, fmt.Errorf("line %d: invalid meta activities", lineNo)
				}
			}
			source.ProofVector = values["proof"]
			source.IndicatorVector = values["indicator"]
			if values["closed"] != "" {
				source.FixedClosed, err = strconv.Atoi(values["closed"])
				if err != nil {
					return Source{}, fmt.Errorf("line %d: invalid closed count", lineNo)
				}
			}
			if values["unknown"] != "" {
				source.FixedUnknown, err = strconv.Atoi(values["unknown"])
				if err != nil {
					return Source{}, fmt.Errorf("line %d: invalid unknown count", lineNo)
				}
			}
			if values["refuted"] != "" {
				source.FixedRefuted, err = strconv.Atoi(values["refuted"])
				if err != nil {
					return Source{}, fmt.Errorf("line %d: invalid refuted count", lineNo)
				}
			}
		case "precedence":
			if valueErr != nil || values["states"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid precedence: %w", lineNo, valueErr)
			}
			source.Precedence = strings.Split(values["states"], ">")
		case "unknown_fields":
			if valueErr != nil || values["fields"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid unknown fields: %w", lineNo, valueErr)
			}
			source.UnknownFields = strings.Split(values["fields"], ",")
		case "cell":
			if valueErr != nil || values["name"] == "" || values["activity"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid cell: %w", lineNo, valueErr)
			}
			source.Cells = append(source.Cells, values["name"])
			source.MetaActivities = append(source.MetaActivities, values["activity"])
		case "merge":
			if valueErr != nil || values["strategy"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid merge strategy: %w", lineNo, valueErr)
			}
			source.MergeStrategy = values["strategy"]
		case "gate":
			if valueErr != nil || values["name"] == "" || values["value"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid gate declaration: %w", lineNo, valueErr)
			}
			if values["name"] == "cross_project_required_gates" {
				source.CrossProjectRequiredGates, err = strconv.Atoi(values["value"])
				if err != nil {
					return Source{}, fmt.Errorf("line %d: invalid cross-project gate count", lineNo)
				}
			}
		case "repository":
			if valueErr != nil || values["identity"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid repository identity: %w", lineNo, valueErr)
			}
			source.Identity.RepositoryIdentity = values["identity"]
		case "release":
			if valueErr != nil || values["id"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid immutable release identity: %w", lineNo, valueErr)
			}
			source.Identity.ImmutableReleaseID = values["id"]
		case "tag":
			if valueErr != nil || values["object"] == "" || values["target"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid annotated tag identity: %w", lineNo, valueErr)
			}
			source.Identity.AnnotatedTagObject = values["object"]
			source.Identity.AnnotatedTagTarget = values["target"]
		case "asset":
			if valueErr != nil || values["id"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid release asset identity: %w", lineNo, valueErr)
			}
			source.Identity.AssetID = values["id"]
			source.Identity.AssetDigest = values["digest"]
		case "symbol_set":
			if valueErr != nil || values["digest"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid exported symbol set digest: %w", lineNo, valueErr)
			}
			source.SymbolSetDigest = values["digest"]
			source.Identity.ExportedSymbolSetDigest = values["digest"]
		case "contract":
			if valueErr != nil || values["digest"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid contract digest: %w", lineNo, valueErr)
			}
			source.Identity.ContractDigest = values["digest"]
		case "toolchain":
			if valueErr != nil || values["digest"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid Go toolchain digest: %w", lineNo, valueErr)
			}
			source.Identity.GoToolchainDigest = values["digest"]
		case "package", "version", "semantic_id", "stage":
			if valueErr != nil {
				return Source{}, fmt.Errorf("line %d: invalid %s: %w", lineNo, fields[0], valueErr)
			}
			switch fields[0] {
			case "package":
				source.Name = values["name"]
			case "version":
				source.Version = values["value"]
			case "semantic_id":
				source.SemanticID = values["value"]
				source.Identity.PackageSemanticID = values["value"]
			case "stage":
				source.Stage = values["value"]
			}
		case "origin":
			if valueErr != nil {
				return Source{}, fmt.Errorf("line %d: invalid origin: %w", lineNo, valueErr)
			}
			source.Origin = Origin{Owner: values["owner"], Kind: values["kind"]}
		case "version_contract":
			if valueErr != nil {
				return Source{}, fmt.Errorf("line %d: invalid version contract: %w", lineNo, valueErr)
			}
			source.VersionContract = VersionContract{Compatibility: values["compatibility"]}
			source.VersionContract.Major, err = parseInt(values, "major")
			if err != nil {
				return Source{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			source.VersionContract.Minor, err = parseInt(values, "minor")
			if err != nil {
				return Source{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			source.VersionContract.Patch, err = parseInt(values, "patch")
			if err != nil {
				return Source{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
		case "export":
			if valueErr != nil {
				return Source{}, fmt.Errorf("line %d: invalid export: %w", lineNo, valueErr)
			}
			export, err := parseExport(values)
			if err != nil {
				return Source{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			source.Exports = append(source.Exports, export)
		case "import":
			if valueErr != nil {
				return Source{}, fmt.Errorf("line %d: invalid import: %w", lineNo, valueErr)
			}
			imp, err := parseImport(values)
			if err != nil {
				return Source{}, fmt.Errorf("line %d: %w", lineNo, err)
			}
			source.Imports = append(source.Imports, imp)
		case "immutable":
			if valueErr != nil {
				return Source{}, fmt.Errorf("line %d: invalid immutable proof: %w", lineNo, valueErr)
			}
			source.Immutable = ImmutableProof{Digest: values["digest"], Proof: values["proof"], Adapter: values["adapter"]}
		case "selection":
			if valueErr != nil {
				return Source{}, fmt.Errorf("line %d: invalid selection policy: %w", lineNo, valueErr)
			}
			source.Selection = SelectionPolicy{Strategy: values["strategy"], VersionOrder: values["version_order"], Registry: values["registry"], NoLatest: values["no_latest"] == "true"}
		case "lock":
			if valueErr != nil || values["package"] == "" || values["digest"] == "" {
				return Source{}, fmt.Errorf("line %d: invalid lock: %w", lineNo, valueErr)
			}
			source.Locks = append(source.Locks, LockRef{Package: values["package"], Digest: values["digest"]})
		default:
			return Source{}, fmt.Errorf("line %d: unknown declaration %q", lineNo, fields[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return Source{}, err
	}
	if source.Kind == "" || source.Name == "" {
		return Source{}, fmt.Errorf("source is missing a gooo header")
	}
	return source, nil
}

func parseKeyValues(fields []string) (map[string]string, error) {
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("invalid key/value %q", field)
		}
		values[parts[0]] = strings.Trim(parts[1], "\"")
	}
	return values, nil
}

func parseInt(values map[string]string, key string) (int, error) {
	value, ok := values[key]
	if !ok || value == "" {
		return 0, fmt.Errorf("missing %s", key)
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q", key, value)
	}
	return number, nil
}

func parseBool(values map[string]string, key string) bool {
	return values[key] == "true"
}

func parseRows(value string) []string {
	if value == "" || value == "-" {
		return nil
	}
	return sortStrings(strings.Split(value, ","))
}

func parseExport(values map[string]string) (ExportDecl, error) {
	if values["id"] == "" || values["semantic_id"] == "" || values["type"] == "" || values["stage"] == "" {
		return ExportDecl{}, fmt.Errorf("export requires id, semantic_id, type, stage")
	}
	return ExportDecl{ID: values["id"], SemanticID: values["semantic_id"], Type: values["type"], Stage: values["stage"], Capabilities: parseRows(values["capability"]), Effects: parseRows(values["effect"]), Optional: parseBool(values, "optional")}, nil
}

func parseImport(values map[string]string) (ImportDecl, error) {
	if values["package"] == "" || values["constraint"] == "" || values["semantic_id"] == "" || values["type"] == "" || values["stage"] == "" {
		return ImportDecl{}, fmt.Errorf("import requires package, constraint, semantic_id, type, stage")
	}
	return ImportDecl{
		Package: values["package"], PackageDigest: values["package_digest"], Constraint: values["constraint"],
		SemanticID: values["semantic_id"], Type: values["type"], Stage: values["stage"],
		Capabilities: parseRows(values["capability"]), Effects: parseRows(values["effect"]), Owner: values["owner"],
		Identity: ImportIdentity{
			RepositoryIdentity: values["repository"], ImmutableReleaseID: values["release_id"],
			AnnotatedTagObject: values["tag_object"], AnnotatedTagTarget: values["tag_target"],
			AssetID: values["asset_id"], AssetDigest: values["asset_digest"],
			PackageSemanticID: values["package_semantic_id"], ExportedSymbolSetDigest: values["symbols_digest"],
			ContractDigest: values["contract_digest"], GoToolchainDigest: values["toolchain_digest"],
		},
	}, nil
}

func LoadJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s: %w", filepath.Base(path), err)
	}
	return nil
}

func LoadCatalog(path, lockPath string) (Catalog, CatalogLock, error) {
	var catalog Catalog
	if err := LoadJSON(path, &catalog); err != nil {
		return Catalog{}, CatalogLock{}, err
	}
	if catalog.Schema != CatalogSchema {
		return Catalog{}, CatalogLock{}, fmt.Errorf("unexpected catalog schema %q", catalog.Schema)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Catalog{}, CatalogLock{}, err
	}
	var lock CatalogLock
	if err := LoadJSON(lockPath, &lock); err != nil {
		return Catalog{}, CatalogLock{}, err
	}
	if lock.Schema != CatalogLockSchema || lock.CatalogDigest == "" {
		return Catalog{}, CatalogLock{}, fmt.Errorf("invalid catalog lock")
	}
	if digestBytes(data) != lock.CatalogDigest {
		return Catalog{}, CatalogLock{}, fmt.Errorf("catalog immutable digest mismatch")
	}
	return catalog, lock, nil
}
