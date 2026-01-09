package fs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	maxNoirFileSize = 5 << 20 // 5 MB
	configFile      = "Nargo.toml"
)

type NargoManifest struct {
	Package      any                          `toml:"package,omiempty"` //*PackageConfig
	Workspace    any                          `toml:"workspace,omitey"` //*WorkspaceConfig
	Dependencies map[string]map[string]string `toml:"dependencies"`
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
type WorkspaceConfig struct {
	Members       []string `toml:"members"`
	DefaultMember string   `toml:"default-member,omitempty"`
}

// parses a nargo.toml file into a struct
func parseNargo(fp string) (*NargoManifest, error) {
	var nargoFile NargoManifest

	filePath := filepath.Join(fp, configFile)
	fileExists := fileExists(filePath)
	if !fileExists {
		return nil, fmt.Errorf("cannot find a nargo.toml for %v", fp)
	} else {
		log.Println("this file exists!")
	}

	if _, err := toml.DecodeFile(filePath, &nargoFile); err != nil {
		return nil, err
	}
	fmt.Println(nargoFile)
	return &nargoFile, nil
}

// read a singular .nr file into memory.
func readFileNR(filePath string) string {
	//make sure the file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Fatal(err)
	}
	//check file size before reading into mem. limit of 5mb per file.
	if fileInfo.Size() > maxNoirFileSize {
		panic(fmt.Sprintf("File size for %s too big", filePath))
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return string(data)

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
