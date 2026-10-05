//go:build !windows

package distribution

import "os"

func replace(a, b string) error { return os.Rename(a, b) }
