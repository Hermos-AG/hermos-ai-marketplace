//go:build !windows

package main

import (
	"context"
	"os/exec"
)

// hideWindow is a no-op on non-Windows platforms.
func hideWindow(cmd *exec.Cmd) {}

// newShellCmd runs the command line through the POSIX shell.
func newShellCmd(ctx context.Context, command string) *exec.Cmd {
	return exec.CommandContext(ctx, "/bin/sh", "-c", command)
}
