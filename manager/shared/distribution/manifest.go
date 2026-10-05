// Package distribution implements the versioned MOO2-SGC release boundary.
// It has no dependency on a game engine, PRSL, or a particular hosting API.
package distribution

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const Schema = 1
const MaxMetadata = 2 << 20
const MaxPackage = 512 << 20
const MaxExtracted = 1 << 30
const MaxEntries = 10000

var tokenRE = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,79}$`)
var hashRE = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Mirror struct {
	Provider string `json:"provider"`
	Priority int    `json:"priority"`
	URL      string `json:"url"`
}
type Endpoint struct {
	Mirror
	SignatureURL string `json:"signature_url"`
}
type Trust struct {
	Schema       int               `json:"schema"`
	Feed         string            `json:"feed"`
	Channel      string            `json:"channel"`
	Development  bool              `json:"development"`
	Keys         map[string]string `json:"keys"` // Ed25519 public keys, hex. NEVER secret keys.
	Endpoints    []Endpoint        `json:"endpoints"`
	AllowedHosts []string          `json:"allowed_hosts"`
}
type Signature struct {
	KeyID string `json:"key_id"`
	Value string `json:"signature"` // base64 Ed25519 signature
}
type Requirement struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}
type Package struct {
	ID             string                `json:"id"`
	Kind           string                `json:"kind"`
	Version        string                `json:"version"`
	Platform       string                `json:"platform"`
	Filename       string                `json:"filename"`
	Format         string                `json:"format"`
	Entrypoint     string                `json:"entrypoint,omitempty"`
	Size           int64                 `json:"size"`
	SHA256         string                `json:"sha256"`
	Signature      Signature             `json:"signature"`
	Mirrors        []Mirror              `json:"mirrors"`
	Requires       []Requirement         `json:"requires,omitempty"`
	EngineHashes   []string              `json:"engine_hashes,omitempty"`
	Files          map[string]FileRecord `json:"files,omitempty"`
	Redistribution string                `json:"redistribution"`
}
type Manifest struct {
	Schema    int       `json:"schema"`
	Feed      string    `json:"feed"`
	Channel   string    `json:"channel"`
	Revision  uint64    `json:"revision"`
	Release   string    `json:"release"`
	Published time.Time `json:"published"`
	Expires   time.Time `json:"expires"`
	Packages  []Package `json:"packages"`
}
type TrustedState struct {
	Schema   int    `json:"schema"`
	Feed     string `json:"feed"`
	Channel  string `json:"channel"`
	Revision uint64 `json:"revision"`
	Digest   string `json:"digest"`
}
type VerifiedManifest struct {
	Manifest       Manifest
	Bytes          []byte
	SignatureBytes []byte
	Digest         string
}

func DecodeStrict(b []byte, v any) error {
	if len(b) > MaxMetadata {
		return errors.New("metadata size limit exceeded")
	}
	// Duplicate keys must not have different meanings to a signer and verifier.
	d := json.NewDecoder(bytes.NewReader(b))
	var scan func() error
	depth := 0
	scan = func() error {
		depth++
		defer func() { depth-- }()
		if depth > 64 {
			return errors.New("JSON nesting limit")
		}
		t, e := d.Token()
		if e != nil {
			return e
		}
		delim, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			keys := map[string]bool{}
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return e
				}
				s, ok := k.(string)
				if !ok || keys[s] {
					return errors.New("duplicate JSON key")
				}
				keys[s] = true
				if e = scan(); e != nil {
					return e
				}
			}
		case '[':
			for d.More() {
				if e := scan(); e != nil {
					return e
				}
			}
		default:
			return errors.New("unexpected JSON delimiter")
		}
		_, e = d.Token()
		return e
	}
	if e := scan(); e != nil {
		return e
	}
	if _, e := d.Token(); e != io.EOF {
		return errors.New("trailing JSON")
	}
	d = json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return e
	}
	return nil
}
func providerOK(p string) bool { return p == "github" || p == "cloudflare-r2" }
func (t Trust) Validate() error {
	if t.Schema != Schema || !tokenRE.MatchString(t.Feed) || !tokenRE.MatchString(t.Channel) {
		return errors.New("unsupported trust configuration")
	}
	if len(t.Keys) == 0 || len(t.Keys) > 8 {
		return errors.New("release trust is not configured: maintainer must embed public signing keys")
	}
	for id, k := range t.Keys {
		b, e := hex.DecodeString(k)
		if !tokenRE.MatchString(id) || e != nil || len(b) != ed25519.PublicKeySize {
			return errors.New("invalid public key")
		}
	}
	if len(t.Endpoints) > 8 || len(t.AllowedHosts) > 32 {
		return errors.New("too many trust endpoints")
	}
	for _, h := range t.AllowedHosts {
		if h == "" || h != strings.ToLower(h) || strings.ContainsAny(h, "/:@?#* ") {
			return errors.New("invalid exact hostname allowlist")
		}
	}
	for _, e := range t.Endpoints {
		if !providerOK(e.Provider) || e.Priority < 0 {
			return errors.New("unsupported metadata provider")
		}
		if err := t.CheckURL(e.URL); err != nil {
			return err
		}
		if err := t.CheckURL(e.SignatureURL); err != nil {
			return err
		}
	}
	return nil
}
func (t Trust) CheckURL(s string) error {
	u, e := url.Parse(s)
	if e != nil || u.Scheme != "https" || u.User != nil || u.Host == "" || u.Fragment != "" || len(s) > 4096 || (u.Port() != "" && u.Port() != "443" && !(t.Development && net.ParseIP(u.Hostname()) != nil && net.ParseIP(u.Hostname()).IsLoopback())) {
		return errors.New("only approved HTTPS release URLs are accepted")
	}
	for _, h := range t.AllowedHosts {
		if u.Hostname() == h {
			return nil
		}
	}
	return fmt.Errorf("release hostname not approved: %s", u.Hostname())
}
func (t Trust) verify(message []byte, s Signature) error {
	k, ok := t.Keys[s.KeyID]
	if !ok {
		return errors.New("untrusted signing key")
	}
	pub, e := hex.DecodeString(k)
	if e != nil || len(pub) != 32 {
		return errors.New("invalid public key")
	}
	sig, e := base64.StdEncoding.DecodeString(s.Value)
	if e != nil || len(sig) != 64 || !ed25519.Verify(pub, message, sig) {
		return errors.New("signature mismatch")
	}
	return nil
}

// PackageMessage signs content identity plus execution semantics, not a mutable
// download URL. SHA-256 covers EVERY byte of the ZIP. Manifest signs mirrors.
func PackageMessage(p Package) []byte {
	v := struct {
		ID, Kind, Version, Platform, Filename, Format, Entrypoint, SHA256 string
		Size                                                              int64
		Files                                                             map[string]FileRecord
	}{p.ID, p.Kind, p.Version, p.Platform, p.Filename, p.Format, p.Entrypoint, p.SHA256, p.Size, p.Files}
	b, _ := json.Marshal(v)
	return append([]byte("MOO2-SGC-PACKAGE-V1\x00"), b...)
}
func (t Trust) VerifyPackage(p Package) error {
	if !tokenRE.MatchString(p.ID) || !tokenRE.MatchString(p.Version) || !hashRE.MatchString(p.SHA256) || p.Size < 1 || p.Size > MaxPackage {
		return errors.New("invalid package identity/size")
	}
	switch p.Kind {
	case "launcher", "patch", "mod", "configuration", "resources", "documentation":
	default:
		return errors.New("unsupported package kind (commercial base-game distribution is forbidden)")
	}
	switch p.Platform {
	case "windows-amd64", "linux-amd64", "darwin-amd64", "darwin-arm64", "any":
	default:
		return errors.New("unsupported platform")
	}
	if p.Format != "zip" {
		return errors.New("unsupported package format")
	}
	if _, e := SafePath(p.Filename); e != nil || strings.Contains(p.Filename, "/") {
		return errors.New("invalid package filename")
	}
	if p.Kind == "launcher" && (p.ID != "launcher" || p.Platform == "any" || p.Entrypoint == "") {
		return errors.New("launcher must identify its native entrypoint/platform")
	}
	if p.Entrypoint != "" {
		if _, e := SafePath(p.Entrypoint); e != nil {
			return e
		}
		if p.Kind != "launcher" {
			return errors.New("non-launcher package may not declare an executable entrypoint")
		}
	}
	if len(p.Files) > MaxEntries {
		return errors.New("file index limit")
	}
	var expanded int64
	for n, r := range p.Files {
		if _, e := SafePath(n); e != nil {
			return e
		}
		if !hashRE.MatchString(r.SHA256) || r.Size < 0 || r.Size > MaxExtracted-expanded {
			return errors.New("invalid signed file index")
		}
		expanded += r.Size
	}
	if p.Kind == "launcher" {
		if _, ok := p.Files[p.Entrypoint]; !ok {
			return errors.New("launcher entrypoint needs a signed file record")
		}
	}
	if p.Redistribution != "project-owned" && p.Redistribution != "permission-recorded" {
		return errors.New("package redistribution rights have not been recorded")
	}
	if len(p.Mirrors) > 8 || len(p.Requires) > 32 || len(p.EngineHashes) > 16 {
		return errors.New("package limit exceeded")
	}
	for _, r := range p.Requires {
		if !tokenRE.MatchString(r.ID) || !tokenRE.MatchString(r.Version) {
			return errors.New("invalid dependency")
		}
	}
	for _, h := range p.EngineHashes {
		if !hashRE.MatchString(h) {
			return errors.New("invalid supported-engine hash")
		}
	}
	for _, m := range p.Mirrors {
		if !providerOK(m.Provider) || m.Priority < 0 {
			return errors.New("unsupported mirror")
		}
		if e := t.CheckURL(m.URL); e != nil {
			return e
		}
	}
	return t.verify(PackageMessage(p), p.Signature)
}
func (t Trust) VerifyManifest(b, sigBytes []byte, now time.Time, prior TrustedState, checkExpiry bool) (VerifiedManifest, error) {
	v := VerifiedManifest{Bytes: append([]byte{}, b...), SignatureBytes: append([]byte{}, sigBytes...)}
	if e := t.Validate(); e != nil {
		return v, e
	}
	if len(b) > MaxMetadata {
		return v, errors.New("manifest too large")
	}
	var s Signature
	if e := DecodeStrict(sigBytes, &s); e != nil {
		return v, e
	}
	if e := t.verify(append([]byte("MOO2-SGC-MANIFEST-V1\x00"), b...), s); e != nil {
		return v, e
	}
	if e := DecodeStrict(b, &v.Manifest); e != nil {
		return v, e
	}
	m := v.Manifest
	if m.Schema != Schema || m.Feed != t.Feed || m.Channel != t.Channel || m.Revision == 0 || !tokenRE.MatchString(m.Release) {
		return v, errors.New("unsupported schema, feed, channel, or revision")
	}
	if len(m.Packages) < 1 || len(m.Packages) > 256 {
		return v, errors.New("invalid package count")
	}
	if !m.Expires.After(m.Published) || m.Expires.Sub(m.Published) > 93*24*time.Hour {
		return v, errors.New("invalid metadata lifetime (maximum 93 days)")
	}
	if checkExpiry && (m.Published.After(now.Add(5*time.Minute)) || !m.Expires.After(now)) {
		return v, errors.New("metadata expired or published in the future; check the system clock")
	}
	h := sha256.Sum256(b)
	v.Digest = hex.EncodeToString(h[:])
	if prior.Revision > 0 {
		if prior.Schema != Schema || prior.Feed != t.Feed || prior.Channel != t.Channel {
			return v, errors.New("trusted revision state belongs to a different feed")
		}
		if m.Revision < prior.Revision {
			return v, errors.New("manifest rollback refused")
		}
		if m.Revision == prior.Revision && v.Digest != prior.Digest {
			return v, errors.New("same-revision manifest equivocation refused")
		}
	}
	seen := map[string]bool{}
	for _, p := range m.Packages {
		key := p.ID + "/" + p.Version + "/" + p.Platform
		if seen[key] {
			return v, errors.New("duplicate package identity")
		}
		seen[key] = true
		if e := t.VerifyPackage(p); e != nil {
			return v, fmt.Errorf("package %s: %w", p.ID, e)
		}
	}
	return v, nil
}
func (v VerifiedManifest) State() TrustedState {
	return TrustedState{Schema, v.Manifest.Feed, v.Manifest.Channel, v.Manifest.Revision, v.Digest}
}
func (v VerifiedManifest) Select(id, platform string) (Package, error) {
	var found *Package
	for _, p := range v.Manifest.Packages {
		if p.ID == id && (p.Platform == platform || p.Platform == "any") {
			if found != nil {
				return Package{}, errors.New("ambiguous package selection")
			}
			q := p
			found = &q
		}
	}
	if found == nil {
		return Package{}, fmt.Errorf("no signed package for %s / %s", id, platform)
	}
	return *found, nil
}
func Sign(message []byte, id string, key ed25519.PrivateKey) Signature {
	return Signature{id, base64.StdEncoding.EncodeToString(ed25519.Sign(key, message))}
}
func SignManifest(b []byte, id string, key ed25519.PrivateKey) Signature {
	return Sign(append([]byte("MOO2-SGC-MANIFEST-V1\x00"), b...), id, key)
}
