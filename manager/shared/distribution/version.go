package distribution

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var releaseVersionRE = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// VersionAtLeast compares the project's strict numeric three-part release IDs.
// A signed catalog can still be too old for the running bootstrapper.
func VersionAtLeast(have, want string) bool {
	if !releaseVersionRE.MatchString(have) || !releaseVersionRE.MatchString(want) {
		return false
	}
	h, w := strings.Split(have, "."), strings.Split(want, ".")
	var a, b [3]uint64
	for i := 0; i < 3; i++ {
		var e error
		a[i], e = strconv.ParseUint(h[i], 10, 32)
		if e != nil {
			return false
		}
		b[i], e = strconv.ParseUint(w[i], 10, 32)
		if e != nil {
			return false
		}
	}
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return true
}

// RequireLauncher supplements signature/freshness checks with a local minimum.
// Blank minimum retains the old generic verifier contract for non-launcher users.
func (v VerifiedManifest) RequireLauncher(platform, minimum string) error {
	if minimum == "" {
		return nil
	}
	if !releaseVersionRE.MatchString(minimum) {
		return fmt.Errorf("invalid local launcher minimum %q", minimum)
	}
	p, e := v.Select("launcher", platform)
	if e != nil {
		return e
	}
	if p.Version != v.Manifest.Release {
		return fmt.Errorf("signed release %s and selected launcher %s disagree", v.Manifest.Release, p.Version)
	}
	if !VersionAtLeast(p.Version, minimum) {
		return fmt.Errorf("stale signed launcher %s; this executable requires %s or newer (GitHub latest may still point at an older release)", p.Version, minimum)
	}
	return nil
}
