package main

import (
	"context"
	"errors"
	"fmt"
	"moo2manager/shared/buildconfig"
	d "moo2manager/shared/distribution"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func (m *Manager) distributionRoot() string {
	if m.AppRoot != "" {
		return m.AppRoot
	}
	return filepath.Join(m.Data, "application")
}
func (m *Manager) distributionStatus() map[string]any {
	t, e := buildconfig.Trust()
	r := map[string]any{"schema": 1, "primary": "github", "secondary": "cloudflare-r2", "website": "neocities (not an update dependency)", "sourceforge": "deferred", "production_feed_configured": false, "app_root": m.distributionRoot(), "live_prsl": false, "automatic_updates": false}
	if e != nil {
		r["error"] = e.Error()
		return r
	}
	r["feed"] = t.Feed
	r["channel"] = t.Channel
	r["development_trust"] = t.Development
	r["key_count"] = len(t.Keys)
	r["production_feed_configured"] = !t.Development && len(t.Endpoints) > 0 && t.Validate() == nil
	if s, e := d.LoadState(m.distributionRoot(), t); e == nil {
		r["trusted_revision"] = s.Revision
	} else {
		r["state_error"] = e.Error()
	}
	r["self_update"] = "Explicit check/stage, then close launcher and run Setup --command apply-staged. No replacement during game play."
	r["base_source"] = "Import the recognized owned archive or fingerprint-verified DOS 1.31 folder; other editions may require support."
	return r
}
func (m *Manager) importPayload(kind, source string) (any, error) {
	if kind != "base" && kind != "patch" {
		return nil, errors.New("choose base or patch")
	}
	source = strings.Trim(strings.TrimSpace(source), "\"")
	if !filepath.IsAbs(source) {
		return nil, errors.New("supply an absolute local archive path")
	}
	if e := d.NoLinks(source); e != nil {
		return nil, e
	}
	name, want := "base.zip", BaseHash
	if kind == "patch" {
		name, want = "patch-1.50.26.zip", PatchHash
	}
	h, size, e := d.HashFile(source)
	if e != nil {
		return nil, e
	}
	if h != want {
		return nil, fmt.Errorf("unrecognized %s archive; this version accepts only the exact previously supplied baseline (SHA-256 %s). Source not modified", kind, want)
	}
	m.progress("Copying your verified local archive into the private cache; original remains untouched…")
	dest := filepath.Join(m.Data, "cache", name)
	if e = d.CopyVerified(source, dest, d.Package{SHA256: want, Size: size}); e != nil {
		return nil, e
	}
	if kind == "base" {
		if e = os.Remove(filepath.Join(m.Data, "cache", "base-source.json")); e != nil && !os.IsNotExist(e) {
			return nil, e
		}
	}
	return map[string]any{"cached": dest, "kind": kind, "sha256": want, "size": size, "source_modified": false, "uploaded": false}, nil
}
func (m *Manager) launcherRelease(offline string, stage bool) (any, error) {
	trust, e := buildconfig.Trust()
	if e != nil {
		return nil, e
	}
	root := m.distributionRoot()
	prior, e := d.LoadState(root, trust)
	if e != nil {
		return nil, e
	}
	client := d.NewClient(trust)
	client.Log = func(f d.Failure) { m.progress(f.Provider + ": " + f.Kind + " — " + f.Message) }
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	var v d.VerifiedManifest
	if strings.TrimSpace(offline) != "" {
		offline = strings.Trim(strings.TrimSpace(offline), "\"")
		if !filepath.IsAbs(offline) {
			return nil, errors.New("offline release folder must be an absolute path")
		}
		v, e = d.ReadOffline(trust, root, offline)
	} else {
		v, e = client.FetchManifest(ctx, prior, time.Now())
	}
	if e != nil {
		return nil, e
	}
	if e = d.Accept(root, v); e != nil {
		return nil, e
	}
	p, e := v.Select("launcher", runtime.GOOS+"-"+runtime.GOARCH)
	if e != nil {
		return nil, e
	}
	result := map[string]any{"release": p.Version, "revision": v.Manifest.Revision, "sha256": p.SHA256, "size": p.Size, "activated": false, "development_trust": trust.Development}
	if !stage {
		return result, nil
	}
	cache := filepath.Join(root, "cache", "packages")
	var file string
	if offline != "" {
		file, e = d.CacheOffline(trust, p, offline, cache)
	} else {
		file, e = client.Download(ctx, p, cache)
	}
	if e != nil {
		return nil, e
	}
	// Pending is an offline signed release directory consumed by the bootstrapper.
	// Publish metadata last; an interruption cannot authorize partial new bytes.
	pending := filepath.Join(root, "distribution", "pending")
	if e = d.CopyVerified(file, filepath.Join(pending, p.Filename), p); e != nil {
		return nil, e
	}
	if e = d.AtomicWrite(filepath.Join(pending, "manifest.json"), v.Bytes); e != nil {
		return nil, e
	}
	if e = d.AtomicWrite(filepath.Join(pending, "manifest.sig"), v.SignatureBytes); e != nil {
		return nil, e
	}
	result["staged"] = pending
	result["next_step"] = "Exit the game and launcher, then run MOO2-SGC-Setup --command apply-staged --install-root with the app_root shown above."
	return result, nil
}
