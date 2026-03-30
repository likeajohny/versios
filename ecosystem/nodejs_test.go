package ecosystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNodeJSDetect(t *testing.T) {
	dir := t.TempDir()

	n := &NodeJS{}
	if n.Detect(dir) {
		t.Error("should not detect without package.json")
	}

	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{}`), 0644)
	if !n.Detect(dir) {
		t.Error("should detect with package.json")
	}
}

func TestNodeJSReadVersion(t *testing.T) {
	dir := t.TempDir()
	n := &NodeJS{}

	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
  "name": "test",
  "version": "1.2.3"
}`), 0644)

	v, err := n.ReadVersion(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "1.2.3" {
		t.Errorf("got %q, want %q", v, "1.2.3")
	}
}

func TestNodeJSReadVersionMissing(t *testing.T) {
	dir := t.TempDir()
	n := &NodeJS{}

	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "test"}`), 0644)

	_, err := n.ReadVersion(dir)
	if err == nil {
		t.Error("expected error for missing version field")
	}
}

func TestNodeJSWriteVersion(t *testing.T) {
	dir := t.TempDir()
	n := &NodeJS{}

	original := `{
  "name": "my-app",
  "version": "1.2.3",
  "description": "test"
}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(original), 0644)

	if err := n.WriteVersion(dir, "2.0.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	content := string(data)

	v, _ := n.ReadVersion(dir)
	if v != "2.0.0" {
		t.Errorf("version not updated: got %q", v)
	}

	if !strings.Contains(content, `"name": "my-app"`) {
		t.Error("other fields should be preserved")
	}
	if !strings.Contains(content, `"description": "test"`) {
		t.Error("other fields should be preserved")
	}
}

func TestNodeJSPackageManager(t *testing.T) {
	n := &NodeJS{}

	tests := []struct {
		lockFile string
		want     string
	}{
		{"pnpm-lock.yaml", "pnpm"},
		{"yarn.lock", "yarn"},
		{"package-lock.json", "npm"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, tt.lockFile), []byte{}, 0644)

			if got := n.PackageManager(dir); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	// No lock file -> npm default
	dir := t.TempDir()
	if got := n.PackageManager(dir); got != "npm" {
		t.Errorf("default should be npm, got %q", got)
	}
}

func TestNodeJSLockFileOptions(t *testing.T) {
	n := &NodeJS{}

	// With lock file
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pnpm-lock.yaml"), []byte{}, 0644)
	opts := n.LockFileOptions(dir)
	if len(opts) != 2 {
		t.Fatalf("expected 2 options, got %d", len(opts))
	}
	if opts[0].Strategy != LockRunInstall {
		t.Error("first option should be run_install")
	}

	// Without lock file
	dir2 := t.TempDir()
	if opts := n.LockFileOptions(dir2); opts != nil {
		t.Error("should return nil without lock file")
	}
}
