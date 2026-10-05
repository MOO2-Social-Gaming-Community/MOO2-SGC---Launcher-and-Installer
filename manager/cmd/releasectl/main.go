// Release tooling is maintainer-only and not included in downloaded launchers.
package main

import (
	"archive/zip"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	d "moo2manager/shared/distribution"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	op := flag.String("command", "verify", "keygen, sign, or verify")
	trustPath := flag.String("trust", "", "public trust configuration")
	keyPath := flag.String("key", "", "Ed25519 private-key path OUTSIDE repository/artifact folder")
	keyID := flag.String("key-id", "maintainer-1", "key identifier")
	dir := flag.String("dir", ".", "release staging directory")
	flag.Parse()
	switch *op {
	case "keygen":
		if *keyPath == "" {
			return errors.New("--key is required")
		}
		pub, key, e := ed25519.GenerateKey(rand.Reader)
		if e != nil {
			return e
		}
		f, e := os.OpenFile(*keyPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			return e
		}
		_, e = f.WriteString(hex.EncodeToString(key))
		ce := f.Close()
		if e != nil {
			return e
		}
		if ce != nil {
			return ce
		}
		fmt.Printf("Public key %s: %s\n", *keyID, hex.EncodeToString(pub))
		return nil
	case "sign", "verify":
		var trust d.Trust
		if e := d.ReadJSON(*trustPath, &trust); e != nil {
			return e
		}
		if e := trust.Validate(); e != nil {
			return e
		}
		if *op == "sign" {
			var m d.Manifest
			if e := d.ReadJSON(filepath.Join(*dir, "manifest.draft.json"), &m); e != nil {
				return e
			}
			keyAbs, e := filepath.Abs(*keyPath)
			if e != nil {
				return e
			}
			dirAbs, e := filepath.Abs(*dir)
			if e != nil {
				return e
			}
			rel, e := filepath.Rel(dirAbs, keyAbs)
			if e != nil {
				return e
			}
			if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return errors.New("private signing key must not live in release directory")
			}
			b, e := os.ReadFile(keyAbs)
			if e != nil {
				return e
			}
			key, e := hex.DecodeString(strings.TrimSpace(string(b)))
			if e != nil || len(key) != 64 {
				return errors.New("invalid Ed25519 private key")
			}
			priv := ed25519.PrivateKey(key)
			if trust.Keys[*keyID] != hex.EncodeToString(priv.Public().(ed25519.PublicKey)) {
				return errors.New("private key does not match public trust configuration")
			}
			for x, p := range m.Packages {
				if _, e := d.SafePath(p.Filename); e != nil || strings.ContainsAny(p.Filename, "/\\") {
					return errors.New("unsafe artifact filename")
				}
				file := filepath.Join(*dir, p.Filename)
				hash, size, e := d.HashFile(file)
				if e != nil {
					return e
				}
				p.SHA256 = hash
				p.Size = size
				if p.Kind == "launcher" {
					p.Files, e = indexZIP(file)
					if e != nil {
						return e
					}
				}
				p.Signature = d.Sign(d.PackageMessage(p), *keyID, priv)
				m.Packages[x] = p
			}
			b, e = json.MarshalIndent(m, "", "  ")
			if e != nil {
				return e
			}
			b = append(b, '\n')
			sig := d.SignManifest(b, *keyID, priv)
			sb, _ := json.MarshalIndent(sig, "", "  ")
			sb = append(sb, '\n')
			if _, e = trust.VerifyManifest(b, sb, time.Now(), d.TrustedState{}, true); e != nil {
				return e
			}
			if e = d.AtomicWrite(filepath.Join(*dir, "manifest.json"), b); e != nil {
				return e
			}
			if e = d.AtomicWrite(filepath.Join(*dir, "manifest.sig"), sb); e != nil {
				return e
			}
		}
		v, e := d.ReadOffline(trust, filepath.Join(*dir, ".verification-state-unused"), *dir)
		if e != nil {
			return e
		}
		for _, p := range v.Manifest.Packages {
			if e = d.VerifyFile(filepath.Join(*dir, p.Filename), p); e != nil {
				return e
			}
		}
		fmt.Printf("Verified signed manifest revision %d and %d packages.\n", v.Manifest.Revision, len(v.Manifest.Packages))
		return nil
	default:
		return errors.New("unknown command")
	}
}
func indexZIP(p string) (map[string]d.FileRecord, error) {
	z, e := zip.OpenReader(p)
	if e != nil {
		return nil, e
	}
	defer z.Close()
	if len(z.File) > d.MaxEntries {
		return nil, errors.New("too many files")
	}
	out := map[string]d.FileRecord{}
	seen := map[string]bool{}
	var total uint64
	for _, f := range z.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if _, e := d.SafePath(f.Name); e != nil {
			return nil, e
		}
		lower := strings.ToLower(f.Name)
		if seen[lower] || !f.Mode().IsRegular() || f.Mode()&os.ModeSymlink != 0 || f.Flags&1 != 0 {
			return nil, errors.New("invalid ZIP entry")
		}
		seen[lower] = true
		// Release safety fence: launcher packages never contain commercial data,
		// private keys, uploaded payload ZIPs, or credential files.
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext == ".lbx" || ext == ".gam" || ext == ".rac" || ext == ".key" || ext == ".pem" || ext == ".zip" || strings.Contains(lower, ".env") || strings.Contains(lower, "payloads/") || strings.Contains(lower, "orion2.exe") || strings.Contains(lower, "orion150.exe") {
			return nil, fmt.Errorf("forbidden launcher artifact entry: %s", f.Name)
		}
		if f.UncompressedSize64 > d.MaxExtracted-total {
			return nil, errors.New("ZIP expansion limit")
		}
		total += f.UncompressedSize64
		in, e := f.Open()
		if e != nil {
			return nil, e
		}
		h := sha256.New()
		n, e := io.Copy(h, io.LimitReader(in, int64(f.UncompressedSize64)+1))
		in.Close()
		if e != nil {
			return nil, e
		}
		if n != int64(f.UncompressedSize64) {
			return nil, errors.New("ZIP size mismatch")
		}
		out[f.Name] = d.FileRecord{SHA256: hex.EncodeToString(h.Sum(nil)), Size: n}
	}
	return out, nil
}
