package sdk

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Plugin is the base interface all plugins must implement
type Plugin interface {
	Info() PluginInfo
	Configure(config map[string]string) error
	Health() HealthResult
}

// SourcePlugin extends Plugin for data source plugins
type SourcePlugin interface {
	Plugin
	Sync(params SyncParams) (SyncResult, error)
}

// DestinationPlugin extends Plugin for data destination plugins
type DestinationPlugin interface {
	Plugin
	Test() TestResult
	Export(params ExportParams) (ExportResult, error)
}

// ProcessorPlugin extends Plugin for data processor plugins
type ProcessorPlugin interface {
	Plugin
	Capabilities() []string
	Analyze(request AnalyzeRequest) (AnalyzeResult, error)
}

// Serve starts the plugin and handles JSON-RPC communication over stdin/stdout
//
// IMPORTANT: The parent process is responsible for enforcing timeouts on individual
// plugin method calls. However, this function includes a defensive idle timeout to prevent
// runaway plugins if the parent crashes or fails to manage the subprocess lifecycle properly.
func Serve(plugin Plugin) {
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	// Set max request size to 10MB to prevent memory issues
	const maxRequestSize = 10 * 1024 * 1024
	buf := make([]byte, maxRequestSize)
	scanner.Buffer(buf, maxRequestSize)

	// Defensive idle timeout: exit if no requests received for 5 minutes
	// This prevents plugin from running forever if parent crashes
	const maxIdleTime = 5 * time.Minute
	lastActivity := time.Now()
	idleTimer := time.NewTimer(maxIdleTime)
	defer idleTimer.Stop()

	// Monitor for idle timeout in background
	go func() {
		<-idleTimer.C
		if time.Since(lastActivity) >= maxIdleTime {
			fmt.Fprintf(os.Stderr, "plugin idle timeout exceeded (%v), exiting\n", maxIdleTime)
			os.Exit(0)
		}
	}()

	for scanner.Scan() {
		// Reset idle timer on each request
		lastActivity = time.Now()
		if !idleTimer.Stop() {
			select {
			case <-idleTimer.C:
			default:
			}
		}
		idleTimer.Reset(maxIdleTime)
		var request Request
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			response := Response{
				Error: &RPCError{
					Code:    ErrCodePluginInternal,
					Message: fmt.Sprintf("failed to parse request: %v", err),
				},
				ID: "unknown",
			}
			writeResponse(writer, response)
			continue
		}

		response := handleRequest(plugin, request)
		writeResponse(writer, response)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "scanner error: %v\n", err)
		os.Exit(1)
	}
}

func handleRequest(plugin Plugin, request Request) Response {
	switch request.Method {
	case "plugin.info":
		return Response{
			Result: plugin.Info(),
			ID:     request.ID,
		}

	case "plugin.configure":
		var config map[string]string
		if err := unmarshalParams(request.Params, &config); err != nil {
			return errorResponse(request.ID, ErrCodeConfig, err.Error())
		}
		if err := plugin.Configure(config); err != nil {
			return errorResponse(request.ID, ErrCodeConfig, err.Error())
		}
		return Response{
			Result: map[string]string{"status": "ok"},
			ID:     request.ID,
		}

	case "plugin.health":
		return Response{
			Result: plugin.Health(),
			ID:     request.ID,
		}

	case "source.sync":
		sourcePlugin, ok := plugin.(SourcePlugin)
		if !ok {
			return errorResponse(request.ID, ErrCodePluginInternal, "plugin does not support source.sync")
		}
		var params SyncParams
		if err := unmarshalParams(request.Params, &params); err != nil {
			return errorResponse(request.ID, ErrCodeConfig, err.Error())
		}
		result, err := sourcePlugin.Sync(params)
		if err != nil {
			return errorResponseFromError(request.ID, err)
		}
		return Response{
			Result: result,
			ID:     request.ID,
		}

	case "destination.test":
		destPlugin, ok := plugin.(DestinationPlugin)
		if !ok {
			return errorResponse(request.ID, ErrCodePluginInternal, "plugin does not support destination.test")
		}
		result := destPlugin.Test()
		return Response{
			Result: result,
			ID:     request.ID,
		}

	case "destination.export":
		destPlugin, ok := plugin.(DestinationPlugin)
		if !ok {
			return errorResponse(request.ID, ErrCodePluginInternal, "plugin does not support destination.export")
		}
		var params ExportParams
		if err := unmarshalParams(request.Params, &params); err != nil {
			return errorResponse(request.ID, ErrCodeConfig, err.Error())
		}
		result, err := destPlugin.Export(params)
		if err != nil {
			return errorResponseFromError(request.ID, err)
		}
		return Response{
			Result: result,
			ID:     request.ID,
		}

	case "processor.capabilities":
		procPlugin, ok := plugin.(ProcessorPlugin)
		if !ok {
			return errorResponse(request.ID, ErrCodePluginInternal, "plugin does not support processor.capabilities")
		}
		result := procPlugin.Capabilities()
		return Response{
			Result: map[string]interface{}{"capabilities": result},
			ID:     request.ID,
		}

	case "processor.analyze":
		procPlugin, ok := plugin.(ProcessorPlugin)
		if !ok {
			return errorResponse(request.ID, ErrCodePluginInternal, "plugin does not support processor.analyze")
		}
		var analyzeReq AnalyzeRequest
		if err := unmarshalParams(request.Params, &analyzeReq); err != nil {
			return errorResponse(request.ID, ErrCodeConfig, err.Error())
		}
		result, err := procPlugin.Analyze(analyzeReq)
		if err != nil {
			return errorResponseFromError(request.ID, err)
		}
		return Response{
			Result: result,
			ID:     request.ID,
		}

	default:
		return errorResponse(request.ID, ErrCodePluginInternal, fmt.Sprintf("unknown method: %s", request.Method))
	}
}

func unmarshalParams(params interface{}, target interface{}) error {
	data, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	// Prevent memory exhaustion from oversized params
	const maxParamsSize = 10 * 1024 * 1024 // 10MB
	if len(data) > maxParamsSize {
		return fmt.Errorf("params too large: %d bytes (max %d)", len(data), maxParamsSize)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal params: %w", err)
	}
	return nil
}

func errorResponse(id string, code int, message string) Response {
	return Response{
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
}

func errorResponseFromError(id string, err error) Response {
	// Check if it's a PluginError with a specific code
	if pluginErr, ok := err.(*PluginError); ok {
		return errorResponse(id, pluginErr.Code, pluginErr.Message)
	}
	// Default to internal error
	return errorResponse(id, ErrCodePluginInternal, err.Error())
}

func writeResponse(writer *bufio.Writer, response Response) {
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal response: %v\n", err)
		return
	}

	if _, err := writer.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write response: %v\n", err)
		os.Exit(1)
	}

	if err := writer.WriteByte('\n'); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write newline: %v\n", err)
		os.Exit(1)
	}

	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to flush response: %v\n", err)
		os.Exit(1)
	}
}
