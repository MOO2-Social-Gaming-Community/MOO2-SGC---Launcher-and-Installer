package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type FileRecord struct {
	SHA256  string `json:"sha256"`
	Size    int64  `json:"size"`
	Mutable bool   `json:"mutable"`
}
type Manifest struct {
	Schema         int                   `json:"schema"`
	Manager        string                `json:"manager"`
	Created        string                `json:"created"`
	Profile        Profile               `json:"profile"`
	Resolution     Resolution            `json:"resolution"`
	Files          map[string]FileRecord `json:"files"`
	SourceBase     string                `json:"source_base"`
	SourceBaseKind string                `json:"source_base_kind,omitempty"`
	SourcePatch    string                `json:"source_patch"`
}
type Active struct {
	Generation string `json:"generation"`
}
type VerifyResult struct {
	OK          bool     `json:"ok"`
	Checked     int      `json:"checked"`
	Bad         []string `json:"bad"`
	UserChanges []string `json:"user_changes"`
	Fingerprint string   `json:"fingerprint"`
	Generation  string   `json:"generation"`
}
type Job struct {
	Busy    bool   `json:"busy"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Error   string `json:"error"`
	Result  any    `json:"result,omitempty"`
}
type Settings struct {
	RuntimePath string `json:"runtime_path"`
}
type Manager struct {
	Root, Data     string
	AppRoot        string
	mu             sync.Mutex
	op             sync.Mutex
	job            Job
	running        *exec.Cmd
	runningProfile string
	lastExit       string
}

func newManager(root, data string) (*Manager, error) {
	root, e := filepath.Abs(root)
	if e != nil {
		return nil, e
	}
	data, e = filepath.Abs(data)
	if e != nil {
		return nil, e
	}
	if e = noSymlinkAncestors(data); e != nil {
		return nil, e
	}
	for _, d := range []string{"profiles", "active", "environments", "cache", "logs"} {
		if e = os.MkdirAll(filepath.Join(data, d), 0700); e != nil {
			return nil, e
		}
	}
	m := &Manager{Root: root, Data: data, job: Job{Message: "Ready. Select a profile, then Prepare environment."}}
	for _, p := range defaultProfiles() {
		f := filepath.Join(data, "profiles", p.ID+".json")
		if _, e = os.Stat(f); os.IsNotExist(e) {
			if e = atomicJSON(f, p); e != nil {
				return nil, e
			}
		}
	}
	return m, nil
}
func (m *Manager) progress(s string) { m.mu.Lock(); m.job.Message = s; m.mu.Unlock() }
func (m *Manager) startJob(kind string, fn func() (any, error)) error {
	m.mu.Lock()
	if m.job.Busy || m.running != nil {
		m.mu.Unlock()
		return errors.New("another operation or game is running; finish it first")
	}
	m.job = Job{Busy: true, Kind: kind, Message: "Starting " + kind}
	m.mu.Unlock()
	go func() {
		m.op.Lock()
		defer m.op.Unlock()
		r, e := fn()
		m.mu.Lock()
		defer m.mu.Unlock()
		m.job.Busy = false
		m.job.Result = r
		if e != nil {
			m.job.Error = e.Error()
			m.job.Message = kind + " stopped"
		} else {
			m.job.Message = kind + " completed"
		}
	}()
	return nil
}
func (m *Manager) loadProfile(id string) (Profile, error) {
	var p Profile
	if !idPattern.MatchString(id) {
		return p, errors.New("invalid profile ID")
	}
	e := readJSON(filepath.Join(m.Data, "profiles", id+".json"), &p)
	if e != nil {
		return p, e
	}
	if p.ID != id {
		return p, errors.New("profile ID mismatch")
	}
	_, e = resolve(p)
	return p, e
}
func (m *Manager) saveProfile(p Profile) error {
	if _, e := resolve(p); e != nil {
		return e
	}
	return atomicJSON(filepath.Join(m.Data, "profiles", p.ID+".json"), p)
}
func (m *Manager) profiles() ([]Profile, error) {
	entries, e := os.ReadDir(filepath.Join(m.Data, "profiles"))
	if e != nil {
		return nil, e
	}
	out := []Profile{}
	for _, f := range entries {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		p, e := m.loadProfile(strings.TrimSuffix(f.Name(), ".json"))
		if e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	return out, nil
}
func (m *Manager) activePath(id string) (string, string, error) {
	if !idPattern.MatchString(id) {
		return "", "", errors.New("invalid profile ID")
	}
	var a Active
	if e := readJSON(filepath.Join(m.Data, "active", id+".json"), &a); e != nil {
		return "", "", e
	}
	if !idPattern.MatchString(a.Generation) {
		return "", "", errors.New("invalid generation")
	}
	p := filepath.Join(m.Data, "environments", id, a.Generation)
	if e := noSymlinkAncestors(p); e != nil {
		return "", "", e
	}
	return p, a.Generation, nil
}
func (m *Manager) payload(which string) (string, string, string, error) {
	if which == "base" {
		if _, e := os.Stat(filepath.Join(m.Data, "cache", "base-source.json")); e == nil {
			return m.importedBase()
		} else if !os.IsNotExist(e) {
			return "", "", "", e
		}
	}
	name, want, prefix := "base.zip", BaseHash, "Master of Orion 2/"
	if which == "patch" {
		name, want, prefix = "patch-1.50.26.zip", PatchHash, "MOO2-1.50.26/patch/"
	}
	candidates := []string{filepath.Join(m.Data, "cache", name), filepath.Join(m.Root, "payloads", name)}
	for _, p := range candidates {
		if _, e := os.Stat(p); e == nil {
			if e = regularFile(p); e != nil {
				return "", "", "", e
			}
			if e = requireHash(p, want); e != nil {
				return "", "", "", e
			}
			return p, want, prefix, nil
		}
	}
	return "", "", "", fmt.Errorf("missing %s; import your local archive in Packages & updates, or retain the existing private payloads folder", name)
}
func (m *Manager) build(p Profile) (any, error) {
	r, e := resolve(p)
	if e != nil {
		return nil, e
	}
	m.progress("Verifying your source archive before installation…")
	base, baseHash, bp, e := m.payload("base")
	if e != nil {
		return nil, e
	}
	var patch, pp string
	if p.Engine == "1.50.26" {
		patch, _, pp, e = m.payload("patch")
		if e != nil {
			return nil, e
		}
	}
	parent := filepath.Join(m.Data, "environments", p.ID)
	if e = os.MkdirAll(parent, 0700); e != nil {
		return nil, e
	}
	stage, e := os.MkdirTemp(parent, "stage-")
	if e != nil {
		return nil, e
	}
	defer os.RemoveAll(stage)
	game := filepath.Join(stage, "game")
	m.progress("Extracting a fresh game copy; archived personal saves stay in the source ZIP…")
	if e = extractZip(base, game, bp, true); e != nil {
		return nil, e
	}
	if p.Engine == "1.50.26" {
		m.progress("Applying the unmodified community 1.50.26 overlay…")
		if e = extractZip(patch, game, pp, false); e != nil {
			return nil, e
		}
		if e = atomicWrite(filepath.Join(game, "150", "ENABLE.CFG"), []byte(r.Config)); e != nil {
			return nil, e
		}
	}
	// Full local data lives on the emulated C: drive, not the archive owner's H: path.
	if e = atomicWrite(filepath.Join(game, "orioncd.ini"), []byte("C:\\\r\n")); e != nil {
		return nil, e
	}
	if e = requireHash(filepath.Join(game, "ORION2.EXE"), BaseEngineHash); e != nil {
		return nil, e
	}
	if p.Engine == "1.50.26" {
		if e = requireHash(filepath.Join(game, "ORION150.EXE"), EngineHash); e != nil {
			return nil, e
		}
	}
	old, _, oe := m.activePath(p.ID)
	if oe == nil {
		var om Manifest
		if e = readJSON(filepath.Join(old, "manifest.json"), &om); e != nil {
			return nil, e
		}
		// Never migrate personal state across engine versions automatically.
		if om.Profile.Engine == p.Engine {
			m.progress("Preserving this profile's saves and editable settings; retaining the previous generation…")
			e = walkRegular(filepath.Join(old, "game"), func(n, src string) error {
				if isUserFile(n) {
					return copyFile(src, filepath.Join(game, filepath.FromSlash(n)))
				}
				return nil
			})
			if e != nil {
				return nil, e
			}
		}
	} else if !os.IsNotExist(oe) {
		return nil, oe
	}
	manifest := Manifest{Schema: 1, Manager: Version, Created: time.Now().UTC().Format(time.RFC3339Nano), Profile: p, Resolution: r, SourceBase: baseHash, Files: map[string]FileRecord{}}
	if bp == "base/" {
		manifest.SourceBaseKind = "fingerprinted-dos-131"
	}
	if p.Engine == "1.50.26" {
		manifest.SourcePatch = PatchHash
	}
	m.progress("Hashing installed files and writing the resolved profile lock…")
	e = walkRegular(game, func(n, f string) error {
		h, e := hashFile(f)
		if e != nil {
			return e
		}
		i, e := os.Stat(f)
		if e != nil {
			return e
		}
		manifest.Files[n] = FileRecord{h, i.Size(), isUserFile(n)}
		return nil
	})
	if e != nil {
		return nil, e
	}
	if e = atomicJSON(filepath.Join(stage, "manifest.json"), manifest); e != nil {
		return nil, e
	}
	v, e := verifyDir(stage)
	if e != nil {
		return nil, e
	}
	if !v.OK {
		return nil, fmt.Errorf("staging verification failed: %v", v.Bad)
	}
	generation := "g" + fmt.Sprint(time.Now().UnixNano())
	dest := filepath.Join(parent, generation)
	cfg, e := dosboxConfig(filepath.Join(dest, "game"), p)
	if e != nil {
		return nil, e
	}
	if e = atomicWrite(filepath.Join(stage, "dosbox-manager.conf"), []byte(cfg)); e != nil {
		return nil, e
	}
	if e = os.Rename(stage, dest); e != nil {
		return nil, e
	}
	// Activation pointer is the final transaction commit; prior generations are not deleted.
	if e = atomicJSON(filepath.Join(m.Data, "active", p.ID+".json"), Active{generation}); e != nil {
		return nil, e
	}
	if e = m.saveProfile(p); e != nil {
		return nil, e
	}
	v.Generation = generation
	return map[string]any{"verification": v, "path": dest, "prsl": "not injected", "chat": "not injected", "archived_saves": "retained only in original source ZIP", "engine": p.Engine}, nil
}
func verifyDir(dir string) (VerifyResult, error) {
	v := VerifyResult{OK: true, Bad: []string{}, UserChanges: []string{}}
	var mf Manifest
	if e := noSymlinkAncestors(dir); e != nil {
		return v, e
	}
	if e := readJSON(filepath.Join(dir, "manifest.json"), &mf); e != nil {
		return v, e
	}
	if mf.Schema != 1 || len(mf.Files) == 0 || len(mf.Files) > 10000 {
		return v, errors.New("unsupported or empty manifest")
	}
	resolution, e := resolve(mf.Profile)
	if e != nil {
		return v, e
	}
	if mf.Profile.Engine == "1.50.26" && mf.SourcePatch != PatchHash {
		return v, errors.New("unrecognized patch fingerprint")
	}
	game := filepath.Join(dir, "game")
	if e = verifyBaseIdentity(mf, game); e != nil {
		return v, e
	}
	seen := map[string]bool{}
	for n, rec := range mf.Files {
		rel, e := safeRel(n)
		if e != nil {
			return v, e
		}
		f := filepath.Join(game, rel)
		if e = noSymlinkAncestors(f); e != nil {
			return v, e
		}
		if strings.EqualFold(n, "150/ENABLE.CFG") && mf.Profile.Engine == "1.50.26" {
			b, e := os.ReadFile(f)
			if e != nil || string(b) != resolution.Config {
				v.Bad = append(v.Bad, n+" (selected mods differ)")
			}
		}
		v.Checked++
		h, e := hashFile(f)
		if e != nil || h != rec.SHA256 {
			if isUserFile(n) && (e == nil || os.IsNotExist(e) && isPersonalSeed(strings.ToLower(n))) {
				v.UserChanges = append(v.UserChanges, n)
			} else {
				v.Bad = append(v.Bad, n)
			}
		}
		seen[n] = true
	}
	e = walkRegular(game, func(n, f string) error {
		if seen[n] {
			return nil
		}
		if isUserFile(n) {
			v.UserChanges = append(v.UserChanges, n)
			return nil
		}
		switch strings.ToLower(filepath.Ext(n)) {
		case ".exe", ".com", ".dll", ".lbx", ".lua", ".cfg":
			v.Bad = append(v.Bad, n+" (unmanaged code/config/data)")
		}
		return nil
	})
	if e != nil {
		return v, e
	}
	exe, want := "ORION2.EXE", BaseEngineHash
	if mf.Profile.Engine == "1.50.26" {
		exe, want = "ORION150.EXE", EngineHash
	}
	if e = requireHash(filepath.Join(game, exe), want); e != nil {
		v.Bad = append(v.Bad, exe+" (exact engine check)")
	}
	// Include active editable CFG contents in the multiplayer fingerprint; not display/network settings.
	h := sha256.New()
	h.Write([]byte(resolution.Fingerprint))
	cfgNames := []string{}
	for n := range mf.Files {
		if strings.EqualFold(filepath.Ext(n), ".cfg") {
			cfgNames = append(cfgNames, n)
		}
	}
	sort.Strings(cfgNames)
	for _, n := range cfgNames {
		b, e := os.ReadFile(filepath.Join(game, filepath.FromSlash(n)))
		if e != nil {
			v.Bad = append(v.Bad, n)
			continue
		}
		h.Write([]byte(n))
		h.Write([]byte{0})
		h.Write(b)
		h.Write([]byte{0})
	}
	v.Fingerprint = hex.EncodeToString(h.Sum(nil))
	sort.Strings(v.Bad)
	sort.Strings(v.UserChanges)
	v.OK = len(v.Bad) == 0
	return v, nil
}
func (m *Manager) verify(id string) (VerifyResult, error) {
	dir, g, e := m.activePath(id)
	if e != nil {
		return VerifyResult{}, e
	}
	v, e := verifyDir(dir)
	v.Generation = g
	return v, e
}
func (m *Manager) history(id string) ([]map[string]string, error) {
	if !idPattern.MatchString(id) {
		return nil, errors.New("invalid profile")
	}
	entries, e := os.ReadDir(filepath.Join(m.Data, "environments", id))
	if os.IsNotExist(e) {
		return []map[string]string{}, nil
	}
	if e != nil {
		return nil, e
	}
	out := []map[string]string{}
	for _, d := range entries {
		if !d.IsDir() || !strings.HasPrefix(d.Name(), "g") {
			continue
		}
		var mf Manifest
		if e = readJSON(filepath.Join(m.Data, "environments", id, d.Name(), "manifest.json"), &mf); e != nil {
			continue
		}
		out = append(out, map[string]string{"generation": d.Name(), "created": mf.Created, "engine": mf.Profile.Engine, "core": mf.Profile.Core})
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["generation"] > out[j]["generation"] })
	return out, nil
}
func (m *Manager) activate(id, generation string) (any, error) {
	if !idPattern.MatchString(id) || !idPattern.MatchString(generation) {
		return nil, errors.New("invalid generation or profile")
	}
	dir := filepath.Join(m.Data, "environments", id, generation)
	v, e := verifyDir(dir)
	if e != nil {
		return nil, e
	}
	if !v.OK {
		return nil, fmt.Errorf("generation is corrupt: %v", v.Bad)
	}
	var mf Manifest
	if e = readJSON(filepath.Join(dir, "manifest.json"), &mf); e != nil {
		return nil, e
	}
	if mf.Profile.ID != id {
		return nil, errors.New("generation belongs to another profile")
	}
	// No save merging on rollback: the current generation remains retained.
	if e = atomicJSON(filepath.Join(m.Data, "active", id+".json"), Active{generation}); e != nil {
		return nil, e
	}
	e = m.saveProfile(mf.Profile)
	return v, e
}
func (m *Manager) removeEnvironment(id string) (any, error) {
	if !idPattern.MatchString(id) {
		return nil, errors.New("invalid profile")
	}
	e := os.Remove(filepath.Join(m.Data, "active", id+".json"))
	if e != nil {
		return nil, e
	}
	return map[string]string{"status": "deactivated", "saves": "All generations and saves retained. Prepare to reinstall, or activate a retained generation."}, nil
}
func (m *Manager) getSettings() Settings {
	var s Settings
	_ = readJSON(filepath.Join(m.Data, "settings.json"), &s)
	return s
}
func (m *Manager) exportDiagnostics(id string) (any, error) {
	p, e := m.loadProfile(id)
	if e != nil {
		return nil, e
	}
	r, e := resolve(p)
	if e != nil {
		return nil, e
	}
	v, ve := m.verify(id)
	report := map[string]any{"schema": 1, "manager": Version, "platform": platform(), "profile": p, "resolution": r, "verification": v, "runtime_candidates": m.runtimeCandidates(), "live_prsl": false, "new_chat": false, "saved_games_included": false, "timestamp": time.Now().UTC().Format(time.RFC3339)}
	if ve != nil {
		report["verification_error"] = ve.Error()
	}
	name := "diagnostics-" + id + "-" + time.Now().UTC().Format("20060102T150405Z") + ".json"
	f := filepath.Join(m.Data, "logs", name)
	e = atomicJSON(f, report)
	return map[string]any{"path": f, "report": report}, e
}
func (m *Manager) state() (any, error) {
	ps, e := m.profiles()
	if e != nil {
		return nil, e
	}
	active := map[string]string{}
	hist := map[string]any{}
	for _, p := range ps {
		_, g, e := m.activePath(p.ID)
		if e == nil {
			active[p.ID] = g
		}
		h, _ := m.history(p.ID)
		hist[p.ID] = h
	}
	m.mu.Lock()
	j, r, exit := m.job, m.runningProfile, m.lastExit
	m.mu.Unlock()
	return map[string]any{"version": Version, "platform": platform(), "profiles": ps, "catalog": catalog(), "active": active, "history": hist, "job": j, "running": r, "last_exit": exit, "data_path": m.Data, "runtime_candidates": m.runtimeCandidates(), "settings": m.getSettings(), "runtime_recipe": runtimeRecipeForPlatform(), "prsl_available": false, "chat_available": false, "distribution": m.distributionStatus()}, nil
}
func encode(v any) string { b, _ := json.MarshalIndent(v, "", "  "); return string(b) }
