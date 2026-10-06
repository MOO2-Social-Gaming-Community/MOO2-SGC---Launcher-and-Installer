package main

// Source identity is established from DOS engine AND content hashes. Store
// filenames, readmes and archive names are not trusted version detectors.
import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type SourceEdition struct {
	EngineFile string     `json:"engine_file,omitempty"`
	Harness    bool       `json:"harness,omitempty"`
	ID         string     `json:"id"`
	Version    string     `json:"version"`
	Label      string     `json:"label"`
	Files      []BaseFile `json:"files"`
}

func sourceEditions() []SourceEdition {
	b, e := resources.ReadFile("assets/source-editions.json")
	if e != nil {
		panic(e)
	}
	var v struct {
		Schema   int             `json:"schema"`
		Editions []SourceEdition `json:"editions"`
	}
	if e = json.Unmarshal(b, &v); e != nil || v.Schema != 1 {
		panic("invalid compiled edition catalog")
	}
	return v.Editions
}
func editionByID(id string) (SourceEdition, error) {
	for _, s := range sourceEditions() {
		if s.ID == id || s.ID == "legacy-en-131" && s.Label == id {
			return s, nil
		}
	}
	return SourceEdition{}, fmt.Errorf("unsupported source edition %q", id)
}
func editionByEngine(hash string) (SourceEdition, error) {
	for _, s := range sourceEditions() {
		for _, f := range s.Files {
			if f.Name == "ORION2.EXE" && s.EngineFile == "" && f.SHA256 == hash {
				return s, nil
			}
		}
	}
	return SourceEdition{}, errors.New("unrecognized DOS ORION2.EXE: supported sources are the fingerprinted English CD 1.2, Steam 1.40b23, and legacy DOS 1.31. Source untouched")
}
func (m *Manager) currentEdition() (SourceEdition, error) {
	var r ImportedBase
	e := readJSON(filepath.Join(m.Data, "cache", "base-source.json"), &r)
	if os.IsNotExist(e) {
		return editionByID("legacy-en-131")
	}
	if e != nil {
		return SourceEdition{}, e
	}
	return editionByID(r.Edition)
}
func detectFolderEdition(root string) (SourceEdition, error) {
	files, e := folderFiles(root)
	if e != nil {
		return SourceEdition{}, e
	}
	// Manual patches retain ORION2.EXE at 1.31; inspect the real b23 output first.
	if p, ok := files["ORION2V140.EXE"]; ok {
		if e := noSymlinkAncestors(p); e != nil {
			return SourceEdition{}, e
		}
		if e := regularFile(p); e != nil {
			return SourceEdition{}, e
		}
		if st, e := os.Stat(p); e == nil && st.Size() <= 8<<20 {
			if h, e := hashFile(p); e == nil && h == BaselineEngineHash {
				return editionByID("manual-en-140b23")
			}
		}
	}
	p, ok := files["ORION2.EXE"]
	if !ok {
		return SourceEdition{}, errors.New("select the directory containing DOS ORION2.EXE and the LBX game files")
	}
	if e = noSymlinkAncestors(p); e != nil {
		return SourceEdition{}, e
	}
	if e = regularFile(p); e != nil {
		return SourceEdition{}, e
	}
	info, e := os.Stat(p)
	if e != nil {
		return SourceEdition{}, e
	}
	if info.Size() > 8<<20 {
		return SourceEdition{}, errors.New("oversized DOS executable")
	}
	h, e := hashFile(p)
	if e != nil {
		return SourceEdition{}, e
	}
	return editionByEngine(h)
}

