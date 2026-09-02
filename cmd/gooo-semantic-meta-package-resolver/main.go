package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-semantic-meta-package-resolver/internal/resolver"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(64)
	}
	switch os.Args[1] {
	case "resolve", "run":
		if err := resolveCommand(os.Args[2:]); err != nil {
			fatal(err)
		}
	case "execute":
		if err := executeCommand(os.Args[2:]); err != nil {
			fatal(err)
		}
	case "conformance":
		if err := conformanceCommand(os.Args[2:]); err != nil {
			fatal(err)
		}
	case "conformance-v2":
		if err := conformanceV2Command(os.Args[2:]); err != nil {
			fatal(err)
		}
	default:
		usage()
		os.Exit(64)
	}
}

func resolveCommand(args []string) error {
	flags := flag.NewFlagSet("resolve", flag.ContinueOnError)
	sourcePath := flags.String("source", "", "root .gooo source")
	contractPath := flags.String("contract", "", "authoritative contract .gooo source")
	catalogPath := flags.String("catalog", "", "immutable fixture catalog")
	catalogLockPath := flags.String("catalog-lock", "", "immutable catalog lock")
	lockPath := flags.String("lock", "", "optional previously generated lockfile")
	outPath := flags.String("out", "", "absolute caller-owned output directory")
	execute := flags.Bool("execute", false, "execute a CLOSED generated Go artifact")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *sourcePath == "" || *contractPath == "" || *catalogPath == "" || *catalogLockPath == "" || *outPath == "" {
		return fmt.Errorf("resolve requires --source, --contract, --catalog, --catalog-lock, and --out")
	}
	contract, err := resolver.ParseSource(*contractPath)
	if err != nil {
		return err
	}
	root, err := resolver.ParseSource(*sourcePath)
	if err != nil {
		return err
	}
	if contract.LanguageVersion == "v2" {
		if *lockPath != "" {
			return fmt.Errorf("v2 semantic imports do not accept a v1 lockfile")
		}
		catalog, catalogLock, err := resolver.LoadV2Catalog(*catalogPath, *catalogLockPath)
		if err != nil {
			return err
		}
		engine := resolver.V2Resolver{Contract: contract, Catalog: catalog, CatalogLock: catalogLock, CatalogRoot: filepath.Dir(*catalogPath)}
		resolution, err := engine.Resolve(root)
		if err != nil {
			return err
		}
		artifacts, err := resolver.GenerateV2Artifacts(resolution)
		if err != nil {
			return err
		}
		if err := resolver.WriteV2Artifacts(*outPath, artifacts); err != nil {
			return err
		}
		if *execute {
			if resolution.Status != resolver.Closed {
				return fmt.Errorf("cannot execute non-CLOSED v2 resolution: %s", resolution.Status)
			}
			execution, err := resolver.ExecuteV2Artifact(filepath.Join(*outPath, "linked-artifact.go"))
			if err != nil {
				return err
			}
			data, err := resolver.MarshalJSONForCLI(execution)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(*outPath, "execution.json"), data, 0o644); err != nil {
				return err
			}
		}
		fmt.Printf("resolution=%s ir=%s artifact=%s\n", resolution.Status, artifacts.IRDigest, artifacts.ArtifactDigest)
		return nil
	}
	catalog, catalogLock, err := resolver.LoadCatalog(*catalogPath, *catalogLockPath)
	if err != nil {
		return err
	}
	var provided *resolver.Lockfile
	if *lockPath != "" {
		var lock resolver.Lockfile
		if err := resolver.LoadJSON(*lockPath, &lock); err != nil {
			return err
		}
		provided = &lock
	}
	engine := resolver.Resolver{Contract: contract, Catalog: catalog, CatalogLock: catalogLock, CatalogRoot: filepath.Dir(*catalogPath)}
	resolution, err := engine.Resolve(root, provided)
	if err != nil {
		return err
	}
	if err := resolver.VerifyResolution(resolution); err != nil {
		return err
	}
	artifacts, err := resolver.GenerateArtifacts(resolution)
	if err != nil {
		return err
	}
	if err := resolver.WriteArtifacts(*outPath, artifacts); err != nil {
		return err
	}
	if *execute {
		if resolution.Status != resolver.Closed {
			return fmt.Errorf("cannot execute non-CLOSED resolution: %s", resolution.Status)
		}
		execution, err := resolver.ExecuteArtifact(filepath.Join(*outPath, "linked-artifact.go"))
		if err != nil {
			return err
		}
		data, err := resolver.MarshalJSONForCLI(execution)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(*outPath, "execution.json"), data, 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("resolution=%s lock=%s ir=%s artifact=%s\n", resolution.Status, artifacts.LockDigest, artifacts.IRDigest, artifacts.ArtifactDigest)
	return nil
}

func executeCommand(args []string) error {
	flags := flag.NewFlagSet("execute", flag.ContinueOnError)
	artifactPath := flags.String("artifact", "", "generated linked-artifact.go")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *artifactPath == "" {
		return fmt.Errorf("execute requires --artifact")
	}
	execution, err := resolver.ExecuteArtifact(*artifactPath)
	if err != nil {
		return err
	}
	data, err := resolver.MarshalJSONForCLI(execution)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}

func conformanceCommand(args []string) error {
	flags := flag.NewFlagSet("conformance", flag.ContinueOnError)
	rootPath := flags.String("root", ".", "repository root")
	fixturesPath := flags.String("fixtures", "fixtures", "fixture directory")
	outPath := flags.String("out", "", "absolute caller-owned output directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *outPath == "" {
		return fmt.Errorf("conformance requires --out")
	}
	result, err := resolver.RunConformance(*rootPath, *fixturesPath, *outPath)
	if err != nil {
		return err
	}
	fmt.Printf("conformance=%s cases=%d\n", result.Status, result.Cases)
	return nil
}

func conformanceV2Command(args []string) error {
	flags := flag.NewFlagSet("conformance-v2", flag.ContinueOnError)
	rootPath := flags.String("root", ".", "repository root")
	fixturesPath := flags.String("fixtures", "fixtures", "fixture directory")
	outPath := flags.String("out", "", "absolute caller-owned output directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *outPath == "" {
		return fmt.Errorf("conformance-v2 requires --out")
	}
	result, err := resolver.RunV2Conformance(*rootPath, *fixturesPath, *outPath)
	if err != nil {
		return err
	}
	fmt.Printf("conformance-v2=%s cases=%d CLOSED=%d UNKNOWN=%d REFUTED=%d\n", result.Status, result.Cases, result.Closed, result.Unknown, result.Refuted)
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: gooo-semantic-meta-package-resolver {resolve|run|execute|conformance|conformance-v2} [flags]")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
