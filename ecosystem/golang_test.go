package ecosystem

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGolangDetect(t *testing.T) {
	dir := t.TempDir()
	g := &Golang{}

	if g.Detect(dir) {
		t.Error("should not detect without go.mod")
	}

	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)
	if !g.Detect(dir) {
		t.Error("should detect with go.mod")
	}
}

func TestGolangReadVersionNoTags(t *testing.T) {
	dir := t.TempDir()
	g := &Golang{}

	// Init a git repo with no tags
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)
	run(t, dir, "git", "init")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init", "--allow-empty")

	v, err := g.ReadVersion(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "0.0.0" {
		t.Errorf("expected 0.0.0 for no tags, got %q", v)
	}
}

func TestGolangReadVersionWithTag(t *testing.T) {
	dir := t.TempDir()
	g := &Golang{}

	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)
	run(t, dir, "git", "init")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")
	run(t, dir, "git", "tag", "v1.5.2")

	v, err := g.ReadVersion(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "1.5.2" {
		t.Errorf("got %q, want %q", v, "1.5.2")
	}
}

func TestGolangWriteVersionIsNoop(t *testing.T) {
	dir := t.TempDir()
	g := &Golang{}

	err := g.WriteVersion(dir, "1.0.0")
	if err != nil {
		t.Errorf("WriteVersion should be no-op, got error: %v", err)
	}
}

func TestGolangLockFileOptions(t *testing.T) {
	g := &Golang{}
	if opts := g.LockFileOptions(t.TempDir()); opts != nil {
		t.Error("Go should have no lock file options")
	}
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com", "GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}