// Validate every archive path, but copy ONLY the explicitly fingerprinted files.
// Never run archived installers, links, scripts, wrappers, or DOS setup programs.
func zipSourceFiles(z *zip.ReadCloser) (map[string]*zip.File, error) {
	files := map[string]*zip.File{}
	if len(z.File) > 15000 {
		return nil, errors.New("too many archive entries")
	}
	var total uint64
	for _, f := range z.File {
		if _, e := safeRel(f.Name); e != nil {
			return nil, e
		}
		if f.Mode()&os.ModeSymlink != 0 || f.Flags&1 != 0 {
			return nil, errors.New("archive links/encryption are not supported")
		}
		if f.FileInfo().IsDir() {
			continue
		}
		if !f.Mode().IsRegular() {
			return nil, errors.New("nonregular archive member")
		}
		const max uint64 = 2 << 30
		if f.UncompressedSize64 > max || total > max-f.UncompressedSize64 {
			return nil, errors.New("archive exceeds 2 GiB expanded budget")
		}
		total += f.UncompressedSize64
		key := strings.ToUpper(f.Name)
		if _, ok := files[key]; ok {
			return nil, errors.New("duplicate or case-colliding archive member")
		}
		files[key] = f
	}
	return files, nil
}
func zipRoot(files map[string]*zip.File, required string) (string, error) {
	var root string
	count := 0
	for n := range files {
		if path.Base(n) == required {
			root = strings.TrimSuffix(n, required)
			count++
		}
	}
	if count != 1 {
		return "", fmt.Errorf("expected one unambiguous %s in the archive; found %d", required, count)
	}
	return root, nil
}
func readZipKnown(f *zip.File, limit int64) ([]byte, error) {
	if f == nil {
		return nil, errors.New("missing archive member")
	}
	if f.UncompressedSize64 > uint64(limit) {
		return nil, errors.New("oversized archive member")
	}
	r, e := f.Open()
	if e != nil {
		return nil, e
	}
	defer r.Close()
	b, e := io.ReadAll(io.LimitReader(r, limit+1))
	if int64(len(b)) > limit {
		return nil, errors.New("archive member size exceeded")
	}
	return b, e
}
func detectArchiveEdition(archive string) (SourceEdition, error) {
	if e := regularFile(archive); e != nil {
		return SourceEdition{}, e
	}
	z, e := zip.OpenReader(archive)
	if e != nil {
		return SourceEdition{}, e
	}
	defer z.Close()
	fs, e := zipSourceFiles(z)
	if e != nil {
		return SourceEdition{}, e
	}
	root, e := zipRoot(fs, "ORION2.EXE")
	if e != nil {
		return SourceEdition{}, e
	}
	if f := fs[root+"ORION2V140.EXE"]; f != nil {
		b, e := readZipKnown(f, 8<<20)
		if e != nil {
			return SourceEdition{}, e
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) == BaselineEngineHash {
			return editionByID("manual-en-140b23")
		}
	}
	b, e := readZipKnown(fs[root+"ORION2.EXE"], 8<<20)
	if e != nil {
		return SourceEdition{}, e
	}
	h := sha256.Sum256(b)
	return editionByEngine(hex.EncodeToString(h[:]))
}
func writeZipSnapshot(archive, dest string, records []BaseFile) (err error) {
	if e := noSymlinkAncestors(archive); e != nil {
		return e
	}
	if e := noSymlinkAncestors(dest); e != nil {
		return e
	}
	if e := regularFile(archive); e != nil {
		return e
	}
	z, e := zip.OpenReader(archive)
	if e != nil {
		return e
	}
	defer z.Close()
	fs, e := zipSourceFiles(z)
	if e != nil {
		return e
	}
	root, e := zipRoot(fs, "ORION2.EXE")
	if e != nil {
		return e
	}
	out, e := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer func() {
		out.Close()
		if err != nil {
			os.Remove(dest)
		}
	}()
	w := zip.NewWriter(out)
	defer w.Close()
	for _, rec := range records {
		f, ok := fs[root+rec.Name]
		if !ok {
			return fmt.Errorf("supported source requires %s; original untouched", rec.Name)
		}
		if f.UncompressedSize64 != uint64(rec.Size) {
			return fmt.Errorf("unsupported source data variant: %s (size); original untouched", rec.Name)
		}
		in, e := f.Open()
		if e != nil {
			return e
		}
		hdr := &zip.FileHeader{Name: "base/" + rec.Name, Method: zip.Store}
		hdr.SetMode(0600)
		hdr.SetModTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
		dst, e := w.CreateHeader(hdr)
		if e != nil {
			in.Close()
			return e
		}
		h := sha256.New()
		n, e := io.Copy(io.MultiWriter(dst, h), io.LimitReader(in, rec.Size+1))
		in.Close()
		if e != nil {
			return e
		}
		if n != rec.Size || hex.EncodeToString(h.Sum(nil)) != rec.SHA256 {
			return fmt.Errorf("unsupported source data variant: %s (SHA-256); recognized engine is not enough to certify these assets. Original untouched", rec.Name)
		}
	}
	if e = w.Close(); e != nil {
		return e
	}
	if e = out.Sync(); e != nil {
		return e
	}
	return out.Close()
}
func (m *Manager) commitSourceSnapshot(staged string, s SourceEdition) (any, error) {
	hash, e := hashFile(staged)
	if e != nil {
		return nil, e
	}
	name := "base-import-" + hash + ".zip"
	dest := filepath.Join(m.Data, "cache", name)
	if e = requireHash(dest, hash); e != nil {
		if e = replaceFile(staged, dest); e != nil {
			return nil, e
		}
	}
	record := ImportedBase{2, name, hash, len(s.Files), s.ID}
	if e = atomicJSON(filepath.Join(m.Data, "cache", "base-source-"+s.ID+".json"), record); e != nil {
		return nil, e
	}
	if e = atomicJSON(filepath.Join(m.Data, "cache", "base-source.json"), record); e != nil {
		return nil, e
	}
	return map[string]any{"cached": dest, "files": len(s.Files), "source_modified": false, "uploaded": false, "sha256": hash, "edition": s.Label, "source_version": s.Version, "effective_baseline": "1.40b23"}, nil
}
func (m *Manager) importKnownZip(source string) (any, error) {
	p, e := localPath(source)
	if e != nil {
		return nil, e
	}
	s, e := detectArchiveEdition(p)
	if e != nil {
		return nil, e
	}
	tmp, e := os.MkdirTemp(filepath.Join(m.Data, "cache"), "import-")
	if e != nil {
		return nil, e
	}
	defer os.RemoveAll(tmp)
	staged := filepath.Join(tmp, "base.zip")
	m.progress("Recognized " + s.Label + ". Verifying owned assets and copying only required DOS files…")
	if e = writeZipSnapshot(p, staged, s.Files); e != nil {
		return nil, e
	}
	return m.commitSourceSnapshot(staged, s)
}
func (m *Manager) sourceStatus() any {
	// Polling the dashboard must not rehash a 200 MB source ZIP every 1.5 seconds.
	var r ImportedBase
	if e := readJSON(filepath.Join(m.Data, "cache", "base-source.json"), &r); e != nil {
		return map[string]string{"status": "import/select an owned source", "effective_baseline": "1.40b23"}
	}
	s, e := editionByID(r.Edition)
	if e != nil {
		return map[string]string{"error": e.Error()}
	}
	return map[string]string{"edition": s.Label, "version": s.Version, "effective_baseline": "1.40b23", "verification": "content verified at import; verified again during preparation"}
}

