package main

// The portable runtime contract is a translation of the supplied, user-tested
// Windows harness. Source recognition is separate from mutable runtime state.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const ManualBaselineArchiveHash = "50a8985601bc8f73af0486536c3413978bdf260d5d3363b5175ca809d22a2cbb"
const KernelOldHash = "0cf378f98f00e6308805cf5d0678e7478cf055c5ff95a827d1645a77eb5f5013"
const Kernel140Hash = "18e8781f8ce64516e60b2947b487e8999f7d940b25af971359c7e7dea0fe9d97"
const HarnessDOSBoxHash = "67a907289bad3e25a9be7dd4dc38d882c1c7e74949f8daa103a52476d99f24f5"

func normalizeKernel(b []byte) ([]byte, error) {
	h := sha256.Sum256(b)
	identity := hex.EncodeToString(h[:])
	if identity == Kernel140Hash {
		return append([]byte(nil), b...), nil
	}
	if identity != KernelOldHash || len(b) <= 0x7697 || b[0x765C] != 1 || b[0x765D] != 0x74 || b[0x7697] != 1 {
		return nil, errors.New("RKERNEL.COM is not the exact known pre-1.40 or LAN-fixed input; no blind patch applied")
	}
	out := append([]byte(nil), b...)
	out[0x765C] = 0
	out[0x765D] = 0x75
	out[0x7697] = 0
	h = sha256.Sum256(out)
	if hex.EncodeToString(h[:]) != Kernel140Hash {
		return nil, errors.New("RKERNEL output verification failed")
	}
	return out, nil
}
func normalizeHarnessGame(game string, s SourceEdition, target string) error {
	if engineRank(target) < 140 {
		return nil
	}
	if !s.Harness {
		return errors.New("this older imported snapshot omitted RKERNEL.COM. Re-import your original CD/Steam/baseline ZIP or folder; it remains unchanged")
	}
	exe := filepath.Join(game, "ORION2.EXE")
	if s.EngineFile != "" {
		b, e := os.ReadFile(exe)
		if e != nil {
			return e
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != BaseEngineHash {
			return errors.New("manual baseline did not retain the expected official 1.31 executable")
		}
		if e = atomicWrite(filepath.Join(game, "ORION131.EXE"), b); e != nil {
			return e
		}
		src := filepath.Join(game, s.EngineFile)
		if e = requireHash(src, BaselineEngineHash); e != nil {
			return e
		}
		if e = copyFile(src, exe); e != nil {
			return e
		}
	}
	rk := filepath.Join(game, "RKERNEL.COM")
	b, e := os.ReadFile(rk)
	if e != nil {
		return fmt.Errorf("baseline requires RKERNEL.COM: %w", e)
	}
	b, e = normalizeKernel(b)
	if e != nil {
		return e
	}
	return atomicWrite(rk, b)
}
func normalizedBaseFiles(s SourceEdition, target string) []BaseFile {
	records := map[string]BaseFile{}
	for _, f := range expectedBaseFiles(s, target) {
		records[f.Name] = f
	}
	if engineRank(target) >= 140 && s.Harness {
		records["RKERNEL.COM"] = BaseFile{"RKERNEL.COM", 31095, Kernel140Hash}
		if s.EngineFile != "" || engineRank(s.Version) < 140 {
			records["ORION131.EXE"] = BaseFile{"ORION131.EXE", 2612010, BaseEngineHash}
		}
	}
	out := []BaseFile{}
	for _, r := range records {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

type HarnessRuntimeIndex struct {
	Schema   int        `json:"schema"`
	Version  string     `json:"version"`
	Platform string     `json:"platform"`
	Files    []BaseFile `json:"files"`
}

func harnessRuntimeIndex() HarnessRuntimeIndex {
	b, e := resources.ReadFile("assets/harness-runtime.json")
	if e != nil {
		panic(e)
	}
	var idx HarnessRuntimeIndex
	if e = json.Unmarshal(b, &idx); e != nil || idx.Schema != 1 {
		panic("bad embedded harness runtime index")
	}
	return idx
}

// Verify all DLLs/resources as well as the executable. Unknown local files are
// not executed; extra code/config that could change runtime behavior is refused.
func verifyHarnessRuntime(dir string) error {
	if e := noSymlinkAncestors(dir); e != nil {
		return e
	}
	expected := map[string]bool{}
	for _, r := range harnessRuntimeIndex().Files {
		rel, e := safeRel(r.Name)
		if e != nil {
			return e
		}
		f := filepath.Join(dir, rel)
		if e = noSymlinkAncestors(f); e != nil {
			return e
		}
		if e = regularFile(f); e != nil {
			return e
		}
		st, e := os.Stat(f)
		if e != nil {
			return e
		}
		if st.Size() != r.Size {
			return fmt.Errorf("harness runtime size mismatch: %s", r.Name)
		}
		if e = requireHash(f, r.SHA256); e != nil {
			return e
		}
		expected[strings.ToLower(r.Name)] = true
	}
	return walkRegular(dir, func(n, p string) error {
		if expected[strings.ToLower(n)] {
			return nil
		}
		switch strings.ToLower(filepath.Ext(n)) {
		case ".dll", ".exe", ".com", ".conf":
			return fmt.Errorf("unmanaged runtime code/config: %s", n)
		}
		return nil
	})
}
func (m *Manager) harnessRuntimePath() string {
	return filepath.Join(m.Root, "runtime", "windows", "dosbox.exe")
}
func (m *Manager) localHarnessRuntime() (string, error) {
	p := m.harnessRuntimePath()
	if e := regularFile(p); e != nil {
		return "", e
	}
	if e := verifyHarnessRuntime(filepath.Dir(p)); e != nil {
		return "", fmt.Errorf("local portable DOSBox failed verification: %w", e)
	}
	return p, nil
}
func harnessBaseConfig() string {
	b, e := resources.ReadFile("assets/harness/moo2.conf")
	if e != nil {
		panic(e)
	}
	return strings.TrimRight(string(b), "\r\n") + "\n"
}

// Literal file comparison is used in tests; strip only the handoff's stray
// leading backslash artifact, not any validated DOSBox configuration directive.
func validHarnessConfig(b []byte) bool {
	return bytes.Equal(bytes.TrimSpace(b), bytes.TrimSpace([]byte(harnessBaseConfig())))
}
