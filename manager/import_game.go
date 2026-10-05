package main

// Imports only known DOS game bytes from an owned local folder. Extra programs,
// saves, mods and distribution credentials are deliberately not copied.
import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type BaseFile struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}
type BaseIndex struct {
	Schema  int        `json:"schema"`
	Edition string     `json:"edition"`
	Files   []BaseFile `json:"files"`
}
type ImportedBase struct {
	Schema    int    `json:"schema"`
	Filename  string `json:"filename"`
	SHA256    string `json:"sha256"`
	FileCount int    `json:"file_count"`
	Edition   string `json:"edition"`
}

func baseIndex() BaseIndex {
	b, e := resources.ReadFile("assets/base-files.json")
	if e != nil {
		panic(e)
	}
	var v BaseIndex
	if e = json.Unmarshal(b, &v); e != nil {
		panic(e)
	}
	return v
}
func localPath(s string) (string, error) {
	s = strings.Trim(strings.TrimSpace(s), "\"")
	if s == "" || !filepath.IsAbs(s) {
		return "", errors.New("select an absolute local path")
	}
	p := filepath.Clean(s)
	if e := noSymlinkAncestors(p); e != nil {
		return "", e
	}
	return p, nil
}
func folderFiles(root string) (map[string]string, error) {
	entries, e := os.ReadDir(root)
	if e != nil {
		return nil, e
	}
	if len(entries) > 10000 {
		return nil, errors.New("too many files in selected directory")
	}
	out := map[string]string{}
	for _, x := range entries {
		if x.IsDir() {
			continue
		}
		k := strings.ToUpper(x.Name())
		if _, yes := out[k]; yes {
			return nil, errors.New("case-colliding names in source folder")
		}
		out[k] = filepath.Join(root, x.Name())
	}
	return out, nil
}

