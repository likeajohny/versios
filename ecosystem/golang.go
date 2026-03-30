package ecosystem

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/likeajohny/versios/git"
)

func init() {
	Register(&Golang{})
}

type Golang struct{}

func (g *Golang) Name() string         { return "go" }
func (g *Golang) WritesFiles() bool    { return false }
func (g *Golang) ManifestFile() string { return "go.mod" }

func (g *Golang) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}

func (g *Golang) ReadVersion(dir string) (string, error) {
	tag, err := git.LatestVersionTag(dir)
	if err != nil {
		return "0.0.0", nil
	}
	return strings.TrimPrefix(tag, "v"), nil
}

func (g *Golang) WriteVersion(dir string, version string) error {
	// Go versions live in git tags, not files. Tag creation is handled by the git layer.
	return nil
}

func (g *Golang) PackageManager(dir string) string { return "go" }

func (g *Golang) LockFileOptions(dir string) []LockFileOption {
	return nil
}
