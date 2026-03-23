package fs

import (
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Resolver struct {
	visited  map[string]bool         // path → visited
	projects map[string]*NoirProject // path → project
	AllFiles map[string]string       // global merged files
	crate    string                  //Crate name
	cacheDir string
	client   *http.Client
}

func NewResolver() *Resolver {
	cacheDir := defaultDependencyCacheDir()
	return &Resolver{
		visited:  make(map[string]bool),
		projects: make(map[string]*NoirProject),
		AllFiles: make(map[string]string),
		crate:    "",
		cacheDir: cacheDir,
		client:   http.DefaultClient,
	}
}

func newResolverForTest(cacheDir string, client *http.Client) *Resolver {
	if cacheDir == "" {
		cacheDir = defaultDependencyCacheDir()
	}
	if client == nil {
		client = http.DefaultClient
	}
	r := NewResolver()
	r.cacheDir = cacheDir
	r.client = client
	return r
}

func (r *Resolver) Resolve(root string) error {
	abs, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	if r.visited[abs] {
		return nil
	}
	r.visited[abs] = true

	manifest, err := parseNargo(abs)
	if err != nil {
		return err
	}

	project := &NoirProject{
		Root:     abs,
		Manifest: manifest,
		Files:    make(map[string]string),
	}

	if err := project.LoadFiles(); err != nil {
		return err
	}

	var crateName string
	var libName string
	if r.crate == "" {
		for k := range project.Files {
			if strings.Contains(k, "/main.nr") {
				crateName = k
			}
			if strings.Contains(k, "/lib.nr") {
				libName = k
			}
		}
		if crateName != "" {
			r.crate = crateName
		} else if libName != "" {
			r.crate = libName
		} else {
			return fmt.Errorf("Project root must contain\nmain.nr\nor\nlib.nr")
		}
	}
	// merge files globally
	maps.Copy(r.AllFiles, project.Files) // Copies all key-value pairs from src to dst

	for _, dep := range manifest.Dependencies {
		if path, ok := dep["path"]; ok {
			next := filepath.Join(abs, path)
			if err := r.Resolve(next); err != nil {
				return err
			}
			continue
		}
		if _, ok := dep["git"]; ok {
			next, err := r.resolveRemoteDependency(dep)
			if err != nil {
				return err
			}
			if err := r.Resolve(next); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Resolver) Serialize() ([]byte, error) {
	r.AllFiles["CrateName"] = r.crate
	b, err := msgpack.Marshal(r.AllFiles)
	if err != nil {
		return nil, err
	}
	return b, err
}

func defaultDependencyCacheDir() string {
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		cacheRoot = os.TempDir()
	}
	return filepath.Join(cacheRoot, "noir-go")
}
