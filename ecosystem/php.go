package ecosystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func init() {
	Register(&PHP{})
}

type PHP struct{}

func (p *PHP) Name() string         { return "php" }
func (p *PHP) WritesFiles() bool    { return true }
func (p *PHP) ManifestFile() string { return "composer.json" }

func (p *PHP) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "composer.json"))
	return err == nil
}

func (p *PHP) ReadVersion(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "composer.json"))
	if err != nil {
		return "", err
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", fmt.Errorf("invalid composer.json: %w", err)
	}

	v, ok := pkg["version"].(string)
	if !ok || v == "" {
		return "", fmt.Errorf("no version field in composer.json")
	}
	return v, nil
}

func (p *PHP) WriteVersion(dir string, version string) error {
	path := filepath.Join(dir, "composer.json")
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	replaced := false
	result := JSONVersionRe.ReplaceAllFunc(data, func(match []byte) []byte {
		if replaced {
			return match
		}
		replaced = true
		sub := JSONVersionRe.FindSubmatch(match)
		return append(append(sub[1], []byte(version)...), sub[3]...)
	})

	if !replaced {
		return fmt.Errorf("could not find version field in composer.json")
	}

	return os.WriteFile(path, result, info.Mode())
}

func (p *PHP) PackageManager(dir string) string { return "composer" }

func (p *PHP) LockFileOptions(dir string) []LockFileOption {
	if _, err := os.Stat(filepath.Join(dir, "composer.lock")); err != nil {
		return nil
	}

	return []LockFileOption{
		{
			Label:    "Run `composer update --lock` (recommended)",
			Strategy: LockRunInstall,
			Command:  []string{"composer", "update", "--lock"},
		},
		{
			Label:    "Skip",
			Strategy: LockNone,
		},
	}
}
