package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testServer(t *testing.T) *localServer {
	t.Helper()
	m, e := newManager(t.TempDir(), t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	return &localServer{manager: m, token: "testtoken", host: "127.0.0.1:9999"}
}
func req(s *localServer, method, path, body, auth, origin, host string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://127.0.0.1:9999"+path, strings.NewReader(body))
	r.Host = host
	if auth != "" {
		r.Header.Set("Authorization", auth)
	}
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handler(w, r)
	return w
}
func TestAPINeedsToken(t *testing.T) {
	s := testServer(t)
	w := req(s, "GET", "/api/state", "", "", "", s.host)
	if w.Code != 401 {
		t.Fatal(w.Code)
	}
}
func TestHostRebindingRejected(t *testing.T) {
	s := testServer(t)
	w := req(s, "GET", "/api/state", "", "Bearer testtoken", "", "evil.example")
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
}
func TestCrossOriginRejected(t *testing.T) {
	s := testServer(t)
	w := req(s, "GET", "/api/state", "", "Bearer testtoken", "https://evil.example", s.host)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
}
func TestAuthorizedState(t *testing.T) {
	s := testServer(t)
	w := req(s, "GET", "/api/state", "", "Bearer testtoken", "", s.host)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "testtoken") {
		t.Fatal("secret leaked")
	}
}
func TestStaticAppServed(t *testing.T) {
	s := testServer(t)
	for _, p := range []string{"/", "/app.js", "/style.css"} {
		w := req(s, "GET", p, "", "", "", s.host)
		if w.Code != 200 || w.Body.Len() < 100 {
			t.Fatal(p)
		}
		if w.Header().Get("Content-Security-Policy") == "" {
			t.Fatal("missing CSP")
		}
	}
}
func TestNoArbitraryFileServing(t *testing.T) {
	s := testServer(t)
	w := req(s, "GET", "/payloads/base.zip", "", "", "", s.host)
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
}
func TestUnknownActionRejected(t *testing.T) {
	s := testServer(t)
	w := req(s, "POST", "/api/action", `{"action":"exec"}`, "Bearer testtoken", "", s.host)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
func TestActionGETRejected(t *testing.T) {
	s := testServer(t)
	w := req(s, "GET", "/api/action", `{"action":"prepare"}`, "Bearer testtoken", "", s.host)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
func TestUnknownFieldsRejected(t *testing.T) {
	s := testServer(t)
	w := req(s, "POST", "/api/action", `{"action":"prepare","command":"arbitrary"}`, "Bearer testtoken", "", s.host)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
func TestPRSLAPICannotEnable(t *testing.T) {
	s := testServer(t)
	p := validProfile()
	p.Mods = []string{"prsl"}
	b, _ := json.Marshal(p)
	w := req(s, "POST", "/api/resolve", string(b), "Bearer testtoken", "", s.host)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
func TestGenerationPathTraversalRejected(t *testing.T) {
	s := testServer(t)
	if _, e := s.manager.activate("community", "../../etc"); e == nil {
		t.Fatal("accepted")
	}
}
func TestProfileSaveAndReload(t *testing.T) {
	s := testServer(t)
	p := validProfile()
	p.ID = "testcopy"
	p.Name = "Copy"
	if e := s.manager.saveProfile(p); e != nil {
		t.Fatal(e)
	}
	got, e := s.manager.loadProfile(p.ID)
	if e != nil || got.Name != p.Name {
		t.Fatal(got, e)
	}
}
func TestInvalidBuildDoesNotWriteWorkspace(t *testing.T) {
	s := testServer(t)
	p := validProfile()
	p.Mods = []string{"prsl"}
	if _, e := s.manager.build(p); e == nil {
		t.Fatal("accepted")
	}
	es, _ := os.ReadDir(filepath.Join(s.manager.Data, "environments"))
	if len(es) != 0 {
		t.Fatal("wrote environment")
	}
}
func TestAbsentPayloadHonestFailure(t *testing.T) {
	s := testServer(t)
	if _, e := s.manager.build(validProfile()); e == nil {
		t.Fatal("pretended to install")
	}
}
func TestBusyJobRejectsSecondOperation(t *testing.T) {
	s := testServer(t)
	done := make(chan bool)
	if e := s.manager.startJob("test", func() (any, error) { <-done; return nil, nil }); e != nil {
		t.Fatal(e)
	}
	if e := s.manager.startJob("another", func() (any, error) { return nil, nil }); e == nil {
		t.Fatal("accepted simultaneous jobs")
	}
	close(done)
}
func TestNoAPIWildcards(t *testing.T) {
	s := testServer(t)
	w := req(s, "GET", "/api/state", "", "Bearer testtoken", "", s.host)
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("CORS exposed")
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("cacheable")
	}
}

var _ = http.StatusOK
