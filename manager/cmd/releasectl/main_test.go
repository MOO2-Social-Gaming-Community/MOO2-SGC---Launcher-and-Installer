package main

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLauncherPublicationFence(t *testing.T) {
	for _, name := range []string{"ORION2.EXE", "150/ORION150.EXE", "GALAXY.LBX", "SAVE1.GAM", "private.key", "private.pem", "payloads/base.zip", ".env", "../outside"} {
		t.Run(name, func(t *testing.T) {
			var b bytes.Buffer
			z := zip.NewWriter(&b)
			w, _ := z.Create(name)
			w.Write([]byte("forbidden"))
			z.Close()
			p := filepath.Join(t.TempDir(), "test.zip")
			os.WriteFile(p, b.Bytes(), 0600)
			if _, e := indexZIP(p); e == nil {
				t.Fatal("publication fence accepted " + name)
			}
		})
	}
}
func TestLauncherZIPIndexIncludesAllBytes(t *testing.T) {
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	w, _ := z.Create("launcher")
	w.Write([]byte("test"))
	z.Close()
	p := filepath.Join(t.TempDir(), "test.zip")
	os.WriteFile(p, b.Bytes(), 0600)
	v, e := indexZIP(p)
	if e != nil || v["launcher"].Size != 4 || len(v["launcher"].SHA256) != 64 {
		t.Fatal(v, e)
	}
}
