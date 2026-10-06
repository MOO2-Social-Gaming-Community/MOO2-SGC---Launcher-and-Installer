package main

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeKernelZip(t *testing.T, names []string, content []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "rkernel.zip")
	f, e := os.Create(p)
	if e != nil {
		t.Fatal(e)
	}
	z := zip.NewWriter(f)
	for _, n := range names {
		w, e := z.Create(n)
		if e != nil {
			t.Fatal(e)
		}
		w.Write(content)
	}
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	f.Close()
	return p
}
func TestOwnedKernelRejectsAmbiguityTamperAndTraversal(t *testing.T) {
	for _, c := range []struct {
		name  string
		names []string
		b     []byte
	}{{"unknown", []string{"RKERNEL.COM"}, []byte("unrecognized")}, {"duplicate", []string{"RKERNEL.COM", "other/RKERNEL.COM"}, []byte("bad")}, {"traversal", []string{"../RKERNEL.COM"}, []byte("bad")}, {"oversize", []string{"RKERNEL.COM"}, make([]byte, kernelSize+1)}, {"exe-not-com", []string{"RKERNEL.EXE"}, []byte("bad")}} {
		t.Run(c.name, func(t *testing.T) {
			m := distributionTestManager(t)
			p := makeKernelZip(t, c.names, c.b)
			before, _ := os.ReadFile(p)
			if _, e := m.importKernel(p); e == nil {
				t.Fatal("invalid driver accepted")
			}
			after, _ := os.ReadFile(p)
			if !bytes.Equal(before, after) {
				t.Fatal("source changed")
			}
			if exists(m.kernelCache()) {
				t.Fatal("invalid cache activated")
			}
		})
	}
}
func TestMissingKernelPreflightHasActualGamePath(t *testing.T) {
	game := filepath.Join(t.TempDir(), "environment with spaces", "game")
	os.MkdirAll(game, 0700)
	p := inspectNetworkKernel(game, "1.50.26")
	if p.OK || !p.Required || p.KernelPath != filepath.Join(game, "RKERNEL.COM") {
		t.Fatal(p)
	}
	if !inspectNetworkKernel(game, "1.2").OK {
		t.Fatal("original CD changed")
	}
	os.WriteFile(filepath.Join(game, "RKERNEL.COM"), []byte("bad"), 0600)
	p = inspectNetworkKernel(game, "1.40b23")
	if p.OK || p.ActualSHA256 == "" {
		t.Fatal(p)
	}
}
func TestLegacySourceGetsMandatoryKernelPin(t *testing.T) {
	for _, id := range []string{"steam-en-140b23", "legacy-en-131", "cd-en-12"} {
		s, e := editionByID(id)
		if e != nil {
			t.Fatal(e)
		}
		found := false
		for _, f := range normalizedBaseFiles(s, "1.50.26") {
			if f.Name == "RKERNEL.COM" {
				found = f.SHA256 == Kernel140Hash && f.Size == kernelSize
			}
		}
		if !found {
			t.Fatal(id)
		}
	}
}
func TestNetworkModesShareVerifiedGameDirectory(t *testing.T) {
	game := filepath.Join(t.TempDir(), "has spaces", "game")
	p := defaultProfiles()[0]
	p.Engine = "1.50.26"
	p.Core = "150"
	p.Port = 21300
	p.Host = "192.168.1.25"
	for _, role := range []string{"host", "join"} {
		for _, service := range []string{"direct", "dopefish"} {
			p.Role = role
			p.NetworkService = service
			cfg, e := dosboxConfig(game, p)
			if e != nil {
				t.Fatal(e)
			}
			if !strings.Contains(cfg, `mount c "`+game+`"`) || !strings.Contains(cfg, "c:\ncd \\\n") || !strings.Contains(cfg, "\nORION150.EXE\n") {
				t.Fatal(cfg)
			}
			if service == "dopefish" && (!strings.Contains(cfg, "IPXNET CONNECT moo2.thedopefish.com 213") || strings.Contains(cfg, "STARTSERVER")) {
				t.Fatal(cfg)
			}
		}
	}
}
