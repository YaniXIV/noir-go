package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNargoParse(t *testing.T) {
	manifest, err := parseNargo(filepath.Join("..", "compiler", "noirtest"))
	if err != nil {
		t.Fatalf("parseNargo failed: %v", err)
	}
	if manifest == nil {
		t.Fatalf("parseNargo returned nil manifest")
	}
}

func TestNargoParseMissing(t *testing.T) {
	tmp := t.TempDir()
	if _, err := parseNargo(tmp); err == nil {
		t.Fatalf("expected error for missing Nargo.toml")
	}
}

func TestResolverResolveAndSerialize(t *testing.T) {
	r := NewResolver()
	root := filepath.Join("..", "compiler", "noirtest")

	if err := r.Resolve(root); err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	serialized, err := r.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}
	if len(serialized) == 0 {
		t.Fatalf("Serialize returned empty bytes")
	}

	hasMain := false
	for k := range r.AllFiles {
		if filepath.Base(k) == "main.nr" {
			hasMain = true
			break
		}
	}
	if !hasMain {
		t.Fatalf("expected main.nr to be loaded into AllFiles")
	}

	if r.AllFiles["CrateName"] == "" {
		t.Fatalf("expected CrateName to be set")
	}
}

func TestResolverMissingEntry(t *testing.T) {
	tmp := t.TempDir()
	err := os.WriteFile(filepath.Join(tmp, "Nargo.toml"), []byte("[package]\nname = \"tmp\"\ntype = \"bin\"\n\n[dependencies]\n"), 0644)
	if err != nil {
		t.Fatalf("write Nargo.toml: %v", err)
	}

	r := NewResolver()
	if err := r.Resolve(tmp); err == nil {
		t.Fatalf("expected error when main.nr or lib.nr is missing")
	}
}

func TestReadFileNROversized(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "big.nr")

	data := make([]byte, maxNoirFileSize+1)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write oversized noir file: %v", err)
	}

	if _, err := readFileNR(path); err == nil {
		t.Fatalf("expected oversized file error")
	}
}
