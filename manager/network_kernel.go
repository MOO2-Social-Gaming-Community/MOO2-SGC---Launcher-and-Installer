package main

// The network driver is a component of the user's owned game, not a public
// launcher payload. Recovery copies only exact known bytes into a fresh stage.
import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const kernelSize = 31095

var errKernelMissing = errors.New("RKERNEL.COM not present in selected source")

func readOwnedKernel(source string) ([]byte, error) {
	if e := noSymlinkAncestors(source); e != nil {
		return nil, e
	}
	if e := regularFile(source); e != nil {
		return nil, e
	}
	if strings.EqualFold(filepath.Ext(source), ".zip") {
		z, e := zip.OpenReader(source)
		if e != nil {
			return nil, e
		}
		defer z.Close()
		files, e := zipSourceFiles(z)
		if e != nil {
			return nil, e
		}
		var found *zip.File
		for n, f := range files {
			if path.Base(n) == "RKERNEL.COM" {
				if found != nil {
					return nil, errors.New("ambiguous archive: multiple RKERNEL.COM members")
				}
				found = f
			}
		}
		if found == nil {
			return nil, errKernelMissing
		}
		b, e := readZipKnown(found, kernelSize)
		if e != nil {
			return nil, e
		}
		return normalizeKernel(b)
	}
	if !strings.EqualFold(filepath.Base(source), "RKERNEL.COM") {
		return nil, errors.New("select RKERNEL.COM or a ZIP containing it; RKERNEL.EXE is not the required file")
	}
	f, e := os.Open(source)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, kernelSize+1))
	if e != nil {
		return nil, e
	}
	return normalizeKernel(b)
}
func (m *Manager) kernelCache() string {
	return filepath.Join(m.Data, "cache", "network-kernel", "RKERNEL.COM")
}
func (m *Manager) importKernel(source string) (any, error) {
	p, e := localPath(source)
	if e != nil {
		return nil, e
	}
	b, e := readOwnedKernel(p)
	if e != nil {
		return nil, e
	}
	dst := m.kernelCache()
	if e = noSymlinkAncestors(dst); e != nil {
		return nil, e
	}
	if e = os.MkdirAll(filepath.Dir(dst), 0700); e != nil {
		return nil, e
	}
	if e = atomicWrite(dst, b); e != nil {
		return nil, e
	}
	if e = requireHash(dst, Kernel140Hash); e != nil {
		return nil, e
	}
	return map[string]any{"component": "RKERNEL.COM", "sha256": Kernel140Hash, "size": kernelSize, "cached": dst, "source_modified": false, "activation": "Prepare / repair the selected profile to restore the kernel transactionally; no active game files changed"}, nil
}
func (m *Manager) ensureStageKernel(game string) error {
	dst := filepath.Join(game, "RKERNEL.COM")
	if e := regularFile(dst); e == nil {
		return nil
	} else if !os.IsNotExist(e) {
		return e
	}
	candidates := []string{m.kernelCache()}
	// Older imported snapshots may have dropped the kernel. Prefer a complete
	// local source over a workaround in the active game; never fetch random code.
	if source, _, _, e := m.payload("base"); e == nil {
		candidates = append(candidates, source)
	}
	candidates = append(candidates, filepath.Join(m.Root, "RKERNEL.COM"), filepath.Join(m.Root, "rkernel.zip"), filepath.Join(m.Root, "RKERNEL-LAN-FIXED.zip"))
	if source, e := m.portableSource(); e == nil {
		candidates = append(candidates, source)
	}
	candidates = append(candidates, filepath.Join(m.Root, "game", "RKERNEL.COM"))
	for _, source := range candidates {
		b, e := readOwnedKernel(source)
		if os.IsNotExist(e) || errors.Is(e, errKernelMissing) {
			continue
		}
		if e != nil {
			return fmt.Errorf("network kernel recovery source %s was refused: %w", source, e)
		}
		if e = atomicWrite(dst, b); e != nil {
			return e
		}
		m.progress("Recovered and verified RKERNEL.COM from owned local files in the staged environment; saves and original files stay untouched")
		return requireHash(dst, Kernel140Hash)
	}
	return errors.New("this legacy source snapshot lacks RKERNEL.COM. Re-import the complete owned baseline/Steam folder or ZIP, or import your rkernel.zip under Packages > Network kernel, then Prepare / repair. No game has been started")
}

type NetworkPreflight struct {
	GamePath       string `json:"game_path"`
	KernelPath     string `json:"kernel_path"`
	ExpectedSHA256 string `json:"expected_sha256"`
	ActualSHA256   string `json:"actual_sha256,omitempty"`
	Required       bool   `json:"required"`
	OK             bool   `json:"ok"`
	Error          string `json:"error,omitempty"`
}

func inspectNetworkKernel(game, engine string) NetworkPreflight {
	n := NetworkPreflight{GamePath: game, KernelPath: filepath.Join(game, "RKERNEL.COM"), ExpectedSHA256: Kernel140Hash, Required: engineRank(engine) >= 140, OK: true}
	if !n.Required {
		return n
	}
	if e := noSymlinkAncestors(n.KernelPath); e != nil {
		n.OK = false
		n.Error = e.Error()
		return n
	}
	h, e := hashFile(n.KernelPath)
	n.ActualSHA256 = h
	if e != nil || h != Kernel140Hash {
		n.OK = false
		if e != nil {
			n.Error = e.Error()
		} else {
			n.Error = "kernel differs from the verified canonical 1.40/1.50 driver"
		}
	}
	return n
}
