package distribution

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
	"strings"
)

var reservedRE = regexp.MustCompile(`^(COM|LPT)[1-9]$`)

func SafePath(n string) (string, error) {
	if n == "" || strings.ContainsAny(n, "\\:\x00\r\n\t\"*?<>|") || strings.HasPrefix(n, "/") || len(n) > 240 {
		return "", errors.New("unsafe package path")
	}
	for _, r := range n {
		if r < 32 {
			return "", errors.New("control in path")
		}
	}
	for _, s := range strings.Split(n, "/") {
		if s == "" || s == "." || s == ".." || strings.HasSuffix(s, ".") || strings.HasSuffix(s, " ") {
			return "", errors.New("unsafe path component")
		}
		stem := strings.ToUpper(strings.Split(s, ".")[0])
		if stem == "CON" || stem == "PRN" || stem == "AUX" || stem == "NUL" || reservedRE.MatchString(stem) {
			return "", errors.New("reserved filename")
		}
	}
	return filepath.FromSlash(n), nil
}
func NoLinks(p string) error {
	for {
		v, e := os.Lstat(p)
		if e == nil && v.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink refused: %s", p)
		}
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		n := filepath.Dir(p)
		if n == p {
			return nil
		}
		p = n
	}
}
func HashFile(p string) (string, int64, error) {
	if e := NoLinks(p); e != nil {
		return "", 0, e
	}
	f, e := os.Open(p)
	if e != nil {
		return "", 0, e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil || !st.Mode().IsRegular() {
		return "", 0, errors.New("not a regular file")
	}
	h := sha256.New()
	n, e := io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), n, e
}
func VerifyFile(p string, pkg Package) error {
	h, n, e := HashFile(p)
	if e != nil {
		return e
	}
	if n != pkg.Size {
		return errors.New("byte count mismatch")
	}
	if h != pkg.SHA256 {
		return errors.New("SHA-256 mismatch")
	}
	return nil
}
func AtomicJSON(p string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return AtomicWrite(p, append(b, '\n'))
}
func AtomicWrite(p string, b []byte) error {
	if e := NoLinks(p); e != nil {
		return e
	}
	if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(p), ".write-")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	return replace(f.Name(), p)
}
func ReadJSON(p string, v any) error {
	if e := NoLinks(p); e != nil {
		return e
	}
	f, e := os.Open(p)
	if e != nil {
		return e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, MaxMetadata+1))
	if e != nil {
		return e
	}
	return DecodeStrict(b, v)
}
func CopyVerified(src, dest string, p Package) error {
	if e := VerifyFile(src, p); e != nil {
		return e
	}
	if e := NoLinks(dest); e != nil {
		return e
	}
	if e := os.MkdirAll(filepath.Dir(dest), 0700); e != nil {
		return e
	}
	in, e := os.Open(src)
	if e != nil {
		return e
	}
	defer in.Close()
	out, e := os.CreateTemp(filepath.Dir(dest), ".copy-")
	if e != nil {
		return e
	}
	tmp := out.Name()
	defer os.Remove(tmp)
	_, e = io.Copy(out, io.LimitReader(in, p.Size+1))
	if e != nil {
		out.Close()
		return e
	}
	if e = out.Sync(); e != nil {
		out.Close()
		return e
	}
	if e = out.Close(); e != nil {
		return e
	}
	if e = VerifyFile(tmp, p); e != nil {
		return e
	}
	return replace(tmp, dest)
}

type FileRecord struct {
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

func Extract(archive, dest string) (map[string]FileRecord, error) {
	if e := NoLinks(dest); e != nil {
		return nil, e
	}
	z, e := zip.OpenReader(archive)
	if e != nil {
		return nil, e
	}
	defer z.Close()
	if len(z.File) == 0 || len(z.File) > MaxEntries {
		return nil, errors.New("ZIP entry limit")
	}
	seen := map[string]bool{}
	var total uint64
	for _, f := range z.File {
		n := strings.TrimSuffix(f.Name, "/")
		if _, e := SafePath(n); e != nil {
			return nil, e
		}
		if f.Mode()&os.ModeSymlink != 0 || f.Flags&1 != 0 {
			return nil, errors.New("symlink/encrypted ZIP entry refused")
		}
		if !f.FileInfo().IsDir() && !f.Mode().IsRegular() {
			return nil, errors.New("special archive entry refused")
		}
		key := strings.ToLower(n)
		if seen[key] {
			return nil, errors.New("duplicate/case-colliding ZIP entry")
		}
		seen[key] = true
		if f.UncompressedSize64 > MaxExtracted-total {
			return nil, errors.New("ZIP expansion limit")
		}
		total += f.UncompressedSize64
	}
	if e = os.MkdirAll(dest, 0700); e != nil {
		return nil, e
	}
	records := map[string]FileRecord{}
	for _, f := range z.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rel, _ := SafePath(f.Name)
		target := filepath.Join(dest, rel)
		if e = NoLinks(target); e != nil {
			return nil, e
		}
		if e = os.MkdirAll(filepath.Dir(target), 0700); e != nil {
			return nil, e
		}
		in, e := f.Open()
		if e != nil {
			return nil, e
		}
		out, e := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			in.Close()
			return nil, e
		}
		h := sha256.New()
		n, e := io.Copy(io.MultiWriter(out, h), io.LimitReader(in, int64(f.UncompressedSize64)+1))
		in.Close()
		se := out.Sync()
		ce := out.Close()
		if e != nil {
			return nil, e
		}
		if se != nil {
			return nil, se
		}
		if ce != nil {
			return nil, ce
		}
		if n != int64(f.UncompressedSize64) {
			return nil, errors.New("extracted byte count mismatch")
		}
		records[f.Name] = FileRecord{hex.EncodeToString(h.Sum(nil)), n}
	}
	if len(records) == 0 {
		return nil, errors.New("empty payload")
	}
	return records, nil
}
func Acquire(root string) (func(), error) {
	p := filepath.Join(root, "installation.lock")
	if e := NoLinks(p); e != nil {
		return nil, e
	}
	if e := os.MkdirAll(root, 0700); e != nil {
		return nil, e
	}
	f, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return nil, fmt.Errorf("installation/launcher is locked; close the launcher and game first; after a crash verify both stopped before removing %s: %w", p, e)
	}
	fmt.Fprintln(f, os.Getpid())
	f.Close()
	return func() { _ = os.Remove(p) }, nil
}
