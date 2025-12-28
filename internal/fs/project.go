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

func (p *NoirProject) LoadFiles() error {
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
