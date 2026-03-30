package ecosystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func init() {
	Register(&NodeJS{})
}

type NodeJS struct{}

func (n *NodeJS) Name() string         { return "nodejs" }
func (n *NodeJS) WritesFiles() bool    { return true }
func (n *NodeJS) ManifestFile() string { return "package.json" }

func (n *NodeJS) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "package.json"))
	return err == nil
}

func (n *NodeJS) ReadVersion(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return "", err
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", fmt.Errorf("invalid package.json: %w", err)
	}

	v, ok := pkg["version"].(string)
	if !ok || v == "" {
		return "", fmt.Errorf("no version field in package.json")
	}
	return v, nil
}

func (n *NodeJS) WriteVersion(dir string, version string) error {
	path := filepath.Join(dir, "package.json")
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
		return fmt.Errorf("could not find version field in package.json")
	}

	return os.WriteFile(path, result, info.Mode())
}

func (n *NodeJS) PackageManager(dir string) string {
	if _, err := os.Stat(filepath.Join(dir, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(dir, "yarn.lock")); err == nil {
		return "yarn"
	}
	return "npm"
}

func (n *NodeJS) LockFileOptions(dir string) []LockFileOption {
	pm := n.PackageManager(dir)

	var lockFile string
	switch pm {
	case "pnpm":
		lockFile = "pnpm-lock.yaml"
	case "yarn":
		lockFile = "yarn.lock"
	default:
		lockFile = "package-lock.json"
	}

	if _, err := os.Stat(filepath.Join(dir, lockFile)); err != nil {
		return nil
	}

	return []LockFileOption{
		{
			Label:    fmt.Sprintf("Run `%s install` (recommended)", pm),
			Strategy: LockRunInstall,
			Command:  []string{pm, "install"},
		},
		{
			Label:    "Skip",
			Strategy: LockNone,
		},
	}
}
