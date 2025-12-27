package fs

import (
	"io/fs"
	"path/filepath"
)

type NoirProject struct {
	Root     string
	Manifest *NargoManifest
	Files    map[string]string // project-local .nr files
}

// findFilesRecursive walks the file tree rooted at root and returns a slice of all file paths.

func (p *NoirProject) LoadFiles() error {
	p.Files = make(map[string]string)
	err := filepath.WalkDir(p.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		//pass directories
		if d.IsDir() {
			return nil
		}

		// match *nr globbing pattern
		matched, _ := filepath.Match("*.nr", filepath.Base(path))
		if !matched {
			return nil
		}
		content := readFileNR(path)

		// add to our map
		p.Files[path] = content

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
