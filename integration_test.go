package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func projectRoot() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Dir(f)
}

func buildVrs(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "vrs")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = projectRoot()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func setupGitDir(t *testing.T, dir string) {
	t.Helper()
	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.email", "test@test.com")
	gitRun(t, dir, "config", "user.name", "test")
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-m", "init")
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func runVrs(t *testing.T, bin, dir string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	return stdout.String(), stderr.String(), exitCode
}

func TestIntegrationNodeJSPatch(t *testing.T) {
	bin := buildVrs(t)
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
  "name": "test-app",
  "version": "1.2.3"
}`), 0644)
	setupGitDir(t, dir)

	_, stderr, code := runVrs(t, bin, dir, "patch", "--yes")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstderr: %s", code, stderr)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	var pkg map[string]interface{}
	json.Unmarshal(data, &pkg)
	if pkg["version"] != "1.2.4" {
		t.Errorf("expected version 1.2.4, got %v", pkg["version"])
	}

	// Check git tag was created
	out, _ := exec.Command("git", "-C", dir, "tag", "-l", "v1.2.4").Output()
	if strings.TrimSpace(string(out)) != "v1.2.4" {
		t.Error("expected git tag v1.2.4")
	}
}

func TestIntegrationPHPExplicit(t *testing.T) {
	bin := buildVrs(t)
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{
  "name": "vendor/pkg",
  "version": "2.0.0"
}`), 0644)
	setupGitDir(t, dir)

	_, stderr, code := runVrs(t, bin, dir, "2.1.0", "--yes")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstderr: %s", code, stderr)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "composer.json"))
	var pkg map[string]interface{}
	json.Unmarshal(data, &pkg)
	if pkg["version"] != "2.1.0" {
		t.Errorf("expected version 2.1.0, got %v", pkg["version"])
	}
}

func TestIntegrationRust(t *testing.T) {
	bin := buildVrs(t)
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(`[package]
name = "test-crate"
version = "0.1.0"
edition = "2021"
`), 0644)
	setupGitDir(t, dir)

	_, stderr, code := runVrs(t, bin, dir, "minor", "--yes")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstderr: %s", code, stderr)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "Cargo.toml"))
	if !strings.Contains(string(data), `version = "0.2.0"`) {
		t.Errorf("expected version 0.2.0 in Cargo.toml, got:\n%s", data)
	}
}

func TestIntegrationNoArgs(t *testing.T) {
	bin := buildVrs(t)
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "x", "version": "5.0.1"}`), 0644)

	stdout, _, code := runVrs(t, bin, dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if strings.TrimSpace(stdout) != "5.0.1" {
		t.Errorf("expected stdout '5.0.1', got %q", stdout)
	}
}

func TestIntegrationNoProject(t *testing.T) {
	bin := buildVrs(t)
	dir := t.TempDir()

	_, _, code := runVrs(t, bin, dir, "patch")
	if code == 0 {
		t.Error("expected non-zero exit code for empty directory")
	}
}

func TestIntegrationInvalidVersion(t *testing.T) {
	bin := buildVrs(t)
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "x", "version": "1.0.0"}`), 0644)

	_, _, code := runVrs(t, bin, dir, "not-a-version")
	if code == 0 {
		t.Error("expected non-zero exit code for invalid version")
	}
}

func TestIntegrationMultiEcosystem(t *testing.T) {
	bin := buildVrs(t)
	dir := t.TempDir()

	// Create both package.json and composer.json in the same directory
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
  "name": "test-app",
  "version": "1.0.0"
}`), 0644)
	os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{
  "name": "vendor/pkg",
  "version": "1.0.0"
}`), 0644)
	setupGitDir(t, dir)

	// --yes selects all ecosystems when multiple detected
	_, stderr, code := runVrs(t, bin, dir, "1.1.0", "--yes")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstderr: %s", code, stderr)
	}

	// Both files should be updated
	nodeData, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	var nodePkg map[string]interface{}
	json.Unmarshal(nodeData, &nodePkg)
	if nodePkg["version"] != "1.1.0" {
		t.Errorf("package.json: expected 1.1.0, got %v", nodePkg["version"])
	}

	phpData, _ := os.ReadFile(filepath.Join(dir, "composer.json"))
	var phpPkg map[string]interface{}
	json.Unmarshal(phpData, &phpPkg)
	if phpPkg["version"] != "1.1.0" {
		t.Errorf("composer.json: expected 1.1.0, got %v", phpPkg["version"])
	}

	// Git commit should include both files
	out, _ := exec.Command("git", "-C", dir, "diff-tree", "--no-commit-id", "--name-only", "-r", "HEAD").Output()
	files := strings.TrimSpace(string(out))
	if !strings.Contains(files, "package.json") {
		t.Error("git commit should include package.json")
	}
	if !strings.Contains(files, "composer.json") {
		t.Error("git commit should include composer.json")
	}

	// One tag for the version
	tagOut, _ := exec.Command("git", "-C", dir, "tag", "-l", "v1.1.0").Output()
	if strings.TrimSpace(string(tagOut)) != "v1.1.0" {
		t.Error("expected git tag v1.1.0")
	}
}

func TestIntegrationGoProject(t *testing.T) {
	bin := buildVrs(t)
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	setupGitDir(t, dir)

	_, stderr, code := runVrs(t, bin, dir, "1.0.0", "--yes")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstderr: %s", code, stderr)
	}

	// Check git tag was created
	out, _ := exec.Command("git", "-C", dir, "tag", "-l", "v1.0.0").Output()
	if strings.TrimSpace(string(out)) != "v1.0.0" {
		t.Error("expected git tag v1.0.0")
	}
}
