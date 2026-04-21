package ecosystem

import (
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
	return readJSONVersion(dir, "package.json")
}

func (n *NodeJS) WriteVersion(dir string, version string) error {
	return writeJSONVersion(dir, "package.json", version)
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
			LockFile: lockFile,
		},
		{
			Label:    "Skip",
			Strategy: LockNone,
		},
	}
}
