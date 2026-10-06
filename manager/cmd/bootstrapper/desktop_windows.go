//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

func showSetupError(err error) {
	// Preserve a double-click failure on screen; no administrator request or
	// antivirus bypass. CLI callers with arguments retain ordinary stderr only.
	if len(os.Args) > 1 {
		return
	}
	text, _ := syscall.UTF16PtrFromString("MOO2-SGC setup could not finish.\n\n" + err.Error() + "\n\nSee C:\\Games\\MOO2-SGC\\logs\\bootstrap.log for new portable installs, or the explicit installation root when updating an older install.")
	title, _ := syscall.UTF16PtrFromString("MOO2-SGC Setup")
	_, _, _ = syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x10)
}
func installShortcut(root, entry string) error {
	helper := filepath.Join(filepath.Dir(entry), "MOO2-SGC-Setup.exe")
	if _, e := os.Stat(helper); e != nil {
		return nil
	} // Older signed generations lack it.
	if strings.ContainsAny(root, "\"\r\n") {
		return fmt.Errorf("unsupported shortcut path")
	}
	folder := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs")
	if e := os.MkdirAll(folder, 0700); e != nil {
		return e
	}
	script := `$ErrorActionPreference='Stop'; $w=New-Object -ComObject WScript.Shell; $s=$w.CreateShortcut($env:SGC_SHORTCUT); $s.TargetPath=$env:SGC_TARGET; $s.Arguments='--command launch-installed --install-root "'+$env:SGC_ROOT+'"'; $s.WorkingDirectory=$env:SGC_ROOT; $s.Description='MOO2-SGC Launcher'; $s.Save()`
	c := exec.Command("powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", script)
	c.Env = append(os.Environ(), "SGC_SHORTCUT="+filepath.Join(folder, "MOO2-SGC.lnk"), "SGC_TARGET="+helper, "SGC_ROOT="+root)
	b, e := c.CombinedOutput()
	if e != nil {
		return fmt.Errorf("Start menu link: %s: %w", b, e)
	}
	return nil
}
