package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func IsRepo(dir string) bool {
	err := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree").Run()
	return err == nil
}

func LatestVersionTag(dir string) (string, error) {
	// Try v-prefixed tags first
	out, err := exec.Command("git", "-C", dir, "describe", "--tags", "--abbrev=0", "--match", "v*").Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}

	// Fall back to plain numeric tags
	out, err = exec.Command("git", "-C", dir, "describe", "--tags", "--abbrev=0", "--match", "[0-9]*").Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}

	return "", fmt.Errorf("no version tags found")
}

func DetectTagPrefix(dir string) string {
	tag, err := LatestVersionTag(dir)
	if err != nil {
		return "v"
	}
	if strings.HasPrefix(tag, "v") {
		return "v"
	}
	return ""
}

func CommitVersionBump(dir string, version string, files []string) error {
	args := append([]string{"-C", dir, "add"}, files...)
	add := exec.Command("git", args...)
	if out, err := add.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %s", strings.TrimSpace(string(out)))
	}

	commit := exec.Command("git", "-C", dir, "commit", "-m", fmt.Sprintf("bump version to %s", version))
	if out, err := commit.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %s", strings.TrimSpace(string(out)))
	}

	return nil
}

func CreateTag(dir string, tag string, annotated bool, message string) error {
	var cmd *exec.Cmd
	if annotated {
		msg := tag
		if message != "" {
			msg = message
		}
		cmd = exec.Command("git", "-C", dir, "tag", "-a", tag, "-m", msg)
	} else {
		cmd = exec.Command("git", "-C", dir, "tag", tag)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git tag failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}
