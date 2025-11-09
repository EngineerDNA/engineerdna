package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/engineerdna/engineerdna/internal/config"
	apperrors "github.com/engineerdna/engineerdna/internal/errors"
	"github.com/engineerdna/engineerdna/plugins/plugin-sdk"
	"github.com/google/uuid"
)

// Client represents a client connection to a plugin process
type Client struct {
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     *bufio.Reader
	stderr     io.ReadCloser
	pluginName string
	stderrBuf  *strings.Builder
}

// NewClient creates a new plugin client for the given plugin path
func NewClient(pluginPath string) (*Client, error) {
	cmd := exec.Command(pluginPath)

	// Configure resource limits and process isolation (platform-specific)
	configureResourceLimits(cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}

	stderrBuf := &strings.Builder{}

	client := &Client{
		cmd:        cmd,
		stdin:      stdin,
		stdout:     bufio.NewReader(stdout),
		stderr:     stderr,
		pluginName: pluginPath,
		stderrBuf:  stderrBuf,
	}

	// Forward plugin stderr to parent stderr and capture for error reporting
	// Limit total stderr to 10MB to prevent flood attacks
	go func() {
		const maxStderrBytes = 10 * 1024 * 1024 // 10 MB limit
		limitedStderr := io.LimitReader(stderr, maxStderrBytes)

		// Use MultiWriter to both forward to os.Stderr and capture in buffer
		multiWriter := io.MultiWriter(os.Stderr, stderrBuf)
		if written, err := io.Copy(multiWriter, limitedStderr); err != nil {
			log.Printf("Error copying plugin stderr: %v", err)
		} else if written >= maxStderrBytes {
			fmt.Fprintf(os.Stderr, "\n[plugin stderr limit exceeded, truncating...]\n")
		}
	}()

	return client, nil
}

// Call sends a request to the plugin and waits for a response
func (c *Client) Call(method string, params interface{}) (interface{}, error) {
	id := uuid.New().String()
	request := sdk.NewRequest(method, params, id)

	// Send request
	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Read response with timeout and size limit
	const maxResponseSize = 100 * 1024 * 1024 // 100 MB limit
	done := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go func() {
		// Read response incrementally to prevent memory exhaustion
		// Use a buffer that grows as needed but fails fast if size exceeded
		buf := make([]byte, 0, 4096) // Start with 4KB capacity
		reader := bufio.NewReader(c.stdout)
		bytesRead := 0

		for {
			b, err := reader.ReadByte()
			if err != nil {
				if err == io.EOF {
					errChan <- fmt.Errorf("unexpected EOF before newline")
				} else {
					errChan <- err
				}
				return
			}

			bytesRead++
			// Check size limit BEFORE allocating memory
			if bytesRead > maxResponseSize {
				errChan <- fmt.Errorf("response too large: exceeds %d bytes (100MB)", maxResponseSize)
				return
			}

			buf = append(buf, b)

			// Found newline - success
			if b == '\n' {
				done <- buf
				return
			}
		}
	}()

	var line []byte
	select {
	case line = <-done:
		// Success
	case err := <-errChan:
		// Check if plugin crashed (EOF before response)
		if strings.Contains(err.Error(), "EOF") {
			return nil, apperrors.NewPluginCrashError(c.pluginName, c.stderrBuf.String())
		}
		return nil, apperrors.NewInternalError("plugin communication", err)
	case <-time.After(config.PluginTimeout):
		// Kill the plugin process on timeout
		killProcessGroup(c.cmd)
		return nil, apperrors.NewTimeoutError("plugin call", config.PluginTimeout)
	}

	var response sdk.Response
	if err := json.Unmarshal(line, &response); err != nil {
		return nil, apperrors.NewInternalError("plugin response parsing", err)
	}

	if response.Error != nil {
		// Convert plugin error to structured error
		return nil, apperrors.WrapError(
			fmt.Errorf("plugin error (code %d): %s", response.Error.Code, response.Error.Message),
			apperrors.ErrorInternal,
			fmt.Sprintf("Plugin '%s' reported an error: %s", c.pluginName, response.Error.Message),
			[]string{
				"Check the plugin configuration",
				"Verify the plugin has necessary permissions",
				"Review plugin logs for details",
			},
		)
	}

	return response.Result, nil
}

// Close terminates the plugin process with timeout
func (c *Client) Close() error {
	c.stdin.Close()

	// Wait for process with timeout
	done := make(chan error, 1)
	go func() {
		done <- c.cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("plugin process error: %w", err)
		}
		return nil
	case <-time.After(5 * time.Second):
		// Force kill if timeout (kills entire process group on Unix)
		killProcessGroup(c.cmd)
		return fmt.Errorf("plugin process killed after timeout")
	}
}

// GetInfo retrieves plugin metadata
func (c *Client) GetInfo() (*sdk.PluginInfo, error) {
	result, err := c.Call("plugin.info", nil)
	if err != nil {
		return nil, err
	}

	// Convert result to PluginInfo
	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var info sdk.PluginInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin info: %w", err)
	}

	return &info, nil
}

// Configure sends configuration to the plugin
func (c *Client) Configure(config map[string]string) error {
	_, err := c.Call("plugin.configure", config)
	return err
}

// Health checks the plugin health
func (c *Client) Health() (*sdk.HealthResult, error) {
	result, err := c.Call("plugin.health", nil)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var health sdk.HealthResult
	if err := json.Unmarshal(data, &health); err != nil {
		return nil, fmt.Errorf("failed to unmarshal health result: %w", err)
	}

	return &health, nil
}

// Sync calls the source plugin sync method
func (c *Client) Sync(params sdk.SyncParams) (*sdk.SyncResult, error) {
	result, err := c.Call("source.sync", params)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var syncResult sdk.SyncResult
	if err := json.Unmarshal(data, &syncResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sync result: %w", err)
	}

	return &syncResult, nil
}

// Test calls the destination plugin test method
func (c *Client) Test() (*sdk.TestResult, error) {
	result, err := c.Call("destination.test", nil)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var testResult sdk.TestResult
	if err := json.Unmarshal(data, &testResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal test result: %w", err)
	}

	return &testResult, nil
}

// Export calls the destination plugin export method
func (c *Client) Export(params sdk.ExportParams) (*sdk.ExportResult, error) {
	result, err := c.Call("destination.export", params)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var exportResult sdk.ExportResult
	if err := json.Unmarshal(data, &exportResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal export result: %w", err)
	}

	return &exportResult, nil
}

// GetCapabilities calls the processor plugin capabilities method
func (c *Client) GetCapabilities() ([]string, error) {
	result, err := c.Call("processor.capabilities", nil)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var capResult struct {
		Capabilities []string `json:"capabilities"`
	}
	if err := json.Unmarshal(data, &capResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	return capResult.Capabilities, nil
}

// Analyze calls the processor plugin analyze method
func (c *Client) Analyze(request sdk.AnalyzeRequest) (*sdk.AnalyzeResult, error) {
	result, err := c.Call("processor.analyze", request)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var analyzeResult sdk.AnalyzeResult
	if err := json.Unmarshal(data, &analyzeResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analyze result: %w", err)
	}

	return &analyzeResult, nil
}
