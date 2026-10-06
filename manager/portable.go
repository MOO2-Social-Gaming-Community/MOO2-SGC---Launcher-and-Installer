package main

import (
	"errors"
	"fmt"
	"moo2manager/shared/layout"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Only the named baseline uses ROOT/game. All other engine/mod profiles retain
// the existing generation-based workspaces under ROOT/userdata/environments.
func (m *Manager) isPortableBaseline(id string) bool { return m.Portable && id == "baseline" }
func workspaceManifestPath(dir string) string {
	portable, e := layout.IsPortable(dir)
	if e == nil && portable {
		return filepath.Join(dir, "state", "baseline.manifest.json")
	}
	return filepath.Join(dir, "manifest.json")
}
func readWorkspaceManifest(dir string, mf *Manifest) error {
	if e := noSymlinkAncestors(dir); e != nil {
		return e
	}
	if _, e := os.Lstat(filepath.Join(dir, "state", "baseline-transaction.json")); e == nil {
		return errors.New("interrupted portable baseline transaction: run SGC-Recover-Baseline.cmd or choose Recover interrupted baseline before continuing")
	} else if !os.IsNotExist(e) {
		return e
	}
	return readJSON(workspaceManifestPath(dir), mf)
}
func (m *Manager) portableActive() (string, string, error) {
	var a Active
	if e := readJSON(filepath.Join(m.Root, "state", "baseline-active.json"), &a); e != nil {
		return "", "", e
	}
	if !idPattern.MatchString(a.Generation) {
		return "", "", errors.New("invalid portable generation")
	}
	if _, e := os.Stat(filepath.Join(m.Root, "state", "baseline-transaction.json")); e == nil {
		return "", "", errors.New("portable baseline recovery required")
	}
	return m.Root, a.Generation, nil
}
func (m *Manager) portableSource() (string, error) {
	entries, e := os.ReadDir(m.Root)
	if e != nil {
		return "", e
	}
	candidates := []string{}
	for _, f := range entries {
		if f.IsDir() || !strings.HasSuffix(strings.ToLower(f.Name()), ".zip") {
			continue
		}
		// Never import or expand every ZIP in a large root. Match the documented
		// baseline naming family, then identify its content cryptographically.
		if !strings.HasPrefix(strings.ToLower(f.Name()), "master of orion 2 - v1_40b23") {
			continue
		}
		p := filepath.Join(m.Root, f.Name())
		if e = regularFile(p); e != nil {
			return "", e
		}
		if e = requireHash(p, ManualBaselineArchiveHash); e != nil {
			return "", fmt.Errorf("portable baseline archive was changed: %w; select a supported source explicitly", e)
		}
		candidates = append(candidates, p)
	}
	sort.Strings(candidates)
	if len(candidates) > 0 {
		return candidates[0], nil
	} // Same verified bytes if duplicated.
	return "", os.ErrNotExist
}
func (m *Manager) preserveUnregisteredBaseline(game string) error {
	old := filepath.Join(m.Root, "game")
	if _, e := os.Stat(old); os.IsNotExist(e) {
		return nil
	} else if e != nil {
		return e
	}
	if e := noSymlinkAncestors(old); e != nil {
		return e
	}
	fs, e := folderFiles(old)
	if e != nil {
		return e
	}
	p, ok := fs["ORION2.EXE"]
	if !ok { // A newly extracted harness contains only a placeholder.
		if len(fs) <= 1 {
			return nil
		}
		return errors.New("existing unregistered game directory has no recognized engine; preserve it separately before preparing")
	}
	h, e := hashFile(p)
	if e != nil {
		return e
	}
	if h != BaselineEngineHash {
		if h == CDEngineHash || h == BaseEngineHash {
			return nil
		} // Backed up, but no cross-version save migration.
		return errors.New("existing unregistered game has an unknown engine; it was not overwritten")
	}
	return preserveUserState(old, game, true)
}
func preserveUserState(old, game string, personalOnly bool) error {
	return walkRegular(old, func(n, src string) error {
		if !isUserFile(n) || personalOnly && !isPersonalSeed(strings.ToLower(n)) {
			return nil
		}
		destName := n
		for _, canonical := range []string{"DIG.INI", "MDI.INI"} {
			if strings.EqualFold(n, canonical) {
				destName = canonical
			}
		}
		return copyFile(src, filepath.Join(game, filepath.FromSlash(destName)))
	})
}

type baselineTransaction struct {
	Schema        int    `json:"schema"`
	Backup        string `json:"backup"`
	NewGeneration string `json:"new_generation"`
	HadGame       bool   `json:"had_game"`
}

func (m *Manager) baselineJournal() string {
	return filepath.Join(m.Root, "state", "baseline-transaction.json")
}
func exists(p string) bool { _, e := os.Lstat(p); return e == nil }

// One rename switches the actual game directory. A write-ahead journal covers
// the necessary metadata renames and offers non-destructive crash recovery.
// No Windows symlinks, junctions, admin rights or PowerShell are needed.
func (m *Manager) commitPortableBaseline(stage string, p Profile, v VerifyResult) (result any, err error) {
	if !m.isPortableBaseline(p.ID) || p.Engine != "1.40b23" {
		return nil, errors.New("only the 1.40b23 baseline may occupy portable game/")
	}
	if exists(m.baselineJournal()) {
		return nil, errors.New("recover the interrupted baseline transaction first")
	}
	var mf Manifest
	if err = readJSON(filepath.Join(stage, "manifest.json"), &mf); err != nil {
		return nil, err
	}
	if !v.OK {
		return nil, errors.New("unverified baseline cannot be activated")
	}
	for _, d := range []string{"state", "backups/baseline", "logs", "config"} {
		dest := filepath.Join(m.Root, filepath.FromSlash(d))
		if err = noSymlinkAncestors(dest); err != nil {
			return nil, err
		}
		if err = os.MkdirAll(dest, 0700); err != nil {
			return nil, err
		}
	}
	generation := "g" + fmt.Sprint(time.Now().UnixNano())
	backupID := "b" + fmt.Sprint(time.Now().UnixNano())
	backup := filepath.Join(m.Root, "backups", "baseline", backupID)
	if err = os.Mkdir(backup, 0700); err != nil {
		return nil, err
	}
	for src, n := range map[string]string{workspaceManifestPath(m.Root): "manifest.json", filepath.Join(m.Root, "state", "baseline-active.json"): "active.json"} {
		if exists(src) {
			if err = copyFile(src, filepath.Join(backup, n)); err != nil {
				return nil, err
			}
		}
	}
	j := baselineTransaction{1, backupID, generation, exists(filepath.Join(m.Root, "game"))}
	if err = atomicJSON(m.baselineJournal(), j); err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if _, e := m.recoverPortableBaseline(); e != nil {
				err = fmt.Errorf("%w; recovery also stopped: %v. Backups retained at %s", err, e, backup)
			}
		}
	}()
	game := filepath.Join(m.Root, "game")
	if j.HadGame {
		if err = os.Rename(game, filepath.Join(backup, "game")); err != nil {
			return nil, err
		}
	}
	if err = os.Rename(filepath.Join(stage, "game"), game); err != nil {
		return nil, err
	}
	if err = atomicJSON(workspaceManifestPath(m.Root), mf); err != nil {
		return nil, err
	}
	if err = atomicJSON(filepath.Join(m.Root, "state", "baseline-active.json"), Active{generation}); err != nil {
		return nil, err
	}
	if err = os.Remove(m.baselineJournal()); err != nil {
		return nil, err
	}
	v.Generation = generation
	return map[string]any{"verification": v, "path": m.Root, "game_path": game, "engine": p.Engine, "source_version": mf.Lineage[0].Version, "lineage": mf.Lineage, "backup": backup, "source_modified": false, "prsl": "not injected", "chat": "not injected"}, nil
}
func (m *Manager) recoverPortableBaseline() (any, error) {
	var e error
	if !m.Portable {
		return nil, errors.New("not a portable root")
	}
	var j baselineTransaction
	if e := readJSON(m.baselineJournal(), &j); e != nil {
		if os.IsNotExist(e) {
			return map[string]string{"status": "no interrupted transaction"}, nil
		}
		return nil, e
	}
	if j.Schema != 1 || !idPattern.MatchString(j.Backup) || !strings.HasPrefix(j.Backup, "b") || !idPattern.MatchString(j.NewGeneration) || !strings.HasPrefix(j.NewGeneration, "g") {
		return nil, errors.New("invalid recovery journal; no files changed")
	}
	backup := filepath.Join(m.Root, "backups", "baseline", j.Backup)
	game := filepath.Join(m.Root, "game")
	if e := noSymlinkAncestors(backup); e != nil {
		return nil, e
	}
	if e := noSymlinkAncestors(game); e != nil {
		return nil, e
	}
	var a Active
	if e := readJSON(filepath.Join(m.Root, "state", "baseline-active.json"), &a); e == nil && a.Generation == j.NewGeneration {
		// A committed receipt only authorizes cleanup after checking the complete game.
		var mf Manifest
		if e = readJSON(workspaceManifestPath(m.Root), &mf); e != nil {
			return nil, e
		}
		if e = verifyBaseIdentity(mf, game); e != nil {
			return nil, e
		}
		for n, r := range mf.Files {
			rel, e := safeRel(n)
			if e != nil {
				return nil, e
			}
			if !isUserFile(n) {
				if e = requireHash(filepath.Join(game, rel), r.SHA256); e != nil {
					return nil, e
				}
			}
		}
		if e = os.Remove(m.baselineJournal()); e != nil {
			return nil, e
		}
		return map[string]string{"status": "completed interrupted commit; game and receipts verified"}, nil
	}
	if exists(filepath.Join(backup, "game")) || !j.HadGame {
		if exists(game) {
			failed := filepath.Join(m.Root, "backups", "baseline", "failed-"+fmt.Sprint(time.Now().UnixNano()))
			if e = os.Mkdir(failed, 0700); e != nil {
				return nil, e
			}
			if e = os.Rename(game, filepath.Join(failed, "game")); e != nil {
				return nil, e
			}
		}
		if j.HadGame {
			if e := os.Rename(filepath.Join(backup, "game"), game); e != nil {
				return nil, e
			}
		}
	}
	// If no backup/game exists and HadGame is true, the initial rename never
	// happened (or a prior recovery already restored it). Never delete game/.
	for n, dest := range map[string]string{"manifest.json": workspaceManifestPath(m.Root), "active.json": filepath.Join(m.Root, "state", "baseline-active.json")} {
		src := filepath.Join(backup, n)
		if exists(src) {
			if e := copyFile(src, dest); e != nil {
				return nil, e
			}
		} else if e := os.Remove(dest); e != nil && !os.IsNotExist(e) {
			return nil, e
		}
	}
	if e := os.Remove(m.baselineJournal()); e != nil {
		return nil, e
	}
	return map[string]string{"status": "previous portable baseline restored; failed candidate retained", "backup": backup}, nil
}
func (m *Manager) portableHistory() ([]map[string]string, error) {
	out := []map[string]string{}
	var current Manifest
	if e := readWorkspaceManifest(m.Root, &current); e == nil {
		_, g, e := m.portableActive()
		if e == nil {
			out = append(out, map[string]string{"generation": g, "created": current.Created, "engine": current.Profile.Engine, "core": current.Profile.Core})
		}
	}
	entries, e := os.ReadDir(filepath.Join(m.Root, "backups", "baseline"))
	if os.IsNotExist(e) {
		return out, nil
	}
	if e != nil {
		return nil, e
	}
	for _, f := range entries {
		if !f.IsDir() || !strings.HasPrefix(f.Name(), "b") {
			continue
		}
		var mf Manifest
		d := filepath.Join(m.Root, "backups", "baseline", f.Name())
		if e = readJSON(filepath.Join(d, "manifest.json"), &mf); e != nil || !exists(filepath.Join(d, "game")) {
			continue
		}
		out = append(out, map[string]string{"generation": f.Name(), "created": mf.Created, "engine": mf.Profile.Engine, "core": mf.Profile.Core})
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["created"] > out[j]["created"] })
	return out, nil
}
func (m *Manager) activatePortable(g string) (any, error) {
	if !idPattern.MatchString(g) {
		return nil, errors.New("invalid portable snapshot")
	}
	_, current, e := m.portableActive()
	if e == nil && g == current {
		return m.verify("baseline")
	}
	if !strings.HasPrefix(g, "b") {
		return nil, errors.New("choose a retained baseline backup")
	}
	from := filepath.Join(m.Root, "backups", "baseline", g)
	v, e := verifyDir(from)
	if e != nil {
		return nil, e
	}
	if !v.OK {
		return nil, errors.New("backup verification failed")
	}
	var mf Manifest
	if e = readWorkspaceManifest(from, &mf); e != nil {
		return nil, e
	}
	if mf.Profile.ID != "baseline" || mf.Profile.Engine != "1.40b23" {
		return nil, errors.New("snapshot is not the portable baseline")
	}
	stage, e := os.MkdirTemp(filepath.Join(m.Data, "environments"), "rollback-")
	if e != nil {
		return nil, e
	}
	defer os.RemoveAll(stage)
	if e = walkRegular(from, func(n, src string) error { return copyFile(src, filepath.Join(stage, filepath.FromSlash(n))) }); e != nil {
		return nil, e
	}
	result, e := m.commitPortableBaseline(stage, mf.Profile, v)
	if e != nil {
		return nil, e
	}
	if e = m.saveProfile(mf.Profile); e != nil {
		return nil, e
	}
	return result, nil
}
