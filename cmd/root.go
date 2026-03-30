package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/likeajohny/versios/ecosystem"
	"github.com/likeajohny/versios/git"
	"github.com/likeajohny/versios/prompt"
	"github.com/likeajohny/versios/version"
	"github.com/spf13/cobra"
)

var appVersion = "dev"

var yesFlag bool

var rootCmd = &cobra.Command{
	Use:   "vrs [version|major|minor|patch]",
	Short: "Bump project version across ecosystems",
	Long: `Versios (vrs) detects your project type and bumps the version
in the appropriate manifest file (package.json, composer.json, Cargo.toml, etc.)

Examples:
  vrs              Show current version and detected ecosystem
  vrs patch        Bump patch version (1.2.3 → 1.2.4)
  vrs minor        Bump minor version (1.2.3 → 1.3.0)
  vrs major        Bump major version (1.2.3 → 2.0.0)
  vrs 2.1.0        Set version to 2.1.0`,
	Args:         cobra.MaximumNArgs(1),
	RunE:         run,
	SilenceUsage: true,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip all confirmations")
	rootCmd.Version = appVersion
}

func run(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Detect ecosystems
	detected := ecosystem.DetectAll(dir)
	if len(detected) == 0 {
		return fmt.Errorf("no supported project detected in current directory\n  Checked for: package.json, composer.json, Cargo.toml, go.mod")
	}

	// Select which ecosystems to bump
	var selected []ecosystem.Ecosystem
	if len(detected) == 1 {
		selected = detected
	} else {
		names := make([]string, len(detected))
		for i, e := range detected {
			names[i] = e.Name() + " (" + e.PackageManager(dir) + ")"
		}

		if yesFlag {
			selected = detected
		} else {
			indices := prompt.SelectOrAll("Multiple projects detected. Which to update?", names)
			for _, i := range indices {
				selected = append(selected, detected[i])
			}
		}
	}

	// Use the first ecosystem's version as the reference for resolving
	// semantic bump keywords (major/minor/patch) and safety warnings.
	refEco := selected[0]
	currentStr, err := refEco.ReadVersion(dir)
	if err != nil {
		return fmt.Errorf("could not read version from %s: %w", refEco.Name(), err)
	}

	current, err := version.Parse(currentStr)
	if err != nil {
		return fmt.Errorf("current version %q is not valid semver: %w", currentStr, err)
	}

	// Show detected ecosystems and current version
	if len(selected) == 1 {
		fmt.Fprintf(os.Stderr, "\n  Detected: %s project (%s)\n", refEco.Name(), refEco.PackageManager(dir))
	} else {
		names := make([]string, len(selected))
		for i, e := range selected {
			names[i] = e.Name()
		}
		fmt.Fprintf(os.Stderr, "\n  Detected: %s\n", strings.Join(names, ", "))
	}
	fmt.Fprintf(os.Stderr, "  Current version: %s\n", current)

	// No-args mode: just show info
	if len(args) == 0 {
		fmt.Println(current.String())
		return nil
	}

	// Resolve target version
	target, err := version.Resolve(current, args[0])
	if err != nil {
		return fmt.Errorf("invalid version argument %q: %w", args[0], err)
	}

	fmt.Fprintf(os.Stderr, "  Target version:  %s\n\n", target)

	// Safety warnings
	if version.IsMajorBump(current, target) {
		if !yesFlag && !prompt.Confirm(fmt.Sprintf("This is a MAJOR version bump (%d.x → %d.x). Continue?", current.Major, target.Major), true) {
			fmt.Fprintln(os.Stderr, "\n  Aborted.")
			return nil
		}
	}

	if version.IsDowngrade(current, target) {
		if !yesFlag && !prompt.Confirm(fmt.Sprintf("This is a DOWNGRADE (%s → %s). Continue?", current, target), false) {
			fmt.Fprintln(os.Stderr, "\n  Aborted.")
			return nil
		}
	}

	if version.IsPreRelease(target) {
		if !yesFlag && !prompt.Confirm(fmt.Sprintf("Target is a pre-release version (%s). Continue?", target), true) {
			fmt.Fprintln(os.Stderr, "\n  Aborted.")
			return nil
		}
	}

	// Bump each selected ecosystem
	var changedFiles []string
	for _, eco := range selected {
		if err := eco.WriteVersion(dir, target.String()); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ Failed to update %s: %v\n", eco.Name(), err)
			continue
		}

		if eco.WritesFiles() {
			fmt.Fprintf(os.Stderr, "  ✓ Updated %s\n", eco.ManifestFile())
			changedFiles = append(changedFiles, eco.ManifestFile())
		}

		// Lock file handling per ecosystem
		lockOpts := eco.LockFileOptions(dir)
		if len(lockOpts) > 0 {
			labels := make([]string, len(lockOpts))
			for i, opt := range lockOpts {
				labels[i] = opt.Label
			}

			var choice int
			if yesFlag {
				choice = 0
			} else {
				choice = prompt.Select(
					fmt.Sprintf("[%s] Lock file may be out of sync. How should we update it?", eco.Name()),
					labels,
				)
			}

			sel := lockOpts[choice]
			if sel.Strategy == ecosystem.LockRunInstall && sel.Command != nil {
				fmt.Fprintf(os.Stderr, "\n  Running %s...\n", strings.Join(sel.Command, " "))
				c := exec.Command(sel.Command[0], sel.Command[1:]...)
				c.Dir = dir
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				if err := c.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "  ⚠ %s failed: %v (version file was still updated)\n", strings.Join(sel.Command, " "), err)
				} else {
					fmt.Fprintf(os.Stderr, "  ✓ Lock file updated\n")
				}
			}
		}
	}

	// Git operations (once for all ecosystems)
	if git.IsRepo(dir) {
		doGit := yesFlag || prompt.Confirm("Create git commit and tag?", true)
		if doGit {
			if len(changedFiles) > 0 {
				if err := git.CommitVersionBump(dir, target.String(), changedFiles); err != nil {
					fmt.Fprintf(os.Stderr, "  ⚠ Git commit failed: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "  ✓ Created commit: \"bump version to %s\"\n", target)
				}
			}

			annotated := true // default to annotated
			var tagMessage string
			if !yesFlag {
				tagType := prompt.Select("Tag type?", []string{
					"Annotated (recommended)",
					"Lightweight",
				})
				annotated = tagType == 0

				if annotated {
					tagMessage = prompt.Input("Tag message?", target.StringWithV())
				}
			}

			if err := git.CreateTag(dir, target.String(), annotated, tagMessage); err != nil {
				fmt.Fprintf(os.Stderr, "  ⚠ Git tag failed: %v\n", err)
			} else {
				kind := "annotated"
				if !annotated {
					kind = "lightweight"
				}
				fmt.Fprintf(os.Stderr, "  ✓ Created %s tag: %s\n", kind, target.StringWithV())
			}
		}
	}

	fmt.Fprintf(os.Stderr, "\n  Done! Version bumped from %s → %s\n\n", current, target)
	return nil
}
