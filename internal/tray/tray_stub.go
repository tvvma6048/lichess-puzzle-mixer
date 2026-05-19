//go:build !cgo || !(linux || windows || darwin)

package tray

func available() bool { return false }

func run(_ string, _ func()) {}

func quit() {}
