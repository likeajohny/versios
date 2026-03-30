package ecosystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type LockStrategy string

const (
	LockRunInstall LockStrategy = "run_install"
	LockNone       LockStrategy = "none"
)

type LockFileOption struct {
	Label    string
	Strategy LockStrategy
	Command  []string // e.g. ["pnpm", "install"], nil for none
}

var jsonVersionRe = regexp.MustCompile(`("version"\s*:\s*")([^"]+)(")`)

type Ecosystem interface {
	Name() string
	Detect(dir string) bool
	ReadVersion(dir string) (string, error)
	WriteVersion(dir string, version string) error
	WritesFiles() bool
	ManifestFile() string
	LockFileOptions(dir string) []LockFileOption
	PackageManager(dir string) string
}

// readJSONVersion reads the "version" field from a JSON manifest file.
func readJSONVersion(dir, filename string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return "", err
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", fmt.Errorf("invalid %s: %w", filename, err)
	}

	v, ok := pkg["version"].(string)
	if !ok || v == "" {
		return "", fmt.Errorf("no version field in %s", filename)
	}
	return v, nil
}

// writeJSONVersion replaces the first "version" field value in a JSON manifest file,
// preserving all other formatting.
func writeJSONVersion(dir, filename, version string) error {
	path := filepath.Join(dir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	replaced := false
	result := jsonVersionRe.ReplaceAllFunc(data, func(match []byte) []byte {
		if replaced {
			return match
		}
		replaced = true
		sub := jsonVersionRe.FindSubmatch(match)
		return append(append(sub[1], []byte(version)...), sub[3]...)
	})

	if !replaced {
		return fmt.Errorf("could not find version field in %s", filename)
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, result, info.Mode())
}
