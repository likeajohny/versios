package ecosystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

var ErrNoVersion = errors.New("no version field")

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
		return "", fmt.Errorf("%s: %w", filename, ErrNoVersion)
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
		result, err = insertJSONVersion(data, version)
		if err != nil {
			return fmt.Errorf("could not add version field to %s: %w", filename, err)
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, result, info.Mode())
}

// insertJSONVersion inserts a "version" field into a JSON file that doesn't have one.
func insertJSONVersion(data []byte, version string) ([]byte, error) {
	s := string(data)
	braceIdx := strings.Index(s, "{")
	if braceIdx < 0 {
		return nil, fmt.Errorf("no opening brace found")
	}

	nlIdx := strings.Index(s[braceIdx:], "\n")
	if nlIdx < 0 {
		// Single-line JSON like `{"name": "test"}` — expand it
		return []byte(fmt.Sprintf("{\n  \"version\": \"%s\",\n  %s\n}", version, strings.TrimSpace(s[braceIdx+1:len(s)-1]))), nil
	}

	insertPos := braceIdx + nlIdx + 1

	// Detect indentation from the next non-empty line
	indent := "  "
	for i := insertPos; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			continue
		}
		indent = s[insertPos:i]
		break
	}

	line := fmt.Sprintf("%s\"version\": \"%s\",\n", indent, version)
	return []byte(s[:insertPos] + line + s[insertPos:]), nil
}
