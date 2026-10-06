package distribution

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestVersionFloor(t *testing.T) {
	for _, c := range []struct {
		h, w string
		ok   bool
	}{{"0.4.2", "0.4.6", false}, {"0.4.6", "0.4.6", true}, {"0.4.10", "0.4.6", true}, {"1.0.0", "0.4.6", true}, {"00.4.6", "0.4.6", false}, {"0.4.6-alpha", "0.4.6", false}, {"0.4.6", "", false}, {"99999999999999999999999.1.1", "0.4.6", false}} {
		if VersionAtLeast(c.h, c.w) != c.ok {
			t.Fatal(c)
		}
	}
}
func fixtureVersion(t *testing.T, version string, revision uint64) *fixture {
	f := newFixture(t)
	f.pkg.Version = version
	f.pkg.Signature = Sign(PackageMessage(f.pkg), "test-key", f.key)
	f.manifest.Release = version
	f.manifest.Revision = revision
	f.manifest.Packages = []Package{f.pkg}
	return f
}
func TestStaleLatestFallsThroughToPinnedRelease(t *testing.T) {
	f := fixtureVersion(t, "0.4.6", 46)
	good, sig := f.signed(t)
	f.pkg.Version = "0.4.2"
	f.pkg.Signature = Sign(PackageMessage(f.pkg), "test-key", f.key)
	f.manifest.Release = "0.4.2"
	f.manifest.Revision = 42
	f.manifest.Packages = []Package{f.pkg}
	old, oldsig := f.signed(t)
	var requested []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = append(requested, r.URL.Path)
		switch r.URL.Path {
		case "/latest/manifest.json":
			w.Write(old)
		case "/latest/manifest.sig":
			w.Write(oldsig)
		case "/pinned/manifest.json":
			w.Write(good)
		case "/pinned/manifest.sig":
			w.Write(sig)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	c := NewClient(f.trust)
	c.fixture = true
	c.HTTP = server.Client()
	// Keep Trust validation strict even while mapping the fixture URL in transport.
	c.Trust.Endpoints = []Endpoint{{Mirror: Mirror{"github", 10, "https://release.test/latest/manifest.json"}, SignatureURL: "https://release.test/latest/manifest.sig"}, {Mirror: Mirror{"github", 20, "https://release.test/pinned/manifest.json"}, SignatureURL: "https://release.test/pinned/manifest.sig"}}
	c.HTTP.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		r.URL.Scheme = "http"
		r.URL.Host = strings.TrimPrefix(server.URL, "http://")
		return http.DefaultTransport.RoundTrip(r)
	})
	c.MinimumLauncherVersion = "0.4.6"
	c.LauncherPlatform = "linux-amd64"
	v, e := c.FetchManifest(context.Background(), TrustedState{}, time.Now())
	if e != nil || v.Manifest.Release != "0.4.6" || len(requested) != 4 {
		t.Fatal(v.Manifest.Release, e, requested)
	}
	// Both endpoints now too old for a newer running setup; do not return anything.
	c.MinimumLauncherVersion = "0.4.7"
	_, e = c.FetchManifest(context.Background(), TrustedState{}, time.Now())
	if e == nil || !strings.Contains(e.Error(), "stale signed launcher") {
		t.Fatal(e)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestOfflineAndMismatchedReleaseFloor(t *testing.T) {
	f := fixtureVersion(t, "0.4.2", 42)
	v := f.verified(t)
	if v.RequireLauncher("linux-amd64", "0.4.6") == nil {
		t.Fatal("stale offline release accepted")
	}
	v.Manifest.Release = "0.4.6"
	if v.RequireLauncher("linux-amd64", "0.4.2") == nil {
		t.Fatal("release/package mismatch accepted")
	}
}
