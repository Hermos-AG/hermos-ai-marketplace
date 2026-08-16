//go:build windows

package main

import (
	"context"
	"os/exec"
	"syscall"
)

// hideWindow prevents spawned console processes (cmd.exe, nvidia-smi) from
// flashing a console window when the server runs headless under the
// Claude desktop app.
func hideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= 0x08000000 // CREATE_NO_WINDOW
}

// newShellCmd builds a cmd.exe invocation that receives the command line
// VERBATIM. Go's default argument escaping turns embedded double quotes into
// \" sequences, which cmd.exe does not understand (it is not an MSVCRT
// parser) — so `"C:\path with spaces\tool.exe" -x` would fail. Setting
// SysProcAttr.CmdLine bypasses Go's escaping entirely; cmd /S /C "..." then
// strips the outer quotes and executes the raw line, quotes and all.
func newShellCmd(ctx context.Context, command string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: `/S /C "` + command + `"`,
	}
	return cmd
}
