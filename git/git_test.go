package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "test")
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("init"), 0644)
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")
	return dir
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func TestIsRepo(t *testing.T) {
	dir := setupRepo(t)
	if !IsRepo(dir) {
		t.Error("should detect git repo")
	}

	notRepo := t.TempDir()
	if IsRepo(notRepo) {
		t.Error("should not detect non-repo")
	}
}

func TestLatestVersionTag(t *testing.T) {
	dir := setupRepo(t)

	_, err := LatestVersionTag(dir)
	if err == nil {
		t.Error("should error with no tags")
	}

	run(t, dir, "git", "tag", "v1.0.0")
	tag, err := LatestVersionTag(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v1.0.0" {
		t.Errorf("got %q, want %q", tag, "v1.0.0")
	}
}

func TestCommitVersionBump(t *testing.T) {
	dir := setupRepo(t)

	// Make a change
	os.WriteFile(filepath.Join(dir, "version.txt"), []byte("2.0.0"), 0644)

	if err := CommitVersionBump(dir, "2.0.0", []string{"version.txt"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify commit message
	out, _ := exec.Command("git", "-C", dir, "log", "-1", "--pretty=%s").Output()
	msg := string(out)
	if msg != "bump version to 2.0.0\n" {
		t.Errorf("commit message = %q, want %q", msg, "bump version to 2.0.0\n")
	}
}

func TestCreateTagLightweight(t *testing.T) {
	dir := setupRepo(t)

	if err := CreateTag(dir, "v1.0.0", false, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := exec.Command("git", "-C", dir, "tag", "-l", "v1.0.0").Output()
	if err != nil || string(out) == "" {
		t.Error("tag v1.0.0 should exist")
	}

	// Lightweight tags have type "commit"
	catOut, _ := exec.Command("git", "-C", dir, "cat-file", "-t", "v1.0.0").Output()
	if got := string(catOut); got != "commit\n" {
		t.Errorf("expected lightweight tag (commit), got %q", got)
	}
}

func TestCreateTagAnnotated(t *testing.T) {
	dir := setupRepo(t)

	if err := CreateTag(dir, "v2.0.0", true, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := exec.Command("git", "-C", dir, "tag", "-l", "v2.0.0").Output()
	if err != nil || string(out) == "" {
		t.Error("tag v2.0.0 should exist")
	}

	// Annotated tags have type "tag"
	catOut, _ := exec.Command("git", "-C", dir, "cat-file", "-t", "v2.0.0").Output()
	if got := string(catOut); got != "tag\n" {
		t.Errorf("expected annotated tag (tag), got %q", got)
	}
}

func TestCreateTagAnnotatedWithMessage(t *testing.T) {
	dir := setupRepo(t)

	if err := CreateTag(dir, "v3.0.0", true, "Release 3.0.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Annotated tags have type "tag"
	catOut, _ := exec.Command("git", "-C", dir, "cat-file", "-t", "v3.0.0").Output()
	if got := string(catOut); got != "tag\n" {
		t.Errorf("expected annotated tag (tag), got %q", got)
	}

	// Verify custom message
	msgOut, _ := exec.Command("git", "-C", dir, "tag", "-l", "-n1", "v3.0.0").Output()
	if !strings.Contains(string(msgOut), "Release 3.0.0") {
		t.Errorf("expected message containing 'Release 3.0.0', got %q", string(msgOut))
	}
}

func TestLatestVersionTagPlain(t *testing.T) {
	dir := setupRepo(t)

	run(t, dir, "git", "tag", "1.0.0")
	tag, err := LatestVersionTag(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "1.0.0" {
		t.Errorf("got %q, want %q", tag, "1.0.0")
	}
}

func TestDetectTagPrefix(t *testing.T) {
	t.Run("v-prefixed tags", func(t *testing.T) {
		dir := setupRepo(t)
		run(t, dir, "git", "tag", "v1.0.0")
		if got := DetectTagPrefix(dir); got != "v" {
			t.Errorf("got %q, want %q", got, "v")
		}
	})

	t.Run("plain tags", func(t *testing.T) {
		dir := setupRepo(t)
		run(t, dir, "git", "tag", "1.0.0")
		if got := DetectTagPrefix(dir); got != "" {
			t.Errorf("got %q, want %q", got, "")
		}
	})

	t.Run("no tags defaults to v", func(t *testing.T) {
		dir := setupRepo(t)
		if got := DetectTagPrefix(dir); got != "v" {
			t.Errorf("got %q, want %q", got, "v")
		}
	})
}
