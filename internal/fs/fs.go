package fs

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const maxNoirFileSize = 5 << 20 // 5 MB
const configFile = "Nargo.toml"

type DependencyKind int

const (
	DepUnknown DependencyKind = iota
	DepVersion
	DepGit
	DepPath
)

type NargoManifest struct {
	Package      *PackageConfig            `toml:"package,omitempty"`
	Workspace    *WorkspaceConfig          `toml:"workspace,omitempty"`
	Dependencies map[string]DependencySpec `toml:"dependencies,omitempty"`
}

type PackageConfig struct {
	Name                     string   `toml:"name"`
	Type                     string   `toml:"type"` // bin | lib | contract
	Authors                  []string `toml:"authors,omitempty"`
	CompilerVersion          string   `toml:"compiler_version,omitempty"`
	CompilerUnstableFeatures []string `toml:"compiler_unstable_features,omitempty"`
	Description              string   `toml:"description,omitempty"`
	Entry                    string   `toml:"entry,omitempty"`
	Backend                  string   `toml:"backend,omitempty"`
	License                  string   `toml:"license,omitempty"`
	ExpressionWidth          int      `toml:"expression_width,omitempty"`
}

type DependencySpec struct {
	Kind DependencyKind

	// git
	Git string
	Tag string
	Rev string

	// path
	Path string

	// version shorthand
	Version string
}

type WorkspaceConfig struct {
	Members       []string `toml:"members"`
	DefaultMember string   `toml:"default-member,omitempty"`
}

/*
func collectFiles(projectRoot string) (map[string]string, error) {
	files := map[string]string{}
	err := filepath.Walk(projectRoot, func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".nr") {
			return nil
		}
		rel, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files["/"+rel] = string(data) // convert to virtual FS path
		return nil
	})
	return files, err
}

*/

// findFilesRecursive walks the file tree rooted at root and returns a slice of all file paths.
func findFilesRecursive(root string) (map[string]string, error) {
	files := map[string]string{}
	var files2 []string
	print(files2)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
		files[path] = content

		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func readFileNR(filePath string) string {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Fatal(err)
	}
	if fileInfo.Size() > maxNoirFileSize {
		panic(fmt.Sprintf("File size for %s too big", filePath))
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return string(data)

}

func parseNargo(filePath string) *NargoManifest {
	var nargoFile NargoManifest

	filePath = filepath.Join(filePath, configFile)
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Current directory:", dir)
	fileExists := fileExists(filePath)
	if !fileExists {
		log.Fatalf("Invalid file path to Nargo.toml \n%v", filePath)
	} else {
		log.Println("this file exists!")
	}

	if _, err := toml.DecodeFile(filePath, &nargoFile); err != nil {
		log.Fatal(err)
	}
	fmt.Println(nargoFile)
	return &nargoFile

}

// fileExists checks if a file or directory exists.
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true // File exists, no error
	}
	// Check if the error is specifically because the file does not exist
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	// A different error occurred (e.g., permission denied)
	// You might want to log this error in a real application
	return false
}

func (d *DependencySpec) UnmarshalTOML(v interface{}) error {
	switch val := v.(type) {

	// poseidon = "v0.1.0"
	case string:
		d.Kind = DepVersion
		d.Version = val
		return nil

	// poseidon = { ... }
	case map[string]interface{}:
		if path, ok := val["path"].(string); ok {
			d.Kind = DepPath
			d.Path = path
			return nil
		}

		if git, ok := val["git"].(string); ok {
			d.Kind = DepGit
			d.Git = git
			if tag, ok := val["tag"].(string); ok {
				d.Tag = tag
			}
			if rev, ok := val["rev"].(string); ok {
				d.Rev = rev
			}
			return nil
		}

		return fmt.Errorf("unknown dependency format: %#v", val)

	default:
		return fmt.Errorf("invalid dependency value: %T", v)
	}
}
func (n *NargoManifest) resolveDependencies() {
	for name, dep := range n.Dependencies {
		switch dep.Kind {
		case DepGit:
			fmt.Println(name, "is a git dependency:", dep.Git)
			n.resolveRemote()
		case DepPath:
			fmt.Println(name, "is a path dependency:", dep.Path)
			n.resolveLocal()
		case DepVersion:
			fmt.Println(name, "is a version dependency:", dep.Version)

		}
	}
}

func (n *NargoManifest) resolveRemote() {
	fmt.Println("Handle gi thub stuff here .")
}

func (n *NargoManifest) resolveLocal() {

}
