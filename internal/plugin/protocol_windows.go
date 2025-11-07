//go:build windows

package plugin

import (
	"os/exec"
)

// configureResourceLimits configures process isolation for plugin subprocesses on Windows
func configureResourceLimits(cmd *exec.Cmd) {
	// Windows: Create plugin process in a separate job object for resource control
	// Note: Go doesn't expose Windows Job Objects API directly
	// For production Windows deployments, operators should use:
	// - Windows Job Objects (via external wrapper)
	// - Process quotas via Group Policy
	// - Resource Governor (SQL Server style quotas)
	//
	// The 30-second call timeout and 5-second close timeout provide
	// time-based resource control as a defense-in-depth measure

	// No-op for now - Windows process isolation requires more complex setup
	// that Go doesn't natively support
}

// killProcessGroup kills the plugin process
func killProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}

	// Windows: Kill the process
	// Note: This doesn't kill child processes - for full cleanup,
	// operators should use Job Objects or external process management
	return cmd.Process.Kill()
}
