package version

import (
	"fmt"
	"regexp"
	"strconv"
)

var semverRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-(.+))?$`)

type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
}

func Parse(s string) (Version, error) {
	m := semverRe.FindStringSubmatch(s)
	if m == nil {
		return Version{}, fmt.Errorf("invalid semver: %q", s)
	}

	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: m[4],
	}, nil
}

func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		s += "-" + v.PreRelease
	}
	return s
}

func (v Version) StringWithV() string {
	return "v" + v.String()
}

func BumpMajor(v Version) Version {
	return Version{Major: v.Major + 1}
}

func BumpMinor(v Version) Version {
	return Version{Major: v.Major, Minor: v.Minor + 1}
}

func BumpPatch(v Version) Version {
	return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
}

func Resolve(current Version, arg string) (Version, error) {
	switch arg {
	case "major":
		return BumpMajor(current), nil
	case "minor":
		return BumpMinor(current), nil
	case "patch":
		return BumpPatch(current), nil
	default:
		return Parse(arg)
	}
}

func IsMajorBump(from, to Version) bool {
	return to.Major > from.Major
}

func IsDowngrade(from, to Version) bool {
	if to.Major != from.Major {
		return to.Major < from.Major
	}
	if to.Minor != from.Minor {
		return to.Minor < from.Minor
	}
	if to.Patch != from.Patch {
		return to.Patch < from.Patch
	}
	// Same M.M.P: going from release to pre-release is a downgrade
	if from.PreRelease == "" && to.PreRelease != "" {
		return true
	}
	return false
}

func IsPreRelease(v Version) bool {
	return v.PreRelease != ""
}
