package fs

import (
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"maps"
	"path/filepath"
	"strings"
)

type Resolver struct {
	visited  map[string]bool         // path → visited
	projects map[string]*NoirProject // path → project
	AllFiles map[string]string       // global merged files
	crate    string                  //Crate name
}

func NewResolver() *Resolver {
	return &Resolver{
		visited:  make(map[string]bool),
		projects: make(map[string]*NoirProject),
		AllFiles: make(map[string]string),
		crate:    "",
	}
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
