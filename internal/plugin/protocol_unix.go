//go:build unix

package plugin

import (
	"os/exec"
	"syscall"
)

// configureResourceLimits configures process isolation for plugin subprocesses on Unix systems
func configureResourceLimits(cmd *exec.Cmd) {
	// Set process group ID to isolate plugin process
	// This allows us to kill the entire process group (plugin + any children)
	// and prevents plugin from receiving signals sent to parent
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group
	}

	// Note: Go doesn't provide a mechanism to set per-child rlimits
	// For production deployments, operators should use:
	// - systemd service limits (LimitAS, LimitNPROC, LimitFSIZE, LimitCPU)
	// - cgroups v2 for memory/CPU limits
	// - ulimit settings in the shell that launches the binary
	//
	// The 30-second call timeout and 5-second close timeout provide
	// time-based resource control as a defense-in-depth measure
}

// killProcessGroup kills the plugin process and all its children
func killProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}

	// Kill the entire process group (negative PID)
	// This ensures all child processes spawned by the plugin are also killed
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		// Kill process group with SIGKILL
		syscall.Kill(-pgid, syscall.SIGKILL)
	}

	// Fallback: kill just the main process if getpgid fails
	return cmd.Process.Kill()
}
