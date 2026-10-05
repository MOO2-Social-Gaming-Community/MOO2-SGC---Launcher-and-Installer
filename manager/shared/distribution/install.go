package distribution

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Receipt struct {
	Schema            int                   `json:"schema"`
	Package           Package               `json:"package"`
	Manifest          []byte                `json:"manifest"` // JSON encodes exact signed bytes in base64
	ManifestSignature []byte                `json:"manifest_signature"`
	Files             map[string]FileRecord `json:"files"`
}
type Pointer struct {
	Schema   int    `json:"schema"`
	Current  string `json:"current"`
	Previous string `json:"previous,omitempty"`
}
type Installer struct {
	Root  string
	Trust Trust
	// Health executes ONLY the verified entrypoint, before pointer activation.
	Health func(context.Context, string, string) error
}

func (i Installer) base() string { return filepath.Join(i.Root, "components", "launcher") }
func (i Installer) Current() (Pointer, error) {
	var p Pointer
	e := ReadJSON(filepath.Join(i.base(), "current.json"), &p)
	if e != nil {
		return p, e
	}
	if p.Schema != Schema || !tokenRE.MatchString(p.Current) || (p.Previous != "" && !tokenRE.MatchString(p.Previous)) {
		return p, errors.New("invalid launcher activation pointer")
	}
	return p, nil
}
func (i Installer) VerifyGeneration(g string) (string, Receipt, error) {
	var r Receipt
	if !tokenRE.MatchString(g) {
		return "", r, errors.New("invalid generation")
	}
	base := filepath.Join(i.base(), "versions", g)
	if e := ReadJSON(filepath.Join(base, "receipt.json"), &r); e != nil {
		return "", r, e
	}
	if r.Schema != Schema || len(r.Files) == 0 || len(r.Files) > MaxEntries {
		return "", r, errors.New("invalid installation receipt")
	}
	// Already installed verified software remains runnable offline after metadata
	// expires. Expired metadata can NEVER authorize a new install or update.
	v, e := i.Trust.VerifyManifest(r.Manifest, r.ManifestSignature, time.Now(), TrustedState{}, false)
	if e != nil {
		return "", r, e
	}
	expected, e := v.Select("launcher", r.Package.Platform)
	if e != nil {
		return "", r, e
	}
	if string(PackageMessage(expected)) != string(PackageMessage(r.Package)) {
		return "", r, errors.New("receipt differs from signed package")
	}
	if len(r.Files) != len(expected.Files) {
		return "", r, errors.New("receipt file set differs from signed file index")
	}
	for n, record := range expected.Files {
		if r.Files[n] != record {
			return "", r, errors.New("receipt file hash differs from signed file index")
		}
	}
	dir := filepath.Join(base, "files")
	for n, rec := range r.Files {
		rel, e := SafePath(n)
		if e != nil {
			return "", r, e
		}
		h, size, e := HashFile(filepath.Join(dir, rel))
		if e != nil {
			return "", r, e
		}
		if h != rec.SHA256 || size != rec.Size {
			return "", r, fmt.Errorf("installed file corrupted: %s", n)
		}
	}
	// Detect extra files too; a rogue library beside a valid executable is unsafe.
	e = filepath.WalkDir(dir, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.Type()&os.ModeSymlink != 0 {
			return errors.New("installed symlink")
		}
		if d.IsDir() {
			return nil
		}
		n, e := filepath.Rel(dir, p)
		if e != nil {
			return e
		}
		if _, ok := r.Files[filepath.ToSlash(n)]; !ok {
			return errors.New("unmanaged installed file")
		}
		return nil
	})
	if e != nil {
		return "", r, e
	}
	rel, e := SafePath(r.Package.Entrypoint)
	if e != nil {
		return "", r, e
	}
	if _, ok := r.Files[r.Package.Entrypoint]; !ok {
		return "", r, errors.New("entrypoint not in receipt")
	}
	return filepath.Join(dir, rel), r, nil
}

