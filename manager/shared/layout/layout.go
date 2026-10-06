// Package layout defines the explicit portable-root marker. It carries no paths,
// executable instructions, credentials, or trust overrides.
package layout

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const WindowsRoot = `C:\Games\MOO2-SGC`
const Marker = "moo2-sgc-portable.json"

type Config struct {
	Schema int    `json:"schema"`
	Layout string `json:"layout"`
}

func IsPortable(root string) (bool, error) {
	p := filepath.Join(root, Marker)
	st, e := os.Lstat(p)
	if os.IsNotExist(e) {
		return false, nil
	}
	if e != nil {
		return false, e
	}
	if !st.Mode().IsRegular() || st.Size() > 1024 {
		return false, fmt.Errorf("invalid portable layout marker")
	}
	b, e := os.ReadFile(p)
	if e != nil {
		return false, e
	}
	var c Config
	if e = json.Unmarshal(b, &c); e != nil || c.Schema != 1 || c.Layout != "portable-baseline" {
		return false, fmt.Errorf("unsupported portable layout marker; existing files were not moved")
	}
	return true, nil
}
func Enable(root string) error {
	yes, e := IsPortable(root)
	if e != nil {
		return e
	}
	if yes {
		return nil
	}
	return os.WriteFile(filepath.Join(root, Marker), []byte("{\n  \"schema\": 1,\n  \"layout\": \"portable-baseline\"\n}\n"), 0600)
}
