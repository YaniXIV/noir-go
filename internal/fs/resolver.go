package fs

import (
	"path/filepath"
)

// ==============================
// Types
// ==============================

type Resolver struct {
	visited  map[string]bool         // path → visited
	projects map[string]*NoirProject // path → project
	AllFiles map[string]string       // global merged files
}

// ==============================
// Constructors
// ==============================

func NewResolver() *Resolver {
	return &Resolver{
		visited:  make(map[string]bool),
		projects: make(map[string]*NoirProject),
		AllFiles: make(map[string]string),
	}
}

// ==============================
// Resolver (global orchestration)
// ==============================

func (r *Resolver) Resolve(root string) error {
	abs, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	if r.visited[abs] {
		return nil
	}
	r.visited[abs] = true

	manifest := parseNargo(abs)

	project := &NoirProject{
		Root:     abs,
		Manifest: manifest,
		Files:    make(map[string]string),
	}

	if err := project.LoadFiles(); err != nil {
		return err
	}

	// merge files globally
	for path, contents := range project.Files {
		r.AllFiles[path] = contents
	}

	// recurse dependencies (SIMPLE version)
	for _, dep := range manifest.Dependencies {
		if path, ok := dep["path"]; ok {
			next := filepath.Join(abs, path)
			if err := r.Resolve(next); err != nil {
				return err
			}
		}
	}

	return nil
}
