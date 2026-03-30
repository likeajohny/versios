package version

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input   string
		want    Version
		wantErr bool
	}{
		{"1.2.3", Version{1, 2, 3, ""}, false},
		{"v1.2.3", Version{1, 2, 3, ""}, false},
		{"0.0.1", Version{0, 0, 1, ""}, false},
		{"1.0.0-rc.1", Version{1, 0, 0, "rc.1"}, false},
		{"v2.0.0-beta.3", Version{2, 0, 0, "beta.3"}, false},
		{"10.20.30", Version{10, 20, 30, ""}, false},
		{"abc", Version{}, true},
		{"1.2", Version{}, true},
		{"1.2.3.4", Version{}, true},
		{"", Version{}, true},
		{"v", Version{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		v    Version
		want string
	}{
		{Version{1, 2, 3, ""}, "1.2.3"},
		{Version{1, 0, 0, "rc.1"}, "1.0.0-rc.1"},
		{Version{0, 0, 0, ""}, "0.0.0"},
	}

	for _, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Errorf("Version%v.String() = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestStringWithV(t *testing.T) {
	v := Version{1, 2, 3, ""}
	if got := v.StringWithV(); got != "v1.2.3" {
		t.Errorf("StringWithV() = %q, want %q", got, "v1.2.3")
	}
}

func TestBumpMajor(t *testing.T) {
	tests := []struct {
		from Version
		want Version
	}{
		{Version{1, 2, 3, ""}, Version{2, 0, 0, ""}},
		{Version{0, 1, 0, ""}, Version{1, 0, 0, ""}},
		{Version{1, 0, 0, "rc.1"}, Version{2, 0, 0, ""}},
	}
	for _, tt := range tests {
		if got := BumpMajor(tt.from); got != tt.want {
			t.Errorf("BumpMajor(%v) = %v, want %v", tt.from, got, tt.want)
		}
	}
}

func TestBumpMinor(t *testing.T) {
	got := BumpMinor(Version{1, 2, 3, ""})
	want := Version{1, 3, 0, ""}
	if got != want {
		t.Errorf("BumpMinor() = %v, want %v", got, want)
	}
}

func TestBumpPatch(t *testing.T) {
	got := BumpPatch(Version{1, 2, 3, ""})
	want := Version{1, 2, 4, ""}
	if got != want {
		t.Errorf("BumpPatch() = %v, want %v", got, want)
	}
}

func TestBumpClearsPreRelease(t *testing.T) {
	v := Version{1, 0, 0, "rc.1"}
	if got := BumpPatch(v); got.PreRelease != "" {
		t.Errorf("BumpPatch should clear pre-release, got %v", got)
	}
}

func TestResolve(t *testing.T) {
	current := Version{1, 2, 3, ""}

	tests := []struct {
		arg  string
		want Version
	}{
		{"major", Version{2, 0, 0, ""}},
		{"minor", Version{1, 3, 0, ""}},
		{"patch", Version{1, 2, 4, ""}},
		{"5.0.0", Version{5, 0, 0, ""}},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			got, err := Resolve(current, tt.arg)
			if err != nil {
				t.Errorf("Resolve(%q) unexpected error: %v", tt.arg, err)
				return
			}
			if got != tt.want {
				t.Errorf("Resolve(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}
}

func TestResolveInvalid(t *testing.T) {
	_, err := Resolve(Version{}, "not-a-version")
	if err == nil {
		t.Error("Resolve with invalid arg should return error")
	}
}

func TestIsMajorBump(t *testing.T) {
	if !IsMajorBump(Version{1, 0, 0, ""}, Version{2, 0, 0, ""}) {
		t.Error("1.x -> 2.x should be major bump")
	}
	if IsMajorBump(Version{1, 2, 0, ""}, Version{1, 3, 0, ""}) {
		t.Error("1.2 -> 1.3 should not be major bump")
	}
}

func TestIsDowngrade(t *testing.T) {
	if !IsDowngrade(Version{2, 0, 0, ""}, Version{1, 0, 0, ""}) {
		t.Error("2.0.0 -> 1.0.0 should be downgrade")
	}
	if IsDowngrade(Version{1, 0, 0, ""}, Version{2, 0, 0, ""}) {
		t.Error("1.0.0 -> 2.0.0 should not be downgrade")
	}
	if !IsDowngrade(Version{1, 3, 0, ""}, Version{1, 2, 0, ""}) {
		t.Error("1.3.0 -> 1.2.0 should be downgrade")
	}
	if !IsDowngrade(Version{1, 2, 4, ""}, Version{1, 2, 3, ""}) {
		t.Error("1.2.4 -> 1.2.3 should be downgrade")
	}
	// Pre-release awareness: 1.0.0 -> 1.0.0-rc.1 is a downgrade
	if !IsDowngrade(Version{1, 0, 0, ""}, Version{1, 0, 0, "rc.1"}) {
		t.Error("1.0.0 -> 1.0.0-rc.1 should be downgrade")
	}
	// 1.0.0-rc.1 -> 1.0.0 is NOT a downgrade (it's a promotion)
	if IsDowngrade(Version{1, 0, 0, "rc.1"}, Version{1, 0, 0, ""}) {
		t.Error("1.0.0-rc.1 -> 1.0.0 should not be downgrade")
	}
}

func TestIsPreRelease(t *testing.T) {
	if !IsPreRelease(Version{1, 0, 0, "rc.1"}) {
		t.Error("1.0.0-rc.1 should be pre-release")
	}
	if IsPreRelease(Version{1, 0, 0, ""}) {
		t.Error("1.0.0 should not be pre-release")
	}
}
