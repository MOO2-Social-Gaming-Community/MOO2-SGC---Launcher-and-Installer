package layout

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExplicitMarker(t *testing.T) {
	r := t.TempDir()
	b, e := IsPortable(r)
	if b || e != nil {
		t.Fatal(b, e)
	}
	if e = Enable(r); e != nil {
		t.Fatal(e)
	}
	b, e = IsPortable(r)
	if !b || e != nil {
		t.Fatal(b, e)
	}
	if e = Enable(r); e != nil {
		t.Fatal(e)
	}
}
func TestInvalidMarkerFailsClosed(t *testing.T) {
	r := t.TempDir()
	os.WriteFile(filepath.Join(r, Marker), []byte(`{"schema":999,"layout":"unknown"}`), 0600)
	if _, e := IsPortable(r); e == nil {
		t.Fatal("unknown accepted")
	}
	if e := Enable(r); e == nil {
		t.Fatal("unknown overwritten")
	}
}
