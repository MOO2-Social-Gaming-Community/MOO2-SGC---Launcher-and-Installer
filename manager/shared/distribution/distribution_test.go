package distribution

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fixture struct {
	trust    Trust
	key      ed25519.PrivateKey
	pkg      Package
	manifest Manifest
	blob     []byte
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	pub, key, e := ed25519.GenerateKey(rand.Reader)
	if e != nil {
		t.Fatal(e)
	}
	f := &fixture{trust: Trust{Schema: 1, Feed: "moo2-sgc-test", Channel: "alpha", Development: true, Keys: map[string]string{"test-key": hex.EncodeToString(pub)}, AllowedHosts: []string{"127.0.0.1", "release.test"}}, key: key}
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	w, e := z.Create("moo2-launcher")
	if e != nil {
		t.Fatal(e)
	}
	payload := []byte("unit fixture; not executable\n")
	w.Write(payload)
	z.Close()
	f.blob = b.Bytes()
	h := sha256.Sum256(f.blob)
	fh := sha256.Sum256(payload)
	f.pkg = Package{ID: "launcher", Kind: "launcher", Version: "0.4.0-alpha.1", Platform: "linux-amd64", Filename: "launcher-test.zip", Format: "zip", Entrypoint: "moo2-launcher", Size: int64(len(f.blob)), SHA256: hex.EncodeToString(h[:]), Redistribution: "project-owned", Mirrors: []Mirror{}, Files: map[string]FileRecord{"moo2-launcher": {hex.EncodeToString(fh[:]), int64(len(payload))}}}
	f.pkg.Signature = Sign(PackageMessage(f.pkg), "test-key", key)
	f.manifest = Manifest{Schema: 1, Feed: f.trust.Feed, Channel: f.trust.Channel, Revision: 2, Release: f.pkg.Version, Published: time.Now().UTC().Add(-time.Minute), Expires: time.Now().UTC().Add(time.Hour), Packages: []Package{f.pkg}}
	return f
}
func (f *fixture) signed(t *testing.T) ([]byte, []byte) {
	t.Helper()
	b, e := json.Marshal(f.manifest)
	if e != nil {
		t.Fatal(e)
	}
	s, e := json.Marshal(SignManifest(b, "test-key", f.key))
	if e != nil {
		t.Fatal(e)
	}
	return b, s
}
func (f *fixture) verified(t *testing.T) VerifiedManifest {
	t.Helper()
	b, s := f.signed(t)
	v, e := f.trust.VerifyManifest(b, s, time.Now(), TrustedState{}, true)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func (f *fixture) archive(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "launcher.zip")
	if e := os.WriteFile(p, f.blob, 0600); e != nil {
		t.Fatal(e)
	}
	return p
}
func checkError(t *testing.T, e error) {
	t.Helper()
	if e == nil {
		t.Fatal("expected refusal")
	}
}
func TestManifestValidation(t *testing.T) {
	cases := []struct {
		name   string
		change func(*fixture)
	}{
		{"future-schema", func(f *fixture) { f.manifest.Schema = 2 }},
		{"wrong-feed", func(f *fixture) { f.manifest.Feed = "other" }},
		{"wrong-channel", func(f *fixture) { f.manifest.Channel = "stable" }},
		{"expired", func(f *fixture) { f.manifest.Expires = time.Now().Add(-time.Second) }},
		{"future-publication", func(f *fixture) {
			f.manifest.Published = time.Now().Add(time.Hour)
			f.manifest.Expires = time.Now().Add(2 * time.Hour)
		}},
		{"excessive-lifetime", func(f *fixture) { f.manifest.Expires = time.Now().Add(94 * 24 * time.Hour) }},
		{"zero-revision", func(f *fixture) { f.manifest.Revision = 0 }},
		{"duplicate-package", func(f *fixture) { f.manifest.Packages = append(f.manifest.Packages, f.pkg) }},
		{"unsigned-package", func(f *fixture) { f.manifest.Packages[0].Signature.Value = "bad" }},
		{"unsupported-kind", func(f *fixture) { f.manifest.Packages[0].Kind = "commercial-game" }},
		{"unsupported-platform", func(f *fixture) { f.manifest.Packages[0].Platform = "unknown" }},
		{"unrecorded-rights", func(f *fixture) { f.manifest.Packages[0].Redistribution = "pending" }},
		{"missing-file-index", func(f *fixture) { f.manifest.Packages[0].Files = nil }},
		{"file-traversal", func(f *fixture) { f.manifest.Packages[0].Entrypoint = "../bad" }},
		{"untrusted-package-host", func(f *fixture) {
			f.manifest.Packages[0].Mirrors = []Mirror{{"github", 10, "https://attacker.test/a.zip"}}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t)
			c.change(f)
			b, s := f.signed(t)
			_, e := f.trust.VerifyManifest(b, s, time.Now(), TrustedState{}, true)
			checkError(t, e)
		})
	}
}
func TestManifestAuthenticityAndRollback(t *testing.T) {
	f := newFixture(t)
	b, s := f.signed(t)
	v, e := f.trust.VerifyManifest(b, s, time.Now(), TrustedState{}, true)
	if e != nil {
		t.Fatal(e)
	}
	t.Run("tampered-bytes", func(t *testing.T) {
		_, e := f.trust.VerifyManifest(append(b, ' '), s, time.Now(), TrustedState{}, true)
		checkError(t, e)
	})
	t.Run("unknown-key", func(t *testing.T) {
		var x Signature
		json.Unmarshal(s, &x)
		x.KeyID = "unknown"
		bs, _ := json.Marshal(x)
		_, e := f.trust.VerifyManifest(b, bs, time.Now(), TrustedState{}, true)
		checkError(t, e)
	})
	t.Run("older-revision", func(t *testing.T) {
		p := v.State()
		p.Revision++
		_, e := f.trust.VerifyManifest(b, s, time.Now(), p, true)
		checkError(t, e)
	})
	t.Run("same-revision-different-digest", func(t *testing.T) {
		p := v.State()
		p.Digest = strings.Repeat("a", 64)
		_, e := f.trust.VerifyManifest(b, s, time.Now(), p, true)
		checkError(t, e)
	})
	t.Run("identical-revision-retry", func(t *testing.T) {
		_, e := f.trust.VerifyManifest(b, s, time.Now(), v.State(), true)
		if e != nil {
			t.Fatal(e)
		}
	})
	t.Run("wrong-state-feed", func(t *testing.T) {
		p := v.State()
		p.Feed = "other"
		_, e := f.trust.VerifyManifest(b, s, time.Now(), p, true)
		checkError(t, e)
	})
	t.Run("signed-duplicate-JSON-key", func(t *testing.T) {
		dup := []byte(strings.Replace(string(b), `"schema":1`, `"schema":1,"schema":1`, 1))
		sig, _ := json.Marshal(SignManifest(dup, "test-key", f.key))
		_, e := f.trust.VerifyManifest(dup, sig, time.Now(), TrustedState{}, true)
		checkError(t, e)
	})
}
func TestStrictJSONAndTrust(t *testing.T) {
	for _, s := range []string{`{"schema":1,"schema":1}`, `{"schema":1} {}`, `{"unexpected":1}`, `{`, strings.Repeat(" ", MaxMetadata+1)} {
		var v struct {
			Schema int `json:"schema"`
		}
		checkError(t, DecodeStrict([]byte(s), &v))
	}
	f := newFixture(t)
	for _, u := range []string{"http://release.test/a", "https://evil.test/a", "https://user:pass@release.test/a", "https://release.test/a#fragment", "https://release.test:8080/a"} {
		checkError(t, f.trust.CheckURL(u))
	}
	if e := f.trust.CheckURL("https://release.test/a"); e != nil {
		t.Fatal(e)
	}
	f.trust.Keys = nil
	checkError(t, f.trust.Validate())
}
func TestPackageSignatureBindsFileIndex(t *testing.T) {
	f := newFixture(t)
	f.pkg.Files["moo2-launcher"] = FileRecord{strings.Repeat("0", 64), 12}
	checkError(t, f.trust.VerifyPackage(f.pkg))
}
func tlsClient(f *fixture, s *httptest.Server) *Client {
	c := NewClient(f.trust)
	original := c.HTTP.CheckRedirect
	c.HTTP = s.Client()
	c.HTTP.Timeout = 2 * time.Second
	c.HTTP.CheckRedirect = original
	return c
}
func TestDownloadMirrorFailover(t *testing.T) {
	f := newFixture(t)
	calls := map[string]int{}
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls[r.URL.Path]++
		if r.URL.Path == "/primary" {
			http.Error(w, "offline", 503)
			return
		}
		w.Write(f.blob)
	}))
	defer s.Close()
	f.pkg.Mirrors = []Mirror{{"cloudflare-r2", 20, s.URL + "/secondary"}, {"github", 10, s.URL + "/primary"}}
	c := tlsClient(f, s)
	path, e := c.Download(context.Background(), f.pkg, t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	if e = VerifyFile(path, f.pkg); e != nil {
		t.Fatal(e)
	}
	if calls["/primary"] != 2 || calls["/secondary"] != 1 {
		t.Fatal(calls)
	}
	before := calls["/secondary"]
	if _, e = c.Download(context.Background(), f.pkg, filepath.Dir(path)); e != nil {
		t.Fatal(e)
	}
	if calls["/secondary"] != before {
		t.Fatal("cache redownloaded")
	}
}
func TestDownloadResume(t *testing.T) {
	for _, mode := range []string{"range-supported", "range-ignored", "range-invalid", "corrupt-partial"} {
		t.Run(mode, func(t *testing.T) {
			f := newFixture(t)
			cache := t.TempDir()
			part := filepath.Join(cache, f.pkg.SHA256+".part")
			prefix := append([]byte{}, f.blob[:20]...)
			if mode == "corrupt-partial" {
				prefix[0] ^= 1
			}
			os.WriteFile(part, prefix, 0600)
			var ranged bool
			s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ranged = ranged || r.Header.Get("Range") == "bytes=20-"
				if mode == "range-ignored" || r.Header.Get("Range") == "" {
					w.Write(f.blob)
					return
				}
				start := 20
				if mode == "range-invalid" {
					start = 19
				}
				w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, len(f.blob)-1, len(f.blob)))
				w.WriteHeader(206)
				w.Write(f.blob[20:])
			}))
			defer s.Close()
			f.pkg.Mirrors = []Mirror{{"github", 10, s.URL + "/file"}}
			p, e := tlsClient(f, s).Download(context.Background(), f.pkg, cache)
			if mode == "range-invalid" || mode == "corrupt-partial" {
				checkError(t, e)
			} else {
				if e != nil {
					t.Fatal(e)
				}
				if e = VerifyFile(p, f.pkg); e != nil {
					t.Fatal(e)
				}
			}
			if !ranged {
				t.Fatal("missing Range request")
			}
		})
	}
}
func TestInterruptedDownloadResumesWithinSession(t *testing.T) {
	f := newFixture(t)
	calls := 0
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Content-Length", fmt.Sprint(len(f.blob)))
			w.Write(f.blob[:25])
			return
		}
		if r.Header.Get("Range") != "bytes=25-" {
			t.Error("partial not resumed")
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes 25-%d/%d", len(f.blob)-1, len(f.blob)))
		w.WriteHeader(206)
		w.Write(f.blob[25:])
	}))
	defer s.Close()
	f.pkg.Mirrors = []Mirror{{"github", 10, s.URL + "/file"}}
	_, e := tlsClient(f, s).Download(context.Background(), f.pkg, t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	if calls != 2 {
		t.Fatal(calls)
	}
}
func TestCorruptPrimaryDoesNotPoisonMirror(t *testing.T) {
	f := newFixture(t)
	bad := bytes.Repeat([]byte{0}, len(f.blob))
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bad" {
			w.Write(bad)
		} else {
			if r.Header.Get("Range") != "" {
				t.Error("corrupt partial carried to mirror")
			}
			w.Write(f.blob)
		}
	}))
	defer s.Close()
	f.pkg.Mirrors = []Mirror{{"github", 10, s.URL + "/bad"}, {"cloudflare-r2", 20, s.URL + "/good"}}
	_, e := tlsClient(f, s).Download(context.Background(), f.pkg, t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
}
func TestDownloadSizeAndSignatureRefusal(t *testing.T) {
	f := newFixture(t)
	calls := 0
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.Write(append(f.blob, 'x')) }))
	defer s.Close()
	f.pkg.Mirrors = []Mirror{{"github", 10, s.URL + "/bad"}}
	c := tlsClient(f, s)
	_, e := c.Download(context.Background(), f.pkg, t.TempDir())
	checkError(t, e)
	f.pkg.Signature.Value = "bad"
	calls = 0
	_, e = c.Download(context.Background(), f.pkg, t.TempDir())
	checkError(t, e)
	if calls != 0 {
		t.Fatal("unsigned package caused request")
	}
}
func TestHTTPSDowngradeRefusal(t *testing.T) {
	f := newFixture(t)
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "http://release.test/payload", 302) }))
	defer s.Close()
	f.pkg.Mirrors = []Mirror{{"github", 10, s.URL + "/redirect"}}
	_, e := tlsClient(f, s).Download(context.Background(), f.pkg, t.TempDir())
	checkError(t, e)
}
func TestManifestMirrorFailoverAndAntiRollback(t *testing.T) {
	f := newFixture(t)
	b, sig := f.signed(t)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/primary", "/primary.sig":
			http.Error(w, "offline", 503)
		case "/mirror":
			w.Write(b)
		default:
			w.Write(sig)
		}
	}))
	defer server.Close()
	f.trust.Endpoints = []Endpoint{{Mirror{"github", 10, server.URL + "/primary"}, server.URL + "/primary.sig"}, {Mirror{"cloudflare-r2", 20, server.URL + "/mirror"}, server.URL + "/mirror.sig"}}
	c := tlsClient(f, server)
	v, e := c.FetchManifest(context.Background(), TrustedState{}, time.Now())
	if e != nil {
		t.Fatal(e)
	}
	prior := v.State()
	prior.Revision++
	_, e = c.FetchManifest(context.Background(), prior, time.Now())
	checkError(t, e)
}
func TestInstallRepairRollbackAndSignedReceipt(t *testing.T) {
	f := newFixture(t)
	root := t.TempDir()
	i := Installer{Root: root, Trust: f.trust, Health: func(context.Context, string, string) error { return nil }}
	v := f.verified(t)
	archive := f.archive(t)
	first, e := i.Install(context.Background(), v, f.pkg, archive)
	if e != nil {
		t.Fatal(e)
	}
	p1, e := i.Current()
	if e != nil {
		t.Fatal(e)
	}
	second, e := i.Install(context.Background(), v, f.pkg, archive)
	if e != nil {
		t.Fatal(e)
	}
	if first == second {
		t.Fatal("live file overwritten")
	}
	p2, _ := i.Current()
	if p2.Previous != p1.Current {
		t.Fatal("rollback missing")
	}
	back, e := i.Rollback()
	if e != nil || back != first {
		t.Fatal(e, back)
	}
	// Tampering with BOTH the file and an unsigned receipt hash must still fail.
	os.WriteFile(first, []byte("tampered"), 0700)
	receipt := filepath.Join(i.base(), "versions", p1.Current, "receipt.json")
	var r Receipt
	ReadJSON(receipt, &r)
	h, n, _ := HashFile(first)
	r.Files[f.pkg.Entrypoint] = FileRecord{h, n}
	AtomicJSON(receipt, r)
	_, _, e = i.VerifyGeneration(p1.Current)
	checkError(t, e)
	third, e := i.Install(context.Background(), v, f.pkg, archive)
	if e != nil {
		t.Fatal(e)
	}
	if third == first {
		t.Fatal("repair in place")
	}
}
func TestFailedHealthLeavesPointerAndUserData(t *testing.T) {
	f := newFixture(t)
	root := t.TempDir()
	i := Installer{Root: root, Trust: f.trust, Health: func(context.Context, string, string) error { return nil }}
	v := f.verified(t)
	a := f.archive(t)
	if _, e := i.Install(context.Background(), v, f.pkg, a); e != nil {
		t.Fatal(e)
	}
	old, _ := os.ReadFile(filepath.Join(i.base(), "current.json"))
	user := filepath.Join(root, "userdata", "save.dat")
	AtomicWrite(user, []byte("keep"))
	i.Health = func(context.Context, string, string) error { return errors.New("deliberate failure") }
	_, e := i.Install(context.Background(), v, f.pkg, a)
	checkError(t, e)
	now, _ := os.ReadFile(filepath.Join(i.base(), "current.json"))
	if !bytes.Equal(old, now) {
		t.Fatal("activation changed")
	}
	b, _ := os.ReadFile(user)
	if string(b) != "keep" {
		t.Fatal("user data changed")
	}
}
func TestInstallerRequiresHealthAndSignedSelection(t *testing.T) {
	f := newFixture(t)
	i := Installer{Root: t.TempDir(), Trust: f.trust}
	_, e := i.Install(context.Background(), f.verified(t), f.pkg, f.archive(t))
	checkError(t, e)
	i.Health = func(context.Context, string, string) error { return nil }
	p := f.pkg
	p.Version = "0.4.1"
	_, e = i.Install(context.Background(), f.verified(t), p, f.archive(t))
	checkError(t, e)
}
func TestExtractUnsafeArchive(t *testing.T) {
	for _, name := range []string{"../escape", "/absolute", "AUX.txt", "a\\b", "x/../x"} {
		t.Run(name, func(t *testing.T) {
			var b bytes.Buffer
			z := zip.NewWriter(&b)
			w, _ := z.Create(name)
			io.WriteString(w, "bad")
			z.Close()
			p := filepath.Join(t.TempDir(), "bad.zip")
			os.WriteFile(p, b.Bytes(), 0600)
			_, e := Extract(p, filepath.Join(t.TempDir(), "stage"))
			checkError(t, e)
		})
	}
}
func TestSymlinksAndLocks(t *testing.T) {
	root := t.TempDir()
	release, e := Acquire(root)
	if e != nil {
		t.Fatal(e)
	}
	_, e = Acquire(root)
	checkError(t, e)
	release()
	r, e := Acquire(root)
	if e != nil {
		t.Fatal(e)
	}
	r()
	target := filepath.Join(t.TempDir(), "target")
	os.WriteFile(target, []byte("x"), 0600)
	link := filepath.Join(root, "link")
	if e = os.Symlink(target, link); e == nil {
		_, _, e = HashFile(link)
		checkError(t, e)
	}
}
func TestTrustedStatePersists(t *testing.T) {
	f := newFixture(t)
	root := t.TempDir()
	v := f.verified(t)
	if e := Accept(root, v); e != nil {
		t.Fatal(e)
	}
	s, e := LoadState(root, f.trust)
	if e != nil || s.Revision != 2 {
		t.Fatal(e)
	}
	AtomicWrite(filepath.Join(root, "distribution", "trusted-state.json"), []byte("bad"))
	_, e = LoadState(root, f.trust)
	checkError(t, e)
}

func TestNetworkDiagnosticsRedactURLTokens(t *testing.T) {
	e := &url.Error{Op: "Get", URL: "https://release.test/payload?token=SECRET", Err: errors.New("connection reset")}
	if strings.Contains(safeFailureMessage(e), "SECRET") {
		t.Fatal("signed URL token leaked")
	}
}