// Destination is a fresh operation-owned staging directory, never the source.
func extractVerifiedMembers(archive, dest string, records []BaseFile) error {
	z, e := zip.OpenReader(archive)
	if e != nil {
		return e
	}
	defer z.Close()
	fs, e := zipSourceFiles(z)
	if e != nil {
		return e
	}
	root, e := zipRoot(fs, "ORION2.EXE")
	if e != nil {
		return e
	}
	if e = noSymlinkAncestors(dest); e != nil {
		return e
	}
	if e = os.MkdirAll(dest, 0700); e != nil {
		return e
	}
	for _, r := range records {
		if _, e = safeRel(r.Name); e != nil {
			return e
		}
		if strings.ContainsAny(r.Name, "/\\") {
			return errors.New("source allowlist must contain root files")
		}
		f, ok := fs[root+r.Name]
		if !ok || f.UncompressedSize64 != uint64(r.Size) {
			return fmt.Errorf("source file missing or wrong size: %s", r.Name)
		}
		in, e := f.Open()
		if e != nil {
			return e
		}
		p := filepath.Join(dest, r.Name)
		if e = noSymlinkAncestors(p); e != nil {
			in.Close()
			return e
		}
		out, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			in.Close()
			return e
		}
		h := sha256.New()
		n, e := io.Copy(io.MultiWriter(out, h), io.LimitReader(in, r.Size+1))
		in.Close()
		ce := out.Close()
		if e != nil {
			return e
		}
		if ce != nil {
			return ce
		}
		if n != r.Size || hex.EncodeToString(h.Sum(nil)) != r.SHA256 {
			return fmt.Errorf("source file changed during installation: %s", r.Name)
		}
	}
	return nil
}
