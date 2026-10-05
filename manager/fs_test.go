package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func zipFixture(t *testing.T, names []string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "fixture.zip")
	f, e := os.Create(p)
	if e != nil {
		t.Fatal(e)
	}
	w := zip.NewWriter(f)
	for _, n := range names {
		b, e := w.Create(n)
		if e != nil {
			t.Fatal(e)
		}
		b.Write([]byte("fixture"))
	}
	w.Close()
	f.Close()
	return p
}
func TestSafePaths(t *testing.T) {
	for _, n := range []string{"../escape", "/absolute", "C:/path", "folder\\file", "a//b", "a/./b", "a/../b", "CON", "aux.cfg", "foo.", "foo ", "x\nfile", "x\tfile", "a:stream", "a?b"} {
		t.Run(strings.ReplaceAll(n, "/", "_"), func(t *testing.T) {
			if _, e := safeRel(n); e == nil {
				t.Fatal(n)
			}
		})
	}
}
func TestGoodPath(t *testing.T) {
	if _, e := safeRel("150/mods/ice/ICE-X.CFG"); e != nil {
		t.Fatal(e)
	}
}
func TestZipTraversalPreflight(t *testing.T) {
	z := zipFixture(t, []string{"good", "../bad"})
	out := filepath.Join(t.TempDir(), "out")
	if e := extractZip(z, out, "", false); e == nil {
		t.Fatal("accepted")
	}
	if _, e := os.Stat(filepath.Join(out, "good")); !os.IsNotExist(e) {
		t.Fatal("partial write before validation")
	}
}
func TestZipCaseCollision(t *testing.T) {
	z := zipFixture(t, []string{"foo", "FOO"})
	if e := extractZip(z, filepath.Join(t.TempDir(), "out"), "", false); e == nil {
		t.Fatal("accepted")
	}
}
func TestZipExistingOverlayRefused(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(filepath.Join(d, "foo"), []byte("original"), 0600)
	z := zipFixture(t, []string{"foo"})
	if e := extractZip(z, d, "", false); e == nil {
		t.Fatal("accepted")
	}
	b, _ := os.ReadFile(filepath.Join(d, "foo"))
	if string(b) != "original" {
		t.Fatal("modified")
	}
}
func TestZipSymlinkRefused(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.zip")
	f, _ := os.Create(p)
	w := zip.NewWriter(f)
	h := &zip.FileHeader{Name: "link"}
	h.SetMode(os.ModeSymlink | 0777)
	b, _ := w.CreateHeader(h)
	b.Write([]byte("/tmp/target"))
	w.Close()
	f.Close()
	if e := extractZip(p, filepath.Join(t.TempDir(), "out"), "", false); e == nil {
		t.Fatal("symlink accepted")
	}
}
func TestFreshInstallOmitsArchivedSaves(t *testing.T) {
	z := zipFixture(t, []string{"Root/ORION2.EXE", "Root/SAVE1.GAM", "Root/lastrace.rac", "Root/HOF.M2", "Root/mox.set"})
	d := t.TempDir()
	if e := extractZip(z, d, "Root/", true); e != nil {
		t.Fatal(e)
	}
	es, _ := os.ReadDir(d)
	if len(es) != 1 || es[0].Name() != "ORION2.EXE" {
		t.Fatal(es)
	}
}
func TestBroadCFGIsNotMutable(t *testing.T) {
	for _, n := range []string{"ORION2.CFG", "150/ENABLE.CFG", "150/mods/ice/ICE-M.CFG", "any.lua", "ORION150.EXE"} {
		if isUserFile(n) {
			t.Fatal(n)
		}
	}
	for _, n := range []string{"SAVE1.GAM", "150/USER.CFG", "150/build/BUILD0.CFG"} {
		if !isUserFile(n) {
			t.Fatal(n)
		}
	}
}
func TestAtomicReplacement(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x")
	if e := atomicWrite(p, []byte("a")); e != nil {
		t.Fatal(e)
	}
	if e := atomicWrite(p, []byte("b")); e != nil {
		t.Fatal(e)
	}
	b, _ := os.ReadFile(p)
	if string(b) != "b" {
		t.Fatal(string(b))
	}
}
func TestAtomicJSONRejectsTrailingData(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.json")
	os.WriteFile(p, []byte(`{"generation":"g1"}{"generation":"g2"}`), 0600)
	var a Active
	if e := readJSON(p, &a); e == nil {
		t.Fatal("trailing data accepted")
	}
}
func TestDownloadValidatesBeforeReplacement(t *testing.T) {
	data := []byte("new package")
	sum := sha256.Sum256(data)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write(data) }))
	defer srv.Close()
	p := filepath.Join(t.TempDir(), "package.zip")
	os.WriteFile(p, []byte("old package"), 0600)
	if e := downloadVerifiedWith(srv.Client(), srv.URL, p, strings.Repeat("0", 64), 1024); e == nil {
		t.Fatal("bad hash accepted")
	}
	b, _ := os.ReadFile(p)
	if string(b) != "old package" {
		t.Fatal("cache damaged")
	}
	if e := downloadVerifiedWith(srv.Client(), srv.URL, p, hex.EncodeToString(sum[:]), 1024); e != nil {
		t.Fatal(e)
	}
	b, _ = os.ReadFile(p)
	if string(b) != string(data) {
		t.Fatal("not activated")
	}
}
func TestDownloadBounds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(strings.Repeat("x", 100))) }))
	defer srv.Close()
	if e := downloadVerifiedWith(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "pkg"), "", 10); e == nil {
		t.Fatal("accepted oversized download")
	}
}
func TestDownloadNetworkFailurePreservesCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer srv.Close()
	p := filepath.Join(t.TempDir(), "pkg")
	os.WriteFile(p, []byte("good"), 0600)
	if e := downloadVerifiedWith(srv.Client(), srv.URL, p, "", 100); e == nil {
		t.Fatal("accepted HTTP500")
	}
	b, _ := os.ReadFile(p)
	if string(b) != "good" {
		t.Fatal("cache damaged")
	}
}
func TestDownloadHostRestriction(t *testing.T) {
	if e := downloadVerified("http://127.0.0.1:1/a", "/tmp/unused", "", 10); e == nil {
		t.Fatal("unapproved URL")
	}
}
