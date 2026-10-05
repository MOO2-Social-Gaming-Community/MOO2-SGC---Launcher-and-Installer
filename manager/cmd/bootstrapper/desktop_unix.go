//go:build !windows

package main

func showSetupError(error)                     {}
func installShortcut(root, entry string) error { return nil }
