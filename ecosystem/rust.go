package ecosystem

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

func init() {
	Register(&Rust{})
}

type Rust struct{}

func (r *Rust) Name() string         { return "rust" }
func (r *Rust) WritesFiles() bool    { return true }
func (r *Rust) ManifestFile() string { return "Cargo.toml" }

func (r *Rust) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "Cargo.toml"))
	return err == nil
}

type cargoToml struct {
	Package struct {
		Version string `toml:"version"`
	} `toml:"package"`
}

func (r *Rust) ReadVersion(dir string) (string, error) {
	var cargo cargoToml
	if _, err := toml.DecodeFile(filepath.Join(dir, "Cargo.toml"), &cargo); err != nil {
		return "", fmt.Errorf("invalid Cargo.toml: %w", err)
	}

	if cargo.Package.Version == "" {
		return "", fmt.Errorf("no version field in Cargo.toml [package] section")
	}
	return cargo.Package.Version, nil
}

var tomlSectionRe = regexp.MustCompile(`^\[(.+)\]`)
var tomlVersionRe = regexp.MustCompile(`^(version\s*=\s*")([^"]+)(")`)

func (r *Rust) WriteVersion(dir string, version string) error {
	path := filepath.Join(dir, "Cargo.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	inPackage := false
	replaced := false

	for i, line := range lines {
		if m := tomlSectionRe.FindString(line); m != "" {
			inPackage = strings.TrimSpace(m) == "[package]"
		}
		if inPackage && !replaced && tomlVersionRe.MatchString(line) {
			sub := tomlVersionRe.FindStringSubmatch(line)
			lines[i] = sub[1] + version + sub[3]
			replaced = true
		}
	}

	if !replaced {
		return fmt.Errorf("could not find version field in Cargo.toml [package] section")
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), info.Mode())
}

func (r *Rust) PackageManager(dir string) string { return "cargo" }

func (r *Rust) LockFileOptions(dir string) []LockFileOption {
	if _, err := os.Stat(filepath.Join(dir, "Cargo.lock")); err != nil {
		return nil
	}

	return []LockFileOption{
		{
			Label:    "Run `cargo update` (recommended)",
			Strategy: LockRunInstall,
			Command:  []string{"cargo", "update"},
		},
		{
			Label:    "Skip",
			Strategy: LockNone,
		},
	}
}
