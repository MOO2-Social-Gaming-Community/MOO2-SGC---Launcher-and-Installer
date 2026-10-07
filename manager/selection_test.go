package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestSelectionPersistenceAndMigration(t *testing.T) {
	root, data := t.TempDir(), t.TempDir()
	m, e := newManager(root, data)
	if e != nil {
		t.Fatal(e)
	}
	if m.selectedProfileID() != "baseline" {
		t.Fatal("first-run default changed")
	}
	setting := []byte(`{"runtime_path":"a custom/runtime"}`)
	_ = atomicWrite(filepath.Join(data, "settings.json"), setting)
	if e = m.rememberProfile("community"); e != nil {
		t.Fatal(e)
	}
	n, e := newManager(root, data)
	if e != nil || n.selectedProfileID() != "community" {
		t.Fatal("selection not restored", e)
	}
	b, _ := os.ReadFile(filepath.Join(data, "settings.json"))
	if string(b) != string(setting) {
		t.Fatal("runtime preferences changed")
	}
	es, _ := os.ReadDir(filepath.Join(data, "environments"))
	if len(es) != 0 {
		t.Fatal("selecting prepared/modified the game")
	}
	if e = m.rememberProfile("../escape"); e == nil {
		t.Fatal("traversal accepted")
	}
	if e = m.rememberProfile("not-found"); e == nil {
		t.Fatal("missing profile accepted")
	}
	if m.selectedProfileID() != "community" {
		t.Fatal("invalid selection lost known preference")
	}
}
func TestSelectionInvalidStateFallsBack(t *testing.T) {
	m, _ := newManager(t.TempDir(), t.TempDir())
	for _, b := range []string{"{", `{"schema":2,"profile_id":"community"}`, `{"schema":1,"profile_id":"gone"}`, `{"schema":1,"profile_id":"../escape"}`} {
		_ = atomicWrite(filepath.Join(m.Data, "ui-selection.json"), []byte(b))
		if m.selectedProfileID() != "baseline" {
			t.Fatal(b)
		}
	}
}
func TestSelectionAPIAuthenticationAndIsolation(t *testing.T) {
	s := testServer(t)
	for _, tc := range []struct {
		method, body, auth, origin string
		code                       int
	}{
		{"POST", `{"id":"community"}`, "", "", 401},
		{"POST", `{"id":"community"}`, "Bearer testtoken", "https://other.test", 403},
		{"GET", `{"id":"community"}`, "Bearer testtoken", "", 400},
		{"POST", `{"id":"community","extra":true}`, "Bearer testtoken", "", 400},
		{"POST", `{"id":"../../escape"}`, "Bearer testtoken", "", 400},
		{"POST", `{"id":"community"}`, "Bearer testtoken", "", 200},
	} {
		w := req(s, tc.method, "/api/selection", tc.body, tc.auth, tc.origin, s.host)
		if w.Code != tc.code {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	w := req(s, "GET", "/api/state", "", "Bearer testtoken", "", s.host)
	if !strings.Contains(w.Body.String(), `"selected_profile_id":"community"`) {
		t.Fatal(w.Body.String())
	}
}
func TestConcurrentSelection(t *testing.T) {
	m, _ := newManager(t.TempDir(), t.TempDir())
	var wg sync.WaitGroup
	for _, id := range []string{"baseline", "community", "original", "cd-original"} {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				if e := m.rememberProfile(id); e != nil {
					t.Error(e)
				}
				_ = m.selectedProfileID()
			}
		}(id)
	}
	wg.Wait()
	if _, e := m.loadProfile(m.selectedProfileID()); e != nil {
		t.Fatal(e)
	}
}