// Install never overwrites a live binary: prepare a new immutable generation,
// verify files, health-check, then atomically switch the small activation pointer.
// Callers hold Acquire(Root), shared with the running launcher.
func (i Installer) Install(ctx context.Context, v VerifiedManifest, p Package, archive string) (string, error) {
	if i.Health == nil {
		return "", errors.New("launcher health check is required")
	}
	prior, e := LoadState(i.Root, i.Trust)
	if e != nil {
		return "", e
	}
	checked, e := i.Trust.VerifyManifest(v.Bytes, v.SignatureBytes, time.Now(), prior, true)
	if e != nil {
		return "", e
	}
	expected, e := checked.Select("launcher", p.Platform)
	if e != nil {
		return "", e
	}
	if string(PackageMessage(expected)) != string(PackageMessage(p)) {
		return "", errors.New("package is not the signed launcher selection")
	}
	p = expected
	if e = i.Trust.VerifyPackage(p); e != nil {
		return "", e
	}
	if e = VerifyFile(archive, p); e != nil {
		return "", e
	}
	var old Pointer
	if q, err := i.Current(); err == nil {
		old = q
	} else if !os.IsNotExist(err) {
		return "", err
	}
	parent := filepath.Join(i.base(), "versions")
	if e = NoLinks(parent); e != nil {
		return "", e
	}
	if e = os.MkdirAll(parent, 0700); e != nil {
		return "", e
	}
	stage, e := os.MkdirTemp(parent, "stage-")
	if e != nil {
		return "", e
	}
	defer os.RemoveAll(stage)
	files, e := Extract(archive, filepath.Join(stage, "files"))
	if e != nil {
		return "", e
	}
	if len(files) != len(p.Files) {
		return "", errors.New("archive file set differs from signed file index")
	}
	for n, r := range p.Files {
		if files[n] != r {
			return "", errors.New("archive entry differs from signed file index")
		}
	}
	if _, ok := files[p.Entrypoint]; !ok {
		return "", errors.New("package entrypoint missing")
	}
	rel, _ := SafePath(p.Entrypoint)
	entry := filepath.Join(stage, "files", rel)
	if e = os.Chmod(entry, 0700); e != nil {
		return "", e
	}
	// The signed launcher update helper must also be executable on Unix.
	if _, ok := expected.Files["MOO2-SGC-Setup"]; ok {
		if e = os.Chmod(filepath.Join(stage, "files", "MOO2-SGC-Setup"), 0700); e != nil {
			return "", e
		}
	}
	r := Receipt{Schema, p, v.Bytes, v.SignatureBytes, files}
	if e = AtomicJSON(filepath.Join(stage, "receipt.json"), r); e != nil {
		return "", e
	}
	b := make([]byte, 8)
	if _, e = rand.Read(b); e != nil {
		return "", e
	}
	generation := "g-" + p.SHA256[:16] + "-" + hex.EncodeToString(b)
	dest := filepath.Join(parent, generation)
	if e = os.Rename(stage, dest); e != nil {
		return "", e
	}
	activated := false
	defer func() {
		if !activated {
			_ = os.RemoveAll(dest)
		}
	}()
	entry, _, e = i.VerifyGeneration(generation)
	if e != nil {
		return "", e
	}
	if e = i.Health(ctx, entry, p.Version); e != nil {
		return "", fmt.Errorf("launcher health check failed; previous version retained: %w", e)
	}
	// Verify again after health check in case a malformed package modified itself.
	if entry, _, e = i.VerifyGeneration(generation); e != nil {
		return "", e
	}
	if e = AtomicJSON(filepath.Join(i.base(), "current.json"), Pointer{Schema, generation, old.Current}); e != nil {
		return "", e
	}
	activated = true
	return entry, nil
}
func (i Installer) Rollback() (string, error) {
	p, e := i.Current()
	if e != nil {
		return "", e
	}
	if p.Previous == "" {
		return "", errors.New("no previous launcher generation")
	}
	entry, _, e := i.VerifyGeneration(p.Previous)
	if e != nil {
		return "", e
	}
	e = AtomicJSON(filepath.Join(i.base(), "current.json"), Pointer{Schema, p.Previous, p.Current})
	return entry, e
}
func LoadState(root string, t Trust) (TrustedState, error) {
	var s TrustedState
	e := ReadJSON(filepath.Join(root, "distribution", "trusted-state.json"), &s)
	if os.IsNotExist(e) {
		return TrustedState{}, nil
	}
	if e != nil {
		return s, e
	}
	if s.Schema != Schema || s.Feed != t.Feed || s.Channel != t.Channel || s.Revision == 0 || !hashRE.MatchString(s.Digest) {
		return s, errors.New("invalid trusted state; do not silently reset it")
	}
	return s, nil
}
func Accept(root string, v VerifiedManifest) error {
	return AtomicJSON(filepath.Join(root, "distribution", "trusted-state.json"), v.State())
}
func ReadOffline(t Trust, root, dir string) (VerifiedManifest, error) {
	prior, e := LoadState(root, t)
	if e != nil {
		return VerifiedManifest{}, e
	}
	paths := []string{filepath.Join(dir, "manifest.json"), filepath.Join(dir, "manifest.sig")}
	var b [2][]byte
	for k, p := range paths {
		if e = NoLinks(p); e != nil {
			return VerifiedManifest{}, e
		}
		st, e := os.Stat(p)
		if e != nil {
			return VerifiedManifest{}, e
		}
		if st.Size() > MaxMetadata {
			return VerifiedManifest{}, errors.New("offline metadata too large")
		}
		b[k], e = os.ReadFile(p)
		if e != nil {
			return VerifiedManifest{}, e
		}
	}
	return t.VerifyManifest(b[0], b[1], time.Now(), prior, true)
}
func CacheOffline(t Trust, p Package, dir, cache string) (string, error) {
	if e := t.VerifyPackage(p); e != nil {
		return "", e
	}
	if strings.ContainsAny(p.Filename, "/\\") {
		return "", errors.New("invalid filename")
	}
	dest := filepath.Join(cache, p.SHA256+".zip")
	return dest, CopyVerified(filepath.Join(dir, p.Filename), dest, p)
}
