package main

import (
	"strings"
	"testing"
)

func validProfile() Profile { return defaultProfiles()[0] }
func TestDefaultProfilesResolve(t *testing.T) {
	for _, p := range defaultProfiles() {
		t.Run(p.ID, func(t *testing.T) {
			if _, e := resolve(p); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestEveryBundledCoreResolves(t *testing.T) {
	for _, m := range catalog() {
		if m.Group != "Core" {
			continue
		}
		t.Run(m.ID, func(t *testing.T) {
			p := validProfile()
			p.Core = m.ID
			r, e := resolve(p)
			if e != nil {
				t.Fatal(e)
			}
			if len(r.Mods) == 0 {
				t.Fatal("no mods")
			}
		})
	}
}
func TestCoreDependenciesAreVisible(t *testing.T) {
	p := validProfile()
	p.Core = "150m"
	r, e := resolve(p)
	if e != nil {
		t.Fatal(e)
	}
	for _, id := range []string{"1_NOREP", "MAP1_150m", "MIRROR_HW"} {
		if !strings.Contains(strings.Join(r.Automatic, ","), id) {
			t.Error(id)
		}
	}
}
func TestExclusiveMapConflict(t *testing.T) {
	p := validProfile()
	p.Core = "150m"
	p.Mods = []string{"MAP5_GM1"}
	if _, e := resolve(p); e == nil {
		t.Fatal("accepted conflicting map")
	}
}
func TestTwoCoreConflict(t *testing.T) {
	p := validProfile()
	p.Mods = []string{"ICE"}
	if _, e := resolve(p); e == nil {
		t.Fatal("accepted two cores")
	}
}
func TestNoPRSLBackdoor(t *testing.T) {
	for _, id := range []string{"prsl", "PRSL", "chat", "Chat"} {
		p := validProfile()
		p.Mods = []string{id}
		if _, e := resolve(p); e == nil {
			t.Fatal(id)
		}
	}
}
func TestUnknownPackageRefused(t *testing.T) {
	p := validProfile()
	p.Mods = []string{"made-up"}
	if _, e := resolve(p); e == nil {
		t.Fatal("accepted")
	}
}
func TestOriginalCannotHaveMods(t *testing.T) {
	p := validProfile()
	p.Engine = "1.31"
	if _, e := resolve(p); e == nil {
		t.Fatal("accepted")
	}
}
func TestUnknownEngineRefused(t *testing.T) {
	p := validProfile()
	p.Engine = "1.50.27"
	if _, e := resolve(p); e == nil {
		t.Fatal("accepted")
	}
}
func TestDisabledExtensionsAbsent(t *testing.T) {
	r, e := resolve(validProfile())
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(strings.ToLower(r.Config), "enable prsl") || strings.Contains(strings.ToLower(r.Config), "enable chat") {
		t.Fatal(r.Config)
	}
}
func TestDeterministicOrder(t *testing.T) {
	p := validProfile()
	p.Mods = []string{"1_NOREP", "RANDOMTECH"}
	a, _ := resolve(p)
	p.Mods = []string{"RANDOMTECH", "1_NOREP", "RANDOMTECH"}
	b, e := resolve(p)
	if e != nil {
		t.Fatal(e)
	}
	if a.Fingerprint != b.Fingerprint || a.Config != b.Config {
		t.Fatal("unstable order")
	}
}
func TestLocalOptionsDoNotChangeRulesFingerprint(t *testing.T) {
	p := validProfile()
	a, _ := resolve(p)
	p.Name = "Other name"
	p.Fullscreen = true
	p.Role = "join"
	p.Host = "127.0.0.1"
	p.Port = 21301
	b, e := resolve(p)
	if e != nil {
		t.Fatal(e)
	}
	if a.Fingerprint != b.Fingerprint {
		t.Fatal("local options affected rules")
	}
}
func TestJoinInjectionRefused(t *testing.T) {
	for _, host := range []string{"127.0.0.1\nEXIT", "example.com & erase", "::1", ""} {
		p := validProfile()
		p.Role = "join"
		p.Host = host
		if _, e := resolve(p); e == nil {
			t.Fatal(host)
		}
	}
}
func TestBadProfileIDs(t *testing.T) {
	for _, id := range []string{"../escape", "foo/bar", "", "A", "..", "bad name"} {
		p := validProfile()
		p.ID = id
		if _, e := resolve(p); e == nil {
			t.Fatal(id)
		}
	}
}
func TestDOSBoxRoles(t *testing.T) {
	for _, role := range []string{"host", "join", "standalone"} {
		t.Run(role, func(t *testing.T) {
			p := validProfile()
			p.Role = role
			p.Host = "127.0.0.1"
			s, e := dosboxConfig(t.TempDir(), p)
			if e != nil {
				t.Fatal(e)
			}
			if !strings.Contains(s, "ORION150.EXE /skipintro") {
				t.Fatal(s)
			}
			if role == "host" && !strings.Contains(s, "IPXNET STARTSERVER 21300") {
				t.Fatal(s)
			}
			if role == "join" && !strings.Contains(s, "IPXNET CONNECT 127.0.0.1 21300") {
				t.Fatal(s)
			}
			if role == "standalone" && strings.Contains(s, "IPXNET ") {
				t.Fatal(s)
			}
		})
	}
}
func TestOriginalLaunchExecutable(t *testing.T) {
	p := defaultProfiles()[2]
	s, e := dosboxConfig(t.TempDir(), p)
	if e != nil || strings.Contains(s, "ORION150.EXE") || !strings.Contains(s, "ORION2.EXE /skipintro") {
		t.Fatalf("%s %v", s, e)
	}
}
func TestConfigPathInjection(t *testing.T) {
	if _, e := dosboxConfig("/tmp/name\"\nexit", validProfile()); e == nil {
		t.Fatal("unsafe path")
	}
}
func TestUpstreamNumericVersions(t *testing.T) {
	v, e := parseUpstreamVersions(`MOO2-1.50.9.zip MOO2-1.50.26.zip MOO2-1.50.26.zip`)
	if e != nil || len(v) != 2 || v[0] != "1.50.26" {
		t.Fatalf("%v %v", v, e)
	}
}
func TestUpstreamFailureDoesNotGuess(t *testing.T) {
	if _, e := parseUpstreamVersions("no release links"); e == nil {
		t.Fatal("guessed version")
	}
}