// Hash during copying too: the source can change between initial detection and
// import. No active source is replaced unless the complete snapshot matches.
func writeBaseSnapshot(source, archive string, records []BaseFile) (err error) {
	paths, e := folderFiles(source)
	if e != nil {
		return e
	}
	f, e := os.OpenFile(archive, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer func() {
		_ = f.Close()
		if err != nil {
			_ = os.Remove(archive)
		}
	}()
	z := zip.NewWriter(f)
	defer z.Close()
	for _, r := range records {
		p, ok := paths[r.Name]
		if !ok {
			return fmt.Errorf("recognized DOS data missing: %s. Select the game-data folder or import your original recognized ZIP", r.Name)
		}
		if e = regularFile(p); e != nil {
			return e
		}
		i, e := os.Stat(p)
		if e != nil {
			return e
		}
		if i.Size() != r.Size {
			return fmt.Errorf("%s differs from supported DOS 1.31 baseline (size); source untouched", r.Name)
		}
		in, e := os.Open(p)
		if e != nil {
			return e
		}
		h := &zip.FileHeader{Name: "base/" + r.Name, Method: zip.Deflate}
		h.SetMode(0600)
		h.SetModTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
		w, e := z.CreateHeader(h)
		if e != nil {
			in.Close()
			return e
		}
		hash := sha256.New()
		n, ce := io.Copy(io.MultiWriter(w, hash), io.LimitReader(in, r.Size+1))
		in.Close()
		if ce != nil {
			return ce
		}
		if n != r.Size || hex.EncodeToString(hash.Sum(nil)) != r.SHA256 {
			return fmt.Errorf("%s differs from supported DOS 1.31 data; mixed/modified editions are not imported", r.Name)
		}
	}
	if e = z.Close(); e != nil {
		return e
	}
	if e = f.Sync(); e != nil {
		return e
	}
	return f.Close()
}
func (m *Manager) importFolder(source string) (any, error) {
	p, e := localPath(source)
	if e != nil {
		return nil, e
	}
	info, e := os.Stat(p)
	if e != nil {
		return nil, e
	}
	if !info.IsDir() {
		return nil, errors.New("select the folder containing ORION2.EXE and LBX files")
	}
	index := baseIndex()
	cache := filepath.Join(m.Data, "cache")
	tmp, e := os.MkdirTemp(cache, "import-")
	if e != nil {
		return nil, e
	}
	defer os.RemoveAll(tmp)
	staged := filepath.Join(tmp, "base.zip")
	m.progress("Verifying and privately copying recognized DOS game files; source, mods and saved games stay untouched…")
	if e = writeBaseSnapshot(p, staged, index.Files); e != nil {
		return nil, e
	}
	hash, e := hashFile(staged)
	if e != nil {
		return nil, e
	}
	name := "base-import-" + hash + ".zip"
	dest := filepath.Join(cache, name)
	if e = requireHash(dest, hash); e != nil {
		if e = replaceFile(staged, dest); e != nil {
			return nil, e
		}
	}
	record := ImportedBase{1, name, hash, len(index.Files), index.Edition}
	if e = atomicJSON(filepath.Join(cache, "base-source.json"), record); e != nil {
		return nil, e
	}
	return map[string]any{"cached": dest, "files": len(index.Files), "source_modified": false, "uploaded": false, "sha256": hash, "edition": index.Edition}, nil
}
func (m *Manager) importedBase() (string, string, string, error) {
	var r ImportedBase
	if e := readJSON(filepath.Join(m.Data, "cache", "base-source.json"), &r); e != nil {
		return "", "", "", e
	}
	if r.Schema != 1 || !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(r.SHA256) || r.Filename != "base-import-"+r.SHA256+".zip" || r.FileCount != len(baseIndex().Files) {
		return "", "", "", errors.New("invalid private base source record")
	}
	p := filepath.Join(m.Data, "cache", r.Filename)
	if e := noSymlinkAncestors(p); e != nil {
		return "", "", "", e
	}
	if e := regularFile(p); e != nil {
		return "", "", "", e
	}
	if e := requireHash(p, r.SHA256); e != nil {
		return "", "", "", e
	}
	return p, r.SHA256, "base/", nil
}

// Scan known Steam/GOG roots only. No full-disk crawl and no registry changes.
func detectGameFolders() []string {
	home, _ := os.UserHomeDir()
	roots := []string{os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles"), `C:\GOG Games`, `C:\Games`, filepath.Join(home, "GOG Games"), filepath.Join(home, "Games"), "/Applications", filepath.Join(home, "Applications")}
	steam := []string{filepath.Join(home, ".steam", "steam"), filepath.Join(home, ".local", "share", "Steam"), filepath.Join(home, "Library", "Application Support", "Steam")}
	for _, r := range []string{os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles")} {
		if r != "" {
			steam = append(steam, filepath.Join(r, "Steam"))
			roots = append(roots, filepath.Join(r, "GOG Galaxy", "Games"))
		}
	}
	var paths []string
	for _, r := range steam {
		paths = append(paths, filepath.Join(r, "steamapps", "common"))
		f, e := os.Open(filepath.Join(r, "steamapps", "libraryfolders.vdf"))
		if e != nil {
			continue
		}
		b, _ := io.ReadAll(io.LimitReader(f, 1<<20))
		f.Close()
		for _, match := range regexp.MustCompile(`"path"\s+"([^"\r\n]+)"`).FindAllStringSubmatch(string(b), -1) {
			p := strings.ReplaceAll(match[1], `\\`, `\`)
			if filepath.IsAbs(p) {
				paths = append(paths, filepath.Join(p, "steamapps", "common"))
			}
		}
	}
	roots = append(roots, paths...)
	out := []string{}
	seen := map[string]bool{}
	for _, r := range roots {
		if r == "" {
			continue
		}
		for _, name := range []string{"Master of Orion 2", "Master of Orion II", "Master of Orion 2 (DOS)", "Master of Orion 1 and 2"} {
			p := filepath.Join(r, name)
			for _, sub := range []string{"", "DOS", "game", "data", "Master of Orion 2.app/Contents/Resources/game"} {
				q := filepath.Join(p, filepath.FromSlash(sub))
				fs, e := folderFiles(q)
				if e != nil {
					continue
				}
				if _, yes := fs["ORION2.EXE"]; yes {
					if _, ok := fs["GAME.LBX"]; ok && !seen[q] {
						seen[q] = true
						out = append(out, q)
					}
				}
			}
		}
	}
	sort.Strings(out)
	return out
}
func (m *Manager) prepareForPlay(p Profile, source string) (any, error) {
	if _, e := resolve(p); e != nil {
		return nil, e
	}
	// Explicit input supersedes detection. This never sends owned files anywhere.
	if strings.TrimSpace(source) != "" {
		path, e := localPath(source)
		if e != nil {
			return nil, e
		}
		i, e := os.Stat(path)
		if e != nil {
			return nil, e
		}
		if i.IsDir() {
			if _, e = m.importFolder(path); e != nil {
				return nil, e
			}
		} else {
			if _, e = m.importPayload("base", path); e != nil {
				return nil, e
			}
		}
	}
	if _, _, _, e := m.payload("base"); e != nil {
		choices := detectGameFolders()
		if len(choices) != 1 {
			return nil, errors.New("owned game source required: choose a detected game folder, paste its path, or import your recognized base.zip. The game is not downloaded or redistributed")
		}
		if _, e = m.importFolder(choices[0]); e != nil {
			return nil, e
		}
	}
	if p.Engine == "1.50.26" {
		if _, _, _, e := m.payload("patch"); e != nil {
			if _, e = m.refreshPatch(); e != nil {
				return nil, fmt.Errorf("community patch: %w", e)
			}
		}
	}
	if len(m.runtimeCandidates()) == 0 {
		if _, e := m.installRuntime(); e != nil {
			return nil, fmt.Errorf("DOSBox runtime: %w", e)
		}
	}
	result, e := m.build(p)
	if e != nil {
		return nil, e
	}
	return map[string]any{"environment": result, "runtime": m.runtimeCandidates()[0], "ready_for_launch": true, "game_started": false, "notice": "Files and configuration verified. Click Launch MOO2; actual gameplay remains an acceptance test."}, nil
}

// Archive provenance and installed game identity are different. A folder
// snapshot has its own container hash, so verify its actual files against the
// compiled baseline rather than trusting an arbitrary manifest/source hash.
func verifyBaseIdentity(mf Manifest, game string) error {
	if mf.SourceBaseKind == "" && mf.SourceBase == BaseHash {
		return nil
	}
	if mf.SourceBaseKind != "fingerprinted-dos-131" || !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(mf.SourceBase) {
		return errors.New("unrecognized source fingerprint")
	}
	return verifyPinnedBase(game, mf.Files, baseIndex().Files)
}
func verifyPinnedBase(game string, installed map[string]FileRecord, expected []BaseFile) error {
	for _, r := range expected {
		rec, ok := installed[r.Name]
		if !ok || rec.SHA256 != r.SHA256 || rec.Size != r.Size {
			return fmt.Errorf("recognized base file identity missing/changed: %s", r.Name)
		}
		p := filepath.Join(game, r.Name)
		if e := noSymlinkAncestors(p); e != nil {
			return e
		}
		if e := regularFile(p); e != nil {
			return e
		}
		if e := requireHash(p, r.SHA256); e != nil {
			return fmt.Errorf("recognized base file changed: %s", r.Name)
		}
	}
	return nil
}
