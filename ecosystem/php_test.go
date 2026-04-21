package ecosystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPHPDetect(t *testing.T) {
	dir := t.TempDir()
	p := &PHP{}

	if p.Detect(dir) {
		t.Error("should not detect without composer.json")
	}

	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{}`), 0644)
	if !p.Detect(dir) {
		t.Error("should detect with composer.json")
	}
}

func TestPHPReadVersion(t *testing.T) {
	dir := t.TempDir()
	p := &PHP{}

	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{
  "name": "vendor/package",
  "version": "3.1.0"
}`), 0644)

	v, err := p.ReadVersion(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "3.1.0" {
		t.Errorf("got %q, want %q", v, "3.1.0")
	}
}

func TestPHPReadVersionMissing(t *testing.T) {
	dir := t.TempDir()
	p := &PHP{}

	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{"name": "vendor/pkg"}`), 0644)

	_, err := p.ReadVersion(dir)
	if err == nil {
		t.Error("expected error for missing version field")
	}
}

func TestPHPWriteVersion(t *testing.T) {
	dir := t.TempDir()
	p := &PHP{}

	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{
    "name": "vendor/package",
    "version": "1.0.0",
    "require": {}
}`), 0644)

	if err := p.WriteVersion(dir, "2.0.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, _ := p.ReadVersion(dir)
	if v != "2.0.0" {
		t.Errorf("version not updated: got %q", v)
	}
}

func TestPHPLockFileOptions(t *testing.T) {
	p := &PHP{}

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "composer.lock"), []byte{}, 0644)
	opts := p.LockFileOptions(dir)
	if len(opts) != 2 {
		t.Fatalf("expected 2 options, got %d", len(opts))
	}
	if opts[0].LockFile != "composer.lock" {
		t.Errorf("expected lock file composer.lock, got %s", opts[0].LockFile)
	}

	dir2 := t.TempDir()
	if opts := p.LockFileOptions(dir2); opts != nil {
		t.Error("should return nil without lock file")
	}
}
