package ecosystem

import (
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
	return readJSONVersion(dir, "composer.json")
}

func (p *PHP) WriteVersion(dir string, version string) error {
	return writeJSONVersion(dir, "composer.json", version)
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
