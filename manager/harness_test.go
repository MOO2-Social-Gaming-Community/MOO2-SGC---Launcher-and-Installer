package main

import (
	"bytes"
	"moo2manager/shared/layout"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestHarnessConfigurationContract(t *testing.T) {
	p := defaultProfiles()[0]
	p.Engine = "1.40b23"
	p.Core = ""
	p.Role = "standalone"
	cfg, e := dosboxConfig(t.TempDir(), p)
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range []string{"window_size = 800x600", "pause_when_inactive = off", "mute_when_inactive = off", "output = opengl", "aspect = auto", "viewport = fit", "integer_scaling = off", "shader = sharp", "sbtype = sb16", "irq = 5", "dma = 1", "hdma = 5"} {
		if !strings.Contains(cfg, v) {
			t.Fatal(v)
		}
	}
	for _, bad := range []string{"[autoexec]", "/skipintro", "[cpu]", "1280x960", "\\\n"} {
		if strings.Contains(cfg, bad) {
			t.Fatal("unapproved standalone override", bad)
		}
	}
	if !validHarnessConfig([]byte(cfg)) {
		t.Fatal("standalone differs from harness")
	}
}
func TestHarnessExactArgumentConstruction(t *testing.T) {
	p := defaultProfiles()[0]
	p.Engine = "1.40b23"
	p.Core = ""
	p.Role = "standalone"
	game := filepath.Join(t.TempDir(), "game with spaces")
	cfg := filepath.Join(filepath.Dir(game), "config", "moo2.conf")
	want := []string{"--noprimaryconf", "--conf", cfg, filepath.Join(game, "ORION2.EXE")}
	if got := dosboxArguments(cfg, game, p); !reflect.DeepEqual(got, want) {
		t.Fatal(got)
	}
	p.Fullscreen = true
	want = append(want[:3], "--fullscreen", filepath.Join(game, "ORION2.EXE"))
	if got := dosboxArguments(cfg, game, p); !reflect.DeepEqual(got, want) {
		t.Fatal(got)
	}
	p.Engine = "1.50.26"
	p.Core = "150"
	if got := dosboxArguments(cfg, game, p); filepath.Base(got[len(got)-1]) != "ORION150.EXE" {
		t.Fatal(got)
	}
	p.Role = "host"
	if got := dosboxArguments(cfg, game, p); len(got) != 4 {
		t.Fatal("host must use only network autoexec, not direct program", got)
	}
}
func TestHarnessAudioByteParity(t *testing.T) {
	game := t.TempDir()
	if e := writeDefaultGameConfig(game); e != nil {
		t.Fatal(e)
	}
	for _, n := range []string{"DIG.INI", "MDI.INI", "ORIONCD.INI"} {
		want, _ := resources.ReadFile("assets/harness/" + n)
		got, e := os.ReadFile(filepath.Join(game, n))
		if e != nil || !bytes.Equal(got, want) {
			t.Fatal(n, e)
		}
	}
	b, _ := os.ReadFile(filepath.Join(game, "DIG.INI"))
	if !bytes.Contains(b, []byte("SBLASTER.DIG")) || !bytes.Contains(b, []byte("DMA_16_BIT  -1")) {
		t.Fatal("unverified sound driver")
	}
}
func TestHarnessKernelRejectsUnknownBytes(t *testing.T) {
	for _, b := range [][]byte{nil, make([]byte, 31095), []byte("unknown")} {
		before := append([]byte(nil), b...)
		if _, e := normalizeKernel(b); e == nil {
			t.Fatal("unknown input accepted")
		}
		if !bytes.Equal(b, before) {
			t.Fatal("source mutation")
		}
	}
}
func TestNormalizedPinsContainRuntimeKernel(t *testing.T) {
	for _, id := range []string{"manual-en-140b23", "cd-en-12-harness", "steam-en-140b23-harness", "legacy-en-131-harness"} {
		s, e := editionByID(id)
		if e != nil {
			t.Fatal(e)
		}
		pins := map[string]BaseFile{}
		for _, r := range normalizedBaseFiles(s, "1.40b23") {
			pins[r.Name] = r
		}
		if pins["RKERNEL.COM"].SHA256 != Kernel140Hash || pins["ORION2.EXE"].SHA256 != BaselineEngineHash {
			t.Fatal(id)
		}
		if id != "steam-en-140b23-harness" && pins["ORION131.EXE"].SHA256 != BaseEngineHash {
			t.Fatal("lost official executable", id)
		}
	}
}
func TestRuntimeIndexPinsExecutableDLLsAndResources(t *testing.T) {
	i := harnessRuntimeIndex()
	if i.Version != "0.83.0" || len(i.Files) != 660 {
		t.Fatal(i.Version, len(i.Files))
	}
	found, dll := false, false
	for _, f := range i.Files {
		if f.Name == "dosbox.exe" {
			found = f.SHA256 == HarnessDOSBoxHash
		}
		if strings.HasSuffix(f.Name, ".dll") {
			dll = true
		}
	}
	if !found || !dll {
		t.Fatal("runtime trust misses executable or DLLs")
	}
	if e := verifyHarnessRuntime(t.TempDir()); e == nil {
		t.Fatal("empty runtime accepted")
	}
}
func newPortableForTest(t *testing.T) *Manager {
	t.Helper()
	root := t.TempDir()
	if e := layout.Enable(root); e != nil {
		t.Fatal(e)
	}
	m, e := newManager(root, filepath.Join(root, "userdata"))
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func TestPortableRootDoesNotAutoImportOrOverwriteGame(t *testing.T) {
	m := newPortableForTest(t)
	if !m.Portable {
		t.Fatal("missing mode")
	}
	if exists(filepath.Join(m.Root, "game")) {
		t.Fatal("game created before consent")
	}
	if _, _, e := m.portableActive(); !os.IsNotExist(e) {
		t.Fatal(e)
	}
	p, _ := m.loadProfile("baseline")
	p.Engine = "1.50.26"
	p.Core = "150"
	if _, e := m.build(p); e == nil {
		t.Fatal("overwrote root game with community engine")
	}
}
func TestPortableRecoveryRestoresSnapshot(t *testing.T) {
	m := newPortableForTest(t)
	backup := filepath.Join(m.Root, "backups", "baseline", "b123")
	os.MkdirAll(filepath.Join(backup, "game"), 0700)
	os.MkdirAll(filepath.Join(m.Root, "game"), 0700)
	os.MkdirAll(filepath.Join(m.Root, "state"), 0700)
	os.WriteFile(filepath.Join(backup, "game", "SAVE1.GAM"), []byte("prior save"), 0600)
	os.WriteFile(filepath.Join(m.Root, "game", "SAVE1.GAM"), []byte("partial candidate"), 0600)
	atomicJSON(m.baselineJournal(), baselineTransaction{1, "b123", "g124", true})
	if _, e := m.recoverPortableBaseline(); e != nil {
		t.Fatal(e)
	}
	b, _ := os.ReadFile(filepath.Join(m.Root, "game", "SAVE1.GAM"))
	if string(b) != "prior save" {
		t.Fatal("prior save not restored")
	}
	failed, _ := filepath.Glob(filepath.Join(m.Root, "backups", "baseline", "failed-*", "game", "SAVE1.GAM"))
	if len(failed) != 1 {
		t.Fatal("discarded partial candidate")
	}
	b, _ = os.ReadFile(failed[0])
	if string(b) != "partial candidate" {
		t.Fatal("lost candidate")
	}
	if exists(m.baselineJournal()) {
		t.Fatal("journal not cleared")
	}
	if _, e := m.recoverPortableBaseline(); e != nil {
		t.Fatal("recovery not idempotent", e)
	}
}
func TestPortableRecoveryBeforeInitialRenameDoesNotDeleteOldGame(t *testing.T) {
	m := newPortableForTest(t)
	os.MkdirAll(filepath.Join(m.Root, "backups", "baseline", "b123"), 0700)
	os.MkdirAll(filepath.Join(m.Root, "game"), 0700)
	os.MkdirAll(filepath.Join(m.Root, "state"), 0700)
	os.WriteFile(filepath.Join(m.Root, "game", "SAVE1.GAM"), []byte("untouched"), 0600)
	atomicJSON(m.baselineJournal(), baselineTransaction{1, "b123", "g124", true})
	if _, e := m.recoverPortableBaseline(); e != nil {
		t.Fatal(e)
	}
	b, _ := os.ReadFile(filepath.Join(m.Root, "game", "SAVE1.GAM"))
	if string(b) != "untouched" {
		t.Fatal("old game changed")
	}
}
func TestPortableRecoveryRejectsTraversal(t *testing.T) {
	m := newPortableForTest(t)
	os.MkdirAll(filepath.Join(m.Root, "state"), 0700)
	atomicJSON(m.baselineJournal(), baselineTransaction{1, "../outside", "g1", true})
	if _, e := m.recoverPortableBaseline(); e == nil {
		t.Fatal("invalid journal accepted")
	}
	if !exists(m.baselineJournal()) {
		t.Fatal("bad journal discarded")
	}
}
func TestStandaloneConfigCannotContainNetworkCommands(t *testing.T) {
	p := defaultProfiles()[0]
	p.Role = "standalone"
	cfg, e := dosboxConfig(t.TempDir(), p)
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(cfg, "mount ") || strings.Contains(cfg, "IPXNET") || strings.Contains(cfg, "ORION") {
		t.Fatal("not the tested direct-program invocation")
	}
}

// Opt-in read-only user fixture. The isolated copy is deliberately corrupted;
// the supplied runtime is never modified or executed.
func TestHarnessPrivateRuntimeFixture(t *testing.T) {
	root := os.Getenv("MOO2_HARNESS_FIXTURE")
	if root == "" {
		t.Skip("private runtime fixture not supplied")
	}
	if err := verifyHarnessRuntime(root); err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	if err := walkRegular(root, func(n, src string) error { return copyFile(src, filepath.Join(temp, filepath.FromSlash(n))) }); err != nil {
		t.Fatal(err)
	}
	if err := verifyHarnessRuntime(temp); err != nil {
		t.Fatal(err)
	}
	var dll string
	for _, r := range harnessRuntimeIndex().Files {
		if strings.HasSuffix(strings.ToLower(r.Name), ".dll") {
			dll = filepath.Join(temp, filepath.FromSlash(r.Name))
			break
		}
	}
	b, err := os.ReadFile(dll)
	if err != nil {
		t.Fatal(err)
	}
	corrupted := append([]byte(nil), b...)
	corrupted[len(corrupted)-1] ^= 1
	if err = os.WriteFile(dll, corrupted, 0600); err != nil {
		t.Fatal(err)
	}
	if err = verifyHarnessRuntime(temp); err == nil {
		t.Fatal("altered companion DLL accepted")
	}
	if err = os.WriteFile(dll, b, 0600); err != nil {
		t.Fatal(err)
	}
	extra := filepath.Join(temp, "dosbox.conf")
	if err = os.WriteFile(extra, []byte("[autoexec]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = verifyHarnessRuntime(temp); err == nil {
		t.Fatal("extra runtime config accepted")
	}
	if err = os.Remove(extra); err != nil {
		t.Fatal(err)
	}
	if err = verifyHarnessRuntime(temp); err != nil {
		t.Fatal(err)
	}
	if err = verifyHarnessRuntime(root); err != nil {
		t.Fatal("source runtime changed", err)
	}
}
