package ecosystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRustDetect(t *testing.T) {
	dir := t.TempDir()
	r := &Rust{}

	if r.Detect(dir) {
		t.Error("should not detect without Cargo.toml")
	}

	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(`[package]
name = "test"
version = "0.1.0"
`), 0644)
	if !r.Detect(dir) {
		t.Error("should detect with Cargo.toml")
	}
}

func TestRustReadVersion(t *testing.T) {
	dir := t.TempDir()
	r := &Rust{}

	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(`[package]
name = "my-crate"
version = "0.3.7"
edition = "2021"

[dependencies]
serde = "1.0"
`), 0644)

	v, err := r.ReadVersion(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "0.3.7" {
		t.Errorf("got %q, want %q", v, "0.3.7")
	}
}

func TestRustWriteVersion(t *testing.T) {
	dir := t.TempDir()
	r := &Rust{}

	original := `[package]
name = "my-crate"
version = "0.3.7"
edition = "2021"

# Some comment
[dependencies]
version = "1.0"
serde = { version = "1.0", features = ["derive"] }
`
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(original), 0644)

	if err := r.WriteVersion(dir, "1.0.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, _ := r.ReadVersion(dir)
	if v != "1.0.0" {
		t.Errorf("version not updated: got %q", v)
	}

	// Verify it only changed the [package] version, not [dependencies]
	data, _ := os.ReadFile(filepath.Join(dir, "Cargo.toml"))
	content := string(data)
	if !strings.Contains(content, `version = "1.0"`) {
		t.Error("dependency version should not be modified")
	}
	if !strings.Contains(content, "# Some comment") {
		t.Error("comments should be preserved")
	}
}

func TestRustWriteVersionNoPackageSection(t *testing.T) {
	dir := t.TempDir()
	r := &Rust{}

	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(`[dependencies]
serde = "1.0"
`), 0644)

	err := r.WriteVersion(dir, "1.0.0")
	if err == nil {
		t.Error("expected error when no [package] section")
	}
}

func TestRustLockFileOptions(t *testing.T) {
	r := &Rust{}

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Cargo.lock"), []byte{}, 0644)
	opts := r.LockFileOptions(dir)
	if len(opts) != 2 {
		t.Fatalf("expected 2 options, got %d", len(opts))
	}

	dir2 := t.TempDir()
	if opts := r.LockFileOptions(dir2); opts != nil {
		t.Error("should return nil without lock file")
	}
}
