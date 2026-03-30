package ecosystem

import "regexp"

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

// JSONVersionRe matches the "version" field in JSON files (package.json, composer.json).
var JSONVersionRe = regexp.MustCompile(`("version"\s*:\s*")([^"]+)(")`)

type Ecosystem interface {
	Name() string
	Detect(dir string) bool
	ReadVersion(dir string) (string, error)
	WriteVersion(dir string, version string) error
	WritesFiles() bool // whether WriteVersion modifies files (false for Go)
	ManifestFile() string // relative path to the manifest file (e.g. "package.json")
	LockFileOptions(dir string) []LockFileOption
	PackageManager(dir string) string
}
