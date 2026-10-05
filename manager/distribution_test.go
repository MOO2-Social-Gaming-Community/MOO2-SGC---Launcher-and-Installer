package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPayloadImportRejectsUnknownAndPreservesCache(t *testing.T) {
	m := distributionTestManager(t)
	source := filepath.Join(t.TempDir(), "unknown.zip")
	os.WriteFile(source, []byte("not original"), 0600)
	cache := filepath.Join(m.Data, "cache", "base.zip")
	os.WriteFile(cache, []byte("keep"), 0600)
	if _, e := m.importPayload("base", source); e == nil {
		t.Fatal("unknown payload accepted")
	}
	b, _ := os.ReadFile(cache)
	if string(b) != "keep" {
		t.Fatal("prior cache changed")
	}
}
func TestPayloadImportPathAndKind(t *testing.T) {
	m := distributionTestManager(t)
	for _, c := range [][2]string{{"other", "/tmp/a"}, {"base", "relative.zip"}} {
		if _, e := m.importPayload(c[0], c[1]); e == nil {
			t.Fatal("invalid import accepted")
		}
	}
}
func TestDistributionDoesNotRequirePRSL(t *testing.T) {
	m := distributionTestManager(t)
	s := m.distributionStatus()
	if s["live_prsl"] != false {
		t.Fatal("PRSL coupled")
	}
	if s["sourceforge"] != "deferred" {
		t.Fatal("SourceForge unexpectedly active")
	}
}

func distributionTestManager(t *testing.T) *Manager {
	t.Helper()
	m, e := newManager(t.TempDir(), t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	return m
}
