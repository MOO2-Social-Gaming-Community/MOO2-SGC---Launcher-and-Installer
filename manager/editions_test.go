package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestKnownSourceEditions(t *testing.T) {
	expected := map[string]int{"cd-en-12": 397, "steam-en-140b23": 405, "legacy-en-131": 408, "manual-en-140b23": 408, "cd-en-12-harness": 398, "steam-en-140b23-harness": 406, "legacy-en-131-harness": 409}
	for _, s := range sourceEditions() {
		t.Run(s.ID, func(t *testing.T) {
			if len(s.Files) != expected[s.ID] {
				t.Fatalf("wrong file count %d", len(s.Files))
			}
			seen := map[string]bool{}
			for _, f := range s.Files {
				if seen[f.Name] || f.Name != strings.ToUpper(f.Name) || strings.ContainsAny(f.Name, "/\\") || len(f.SHA256) != 64 {
					t.Fatal("bad compiled record", f.Name)
				}
				seen[f.Name] = true
			}
			if !seen["ORION2.EXE"] || !seen["GAME.LBX"] {
				t.Fatal("missing essential data")
			}
		})
	}
	for h, id := range map[string]string{CDEngineHash: "cd-en-12-harness", BaseEngineHash: "legacy-en-131-harness", BaselineEngineHash: "steam-en-140b23-harness"} {
		s, e := editionByEngine(h)
		if e != nil || s.ID != id {
			t.Fatal(s, e)
		}
	}
	if _, e := editionByEngine(strings.Repeat("0", 64)); e == nil {
		t.Fatal("unrecognized engine accepted")
	}
}
func TestInstallationLineage(t *testing.T) {
	cases := []struct {
		source, target string
		versions       []string
	}{
		{"cd-en-12", "1.2", []string{"1.2"}}, {"cd-en-12", "1.31", []string{"1.2", "1.31"}},
		{"cd-en-12", "1.40b23", []string{"1.2", "1.31", "1.40b23"}}, {"cd-en-12", "1.50.26", []string{"1.2", "1.31", "1.40b23", "1.50.26"}},
		{"steam-en-140b23", "1.40b23", []string{"1.40b23"}}, {"steam-en-140b23", "1.50.26", []string{"1.40b23", "1.50.26"}},
		{"legacy-en-131", "1.50.26", []string{"1.31", "1.40b23", "1.50.26"}},
	}
	for _, c := range cases {
		t.Run(c.source+"-"+c.target, func(t *testing.T) {
			s, _ := editionByID(c.source)
			steps, e := expectedLineage(s, c.target)
			if e != nil {
				t.Fatal(e)
			}
			got := []string{}
			for _, x := range steps {
				got = append(got, x.Version)
			}
			if !reflect.DeepEqual(got, c.versions) {
				t.Fatal(got)
			}
		})
	}
	s, _ := editionByID("steam-en-140b23")
	for _, v := range []string{"1.2", "1.31", "unknown"} {
		if _, e := expectedLineage(s, v); e == nil {
			t.Fatal("downgrade accepted", v)
		}
	}
}
func TestVersionProfilesAndCommands(t *testing.T) {
	for _, v := range []string{"1.2", "1.31", "1.40b23", "1.50.26"} {
		t.Run(v, func(t *testing.T) {
			p := defaultProfiles()[0]
			p.Engine = v
			if v != "1.50.26" {
				p.Core = ""
				p.Mods = nil
			}
			if _, e := resolve(p); e != nil {
				t.Fatal(e)
			}
			cfg, e := dosboxConfig(t.TempDir(), p)
			if e != nil {
				t.Fatal(e)
			}
			exe := "ORION2.EXE"
			if v == "1.50.26" {
				exe = "ORION150.EXE"
			}
			if got := dosboxArguments("moo2.conf", "/game", p); filepath.Base(got[len(got)-1]) != exe {
				t.Fatal("wrong entrypoint")
			}
			if v == "1.2" && strings.Contains(cfg, "/skipintro") {
				t.Fatal("unverified CD switch")
			}
			if v != "1.50.26" {
				p.Mods = []string{"150m"}
				if _, e = resolve(p); e == nil {
					t.Fatal("1.50 mod accepted on earlier engine")
				}
			}
		})
	}
}
func makeSourceZip(t *testing.T, files map[string][]byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "input.zip")
	f, e := os.Create(p)
	if e != nil {
		t.Fatal(e)
	}
	z := zip.NewWriter(f)
	for n, b := range files {
		w, e := z.Create(n)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = w.Write(b); e != nil {
			t.Fatal(e)
		}
	}
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	f.Close()
	return p
}
func TestContentVerifiedZipImport(t *testing.T) {
	records := []BaseFile{testRecord("ORION2.EXE", []byte("DOS")), testRecord("GAME.LBX", []byte("data"))}
	cases := []struct {
		name  string
		files map[string][]byte
		bad   bool
	}{
		{"root and case", map[string][]byte{"anything/Orion2.exe": []byte("DOS"), "anything/game.lbx": []byte("data"), "anything/SAVE1.GAM": []byte("private")}, false},
		{"modified data", map[string][]byte{"ORION2.EXE": []byte("DOS"), "GAME.LBX": []byte("bad!")}, true},
		{"missing data", map[string][]byte{"ORION2.EXE": []byte("DOS")}, true},
		{"traversal", map[string][]byte{"ORION2.EXE": []byte("DOS"), "GAME.LBX": []byte("data"), "../outside": nil}, true},
		{"ambiguous root", map[string][]byte{"ORION2.EXE": []byte("DOS"), "copy/ORION2.EXE": []byte("DOS"), "GAME.LBX": []byte("data")}, true},
		{"case collision", map[string][]byte{"ORION2.EXE": []byte("DOS"), "orion2.exe": []byte("DOS"), "GAME.LBX": []byte("data")}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := makeSourceZip(t, c.files)
			out := filepath.Join(t.TempDir(), "snapshot.zip")
			e := writeZipSnapshot(p, out, records)
			if (e != nil) != c.bad {
				t.Fatal(e)
			}
			if c.bad {
				if _, e = os.Stat(out); !os.IsNotExist(e) {
					t.Fatal("failed import left snapshot")
				}
				return
			}
			z, e := zip.OpenReader(out)
			if e != nil {
				t.Fatal(e)
			}
			defer z.Close()
			if len(z.File) != 2 {
				t.Fatal("unrelated files copied")
			}
		})
	}
}
func TestMutableAudioNotPinnedButCodeStillPinned(t *testing.T) {
	root := t.TempDir()
	data := testRecord("GAME.LBX", []byte("known"))
	sound := testRecord("SOUND.LBX", []byte("sound"))
	records := map[string]FileRecord{}
	for _, r := range []BaseFile{data, sound} {
		b := []byte("known")
		if r.Name == sound.Name {
			b = []byte("sound")
		}
		os.WriteFile(filepath.Join(root, r.Name), b, 0600)
		records[r.Name] = FileRecord{r.SHA256, r.Size, false}
	}
	os.WriteFile(filepath.Join(root, sound.Name), []byte("newaudio"), 0600)
	if e := verifyPinnedBase(root, records, []BaseFile{data, sound}); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(root, data.Name), []byte("other"), 0600)
	changed := testRecord(data.Name, []byte("other"))
	records[data.Name] = FileRecord{changed.SHA256, changed.Size, true}
	if e := verifyPinnedBase(root, records, []BaseFile{data, sound}); e == nil {
		t.Fatal("mutable flag bypassed compiled game identity")
	}
}
func TestBaselineWrongInputRejected(t *testing.T) {
	if _, e := applyBaselineBytes([]byte("wrong"), baselineRecipe()); e == nil {
		t.Fatal("wrong executable patched")
	}
}
func fixtureFile(t *testing.T, name string) string {
	t.Helper()
	root := os.Getenv("MOO2_TEST_ASSETS")
	if root == "" {
		t.Skip("optional owned-file fixture not provided")
	}
	p := filepath.Join(root, name)
	if _, e := os.Stat(p); e != nil {
		t.Fatal(e)
	}
	return p
}
func fixtureMember(t *testing.T, archive, member string) []byte {
	t.Helper()
	z, e := zip.OpenReader(archive)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	for _, f := range z.File {
		if f.Name == member {
			r, e := f.Open()
			if e != nil {
				t.Fatal(e)
			}
			defer r.Close()
			b, e := io.ReadAll(r)
			if e != nil {
				t.Fatal(e)
			}
			return b
		}
	}
	t.Fatal("missing member", member)
	return nil
}
func TestRealBaselineMatchesSteam(t *testing.T) {
	old := fixtureMember(t, fixtureFile(t, "_moo2v131_patch.zip"), "ORION2.EXE")
	steam := fixtureMember(t, fixtureFile(t, "Master of Orion 2 - steam - v1_40b23.zip"), "Master of Orion 2/Orion2.exe")
	out, e := applyBaselineBytes(old, baselineRecipe())
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(out, steam) {
		t.Fatal("historical patch result differs from clean Steam baseline")
	}
	for _, name := range []string{"out-of-bounds", "changed-delta", "wrong-output-hash"} {
		t.Run(name, func(t *testing.T) {
			r := baselineRecipe()
			switch name {
			case "out-of-bounds":
				r.Writes[0].Offset = len(out) + 1
			case "changed-delta":
				r.Writes[0].Hex = "00"
			default:
				r.OutputSHA = strings.Repeat("0", 64)
			}
			if _, e := applyBaselineBytes(old, r); e == nil {
				t.Fatal("invalid patch accepted")
			}
		})
	}
}
func TestRealOfficialDownloadContentAuthentication(t *testing.T) {
	original := fixtureFile(t, "_moo2v131_patch.zip")
	z, e := zip.OpenReader(original)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	var packed bytes.Buffer
	w := zip.NewWriter(&packed)
	for _, f := range z.File {
		r, _ := f.Open()
		b, _ := io.ReadAll(r)
		r.Close()
		out, _ := w.Create("repacked/" + strings.ToLower(f.Name))
		out.Write(b)
	}
	w.Close()
	for _, kind := range []string{"valid-repacked", "wrong-executable", "not-a-zip", "http-error"} {
		t.Run(kind, func(t *testing.T) {
			m, e := newManager(t.TempDir(), t.TempDir())
			if e != nil {
				t.Fatal(e)
			}
			body := packed.Bytes()
			code := 200
			switch kind {
			case "wrong-executable":
				b, e := os.ReadFile(makeSourceZip(t, map[string][]byte{"ORION2.EXE": []byte("wrong")}))
				if e != nil {
					t.Fatal(e)
				}
				body = b
			case "not-a-zip":
				body = []byte("bad")
			case "http-error":
				code = 503
			}
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(code); w.Write(body) }))
			defer server.Close()
			e = m.downloadOfficial131(server.Client(), server.URL)
			if (e == nil) != (kind == "valid-repacked") {
				t.Fatal(e)
			}
			if kind == "valid-repacked" {
				if e = m.ensureOfficial131(); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
}
func TestForgedLineageRejected(t *testing.T) {
	s, _ := editionByID("steam-en-140b23")
	p := defaultProfiles()[0]
	steps, _ := expectedLineage(s, p.Engine)
	mf := Manifest{Schema: 2, Profile: p, SourceBase: strings.Repeat("a", 64), SourceBaseKind: "known-edition-v2:" + s.ID, Lineage: steps}
	mf.Lineage[0].Version = "1.2"
	if e := verifyEditionManifest(mf, t.TempDir()); e == nil {
		t.Fatal("forged lineage accepted")
	}
}
func TestOldImportedReceiptEditionAlias(t *testing.T) {
	s, e := editionByID(baseIndex().Edition)
	if e != nil || s.ID != "legacy-en-131" {
		t.Fatal(s, e)
	}
	b, _ := json.Marshal(ImportedBase{1, "bad", strings.Repeat("a", 64), 408, baseIndex().Edition})
	var r ImportedBase
	if e = json.Unmarshal(b, &r); e != nil {
		t.Fatal(e)
	}
}
