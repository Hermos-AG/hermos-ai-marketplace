// gpu-mcp — a dependency-free MCP (Model Context Protocol) stdio server that
// exposes the local NVIDIA GPU (via nvidia-smi) and command execution to an
// MCP client such as the Claude desktop app.
//
// Protocol: JSON-RPC 2.0, newline-delimited, over stdin/stdout
// (MCP stdio transport). Logging goes to stderr only — stdout is reserved
// for protocol messages.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	serverName        = "hermos-local-gpu"
	serverVersion     = "0.2.0"
	defaultProtocol   = "2024-11-05"
	nvidiaSmiTimeout  = 20 * time.Second
	defaultCmdTimeout = 60 * time.Second
	maxCmdTimeout     = 3600 * time.Second
	maxOutputBytes    = 200 * 1024 // tool output is truncated beyond this
	defaultMinVRAMMiB = 8192       // HERMOS GPU requirement (override: GPU_MCP_MIN_VRAM_MB)
)

// ---------------------------------------------------------------------------
// JSON-RPC types
// ---------------------------------------------------------------------------

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// ---------------------------------------------------------------------------
// Tool definitions
// ---------------------------------------------------------------------------

type toolAnnotations struct {
	ReadOnlyHint    bool `json:"readOnlyHint"`
	DestructiveHint bool `json:"destructiveHint"`
	IdempotentHint  bool `json:"idempotentHint"`
	OpenWorldHint   bool `json:"openWorldHint"`
}

type toolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema map[string]any  `json:"inputSchema"`
	Annotations toolAnnotations `json:"annotations"`
}

var emptySchema = map[string]any{
	"type":                 "object",
	"properties":           map[string]any{},
	"additionalProperties": false,
}

var readOnly = toolAnnotations{ReadOnlyHint: true, DestructiveHint: false, IdempotentHint: true, OpenWorldHint: false}

var tools = []toolDef{
	{
		Name: "gpu_get_status",
		Description: "Full human-readable `nvidia-smi` report for the NVIDIA GPU(s) in this machine: " +
			"driver and CUDA version, per-GPU utilization, memory, temperature, power, and all processes " +
			"currently using a GPU. Best first call for an overall picture.",
		InputSchema: emptySchema,
		Annotations: readOnly,
	},
	{
		Name: "gpu_query_metrics",
		Description: "Current GPU metrics in CSV form (first line is the header): index, name, driver_version, " +
			"utilization.gpu [%], utilization.memory [%], memory.used [MiB], memory.total [MiB], temperature.gpu, " +
			"power.draw [W], power.limit [W], clocks.sm [MHz], fan.speed [%]. Fields a GPU does not support are " +
			"reported as [N/A]. Use this for compact, machine-readable monitoring.",
		InputSchema: emptySchema,
		Annotations: readOnly,
	},
	{
		Name: "gpu_list_processes",
		Description: "Processes currently using GPU compute (CUDA) as CSV: pid, process_name, " +
			"used_gpu_memory [MiB]. Pure graphics workloads may not be listed here — " +
			"gpu_get_status shows all GPU processes.",
		InputSchema: emptySchema,
		Annotations: readOnly,
	},
	{
		Name: "gpu_check_requirements",
		Description: "Check whether this machine meets the HERMOS GPU requirement: an NVIDIA GPU with at " +
			"least the configured minimum VRAM (default 8192 MiB; override via the GPU_MCP_MIN_VRAM_MB " +
			"environment variable, 0 disables the VRAM minimum). Returns a per-GPU report ending in " +
			"'RESULT: MET' or 'RESULT: NOT MET' (isError mirrors NOT MET). Call this first when unsure " +
			"whether GPU jobs can run on this machine.",
		InputSchema: emptySchema,
		Annotations: readOnly,
	},
	{
		Name: "gpu_run_command",
		Description: "Run a shell command on this machine (Windows: via `cmd /C`) and return its combined " +
			"stdout/stderr plus exit code. Intended for starting GPU jobs, e.g. `python train.py` or " +
			"`ollama run ...`. Runs with the permissions of the logged-in user — use with care. " +
			"Output is truncated after 200 KB.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "The command line to execute, e.g. 'nvidia-smi -L' or 'python C:\\jobs\\train.py'.",
				},
				"timeout_seconds": map[string]any{
					"type":        "number",
					"description": "Maximum runtime in seconds before the command is killed (default 60, max 3600).",
				},
				"working_dir": map[string]any{
					"type":        "string",
					"description": "Directory to run the command in (optional).",
				},
			},
			"required":             []string{"command"},
			"additionalProperties": false,
		},
		Annotations: toolAnnotations{ReadOnlyHint: false, DestructiveHint: true, IdempotentHint: false, OpenWorldHint: false},
	},
}

// ---------------------------------------------------------------------------
// nvidia-smi helpers
// ---------------------------------------------------------------------------

// findNvidiaSmi locates the nvidia-smi executable. The GPU_MCP_NVIDIA_SMI
// environment variable overrides the search (useful for tests and unusual
// driver installations).
func findNvidiaSmi() (string, error) {
	if p := os.Getenv("GPU_MCP_NVIDIA_SMI"); p != "" {
		return p, nil
	}
	if p, err := exec.LookPath("nvidia-smi"); err == nil {
		return p, nil
	}
	if runtime.GOOS == "windows" {
		for _, c := range []string{
			`C:\Windows\System32\nvidia-smi.exe`,
			`C:\Program Files\NVIDIA Corporation\NVSMI\nvidia-smi.exe`,
		} {
			if st, err := os.Stat(c); err == nil && !st.IsDir() {
				return c, nil
			}
		}
	}
	return "", fmt.Errorf("nvidia-smi not found — is the NVIDIA driver installed? " +
		"Searched PATH plus the default Windows locations. " +
		"Set the GPU_MCP_NVIDIA_SMI environment variable to the full path if it lives elsewhere")
}

func runNvidiaSmi(args ...string) (string, error) {
	path, err := findNvidiaSmi()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), nvidiaSmiTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("nvidia-smi timed out after %s", nvidiaSmiTimeout)
	}
	if err != nil {
		return "", fmt.Errorf("nvidia-smi failed: %v\n%s", err, truncate(out))
	}
	return truncate(out), nil
}

// ---------------------------------------------------------------------------
// GPU requirement preflight
// ---------------------------------------------------------------------------

// minVRAMMiB returns the minimum VRAM (MiB) at least one GPU in this machine
// must have for the HERMOS requirement to be met. Default 8192 MiB; override
// via the GPU_MCP_MIN_VRAM_MB environment variable ("0" disables the VRAM
// minimum, leaving only the NVIDIA-driver check).
func minVRAMMiB() int {
	if v := os.Getenv("GPU_MCP_MIN_VRAM_MB"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n >= 0 {
			return n
		}
		log.Printf("ignoring invalid GPU_MCP_MIN_VRAM_MB=%q (want a non-negative integer)", v)
	}
	return defaultMinVRAMMiB
}

// requirementText describes the effective requirement in one phrase.
func requirementText() string {
	if m := minVRAMMiB(); m > 0 {
		return fmt.Sprintf("NVIDIA GPU with >= %d MiB VRAM", m)
	}
	return "NVIDIA GPU present (VRAM minimum disabled)"
}

// checkRequirements verifies that this machine has a working NVIDIA driver
// and at least one GPU meeting the minimum VRAM. It returns a human-readable
// report and whether the requirement is met.
func checkRequirements() (string, bool) {
	min := minVRAMMiB()
	var b strings.Builder
	fmt.Fprintf(&b, "%s v%s — GPU requirement check\n", serverName, serverVersion)
	fmt.Fprintf(&b, "Requirement: %s (override: GPU_MCP_MIN_VRAM_MB)\n\n", requirementText())

	path, err := findNvidiaSmi()
	if err != nil {
		fmt.Fprintf(&b, "[FAIL] %v\n\nRESULT: NOT MET — GPU jobs cannot run on this machine via gpu-mcp.\n", err)
		return b.String(), false
	}
	fmt.Fprintf(&b, "[ OK ] nvidia-smi: %s\n", path)

	out, err := runNvidiaSmi("--query-gpu=index,name,driver_version,memory.total", "--format=csv,noheader,nounits")
	if err != nil {
		fmt.Fprintf(&b, "[FAIL] querying GPUs: %v\n\nRESULT: NOT MET\n", err)
		return b.String(), false
	}

	met, seen := false, 0
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		fields := strings.Split(line, ",")
		if len(fields) < 4 {
			continue
		}
		seen++
		idx := strings.TrimSpace(fields[0])
		name := strings.TrimSpace(strings.Join(fields[1:len(fields)-2], ","))
		driver := strings.TrimSpace(fields[len(fields)-2])
		memStr := strings.TrimSpace(fields[len(fields)-1])
		mem, memErr := strconv.Atoi(memStr)
		switch {
		case memErr != nil:
			fmt.Fprintf(&b, "[ ?? ] GPU %s: %s — VRAM %q not parseable (driver %s)\n", idx, name, memStr, driver)
		case mem >= min:
			fmt.Fprintf(&b, "[ OK ] GPU %s: %s — %d MiB VRAM (driver %s)\n", idx, name, mem, driver)
			met = true
		default:
			fmt.Fprintf(&b, "[LOW ] GPU %s: %s — %d MiB VRAM, below the %d MiB minimum (driver %s)\n",
				idx, name, mem, min, driver)
		}
	}
	if seen == 0 {
		b.WriteString("[FAIL] nvidia-smi reported no GPUs\n")
	}
	if met {
		b.WriteString("\nRESULT: MET — this machine can run GPU jobs via gpu-mcp.\n")
	} else {
		b.WriteString("\nRESULT: NOT MET — no GPU satisfies the minimum. Do not start GPU jobs here.\n")
	}
	return b.String(), met
}

// ---------------------------------------------------------------------------
// Tool implementations
// ---------------------------------------------------------------------------

func textResult(text string, isError bool) map[string]any {
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
		"isError": isError,
	}
}

// wrap converts (text, err) into an MCP tool result; execution errors are
// reported inside the result (isError=true), not as protocol errors.
func wrap(text string, err error) (any, *rpcError) {
	if err != nil {
		return textResult("Error: "+err.Error(), true), nil
	}
	return textResult(text, false), nil
}

func runUserCommand(args map[string]any) map[string]any {
	command, _ := args["command"].(string)
	if strings.TrimSpace(command) == "" {
		return textResult("Error: missing required argument 'command'.", true)
	}

	timeout := defaultCmdTimeout
	if t, ok := args["timeout_seconds"].(float64); ok && t > 0 {
		timeout = time.Duration(t * float64(time.Second))
		if timeout > maxCmdTimeout {
			timeout = maxCmdTimeout
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := newShellCmd(ctx, command)
	if wd, ok := args["working_dir"].(string); ok && wd != "" {
		cmd.Dir = wd
	}
	hideWindow(cmd)
	// After a timeout kills the shell, grandchild processes may still hold the
	// output pipes open. WaitDelay makes CombinedOutput return anyway instead
	// of blocking until the whole process tree exits.
	cmd.WaitDelay = 3 * time.Second

	log.Printf("gpu_run_command: %s", command)
	start := time.Now()
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(start).Round(time.Millisecond)

	if ctx.Err() == context.DeadlineExceeded {
		return textResult(fmt.Sprintf("Error: command killed after timeout of %s. Partial output:\n%s",
			timeout, truncate(out)), true)
	}
	if err != nil && cmd.ProcessState == nil {
		return textResult(fmt.Sprintf("Error: failed to start command: %v", err), true)
	}

	exit := 0
	if cmd.ProcessState != nil {
		exit = cmd.ProcessState.ExitCode()
	}
	text := fmt.Sprintf("exit code: %d (took %s)\n\n%s", exit, elapsed, truncate(out))
	return textResult(text, exit != 0)
}

func truncate(out []byte) string {
	if len(out) <= maxOutputBytes {
		return string(out)
	}
	return string(out[:maxOutputBytes]) + "\n… [output truncated at 200 KB]"
}

// ---------------------------------------------------------------------------
// Request handling
// ---------------------------------------------------------------------------

func handleInitialize(params json.RawMessage) map[string]any {
	var p struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	_ = json.Unmarshal(params, &p)
	proto := p.ProtocolVersion
	if proto == "" {
		proto = defaultProtocol
	}
	return map[string]any{
		"protocolVersion": proto,
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo": map[string]any{
			"name":    serverName,
			"version": serverVersion,
			"title":   "HERMOS - local GPU",
		},
		"instructions": "Exposes the NVIDIA GPU of this machine. Use gpu_get_status for an overview, " +
			"gpu_query_metrics for compact CSV metrics, gpu_list_processes for CUDA processes, and " +
			"gpu_run_command to start GPU jobs. Requirement on this machine: " + requirementText() +
			" — verify with gpu_check_requirements before starting heavy GPU jobs.",
	}
}

func handleToolsCall(params json.RawMessage) (any, *rpcError) {
	var p struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcError{Code: -32602, Message: "invalid params: " + err.Error()}
	}
	switch p.Name {
	case "gpu_get_status":
		return wrap(runNvidiaSmi())
	case "gpu_query_metrics":
		return wrap(runNvidiaSmi(
			"--query-gpu=index,name,driver_version,utilization.gpu,utilization.memory,"+
				"memory.used,memory.total,temperature.gpu,power.draw,power.limit,clocks.sm,fan.speed",
			"--format=csv,nounits"))
	case "gpu_list_processes":
		text, err := runNvidiaSmi("--query-compute-apps=pid,process_name,used_gpu_memory", "--format=csv,nounits")
		if err == nil && len(strings.Split(strings.TrimSpace(text), "\n")) <= 1 {
			text = strings.TrimRight(text, "\n") + "\n(no compute processes are currently running on the GPU)"
		}
		return wrap(text, err)
	case "gpu_check_requirements":
		report, ok := checkRequirements()
		return textResult(report, !ok), nil
	case "gpu_run_command":
		return runUserCommand(p.Arguments), nil
	default:
		return nil, &rpcError{Code: -32602, Message: fmt.Sprintf(
			"unknown tool %q — available: gpu_get_status, gpu_query_metrics, gpu_list_processes, "+
				"gpu_check_requirements, gpu_run_command", p.Name)}
	}
}

func handle(req rpcRequest) rpcResponse {
	resp := rpcResponse{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		resp.Result = handleInitialize(req.Params)
	case "ping":
		resp.Result = map[string]any{}
	case "tools/list":
		resp.Result = map[string]any{"tools": tools}
	case "tools/call":
		result, rpcErr := handleToolsCall(req.Params)
		if rpcErr != nil {
			resp.Error = rpcErr
		} else {
			resp.Result = result
		}
	case "resources/list":
		resp.Result = map[string]any{"resources": []any{}}
	case "prompts/list":
		resp.Result = map[string]any{"prompts": []any{}}
	default:
		resp.Error = &rpcError{Code: -32601, Message: fmt.Sprintf("method %q not found", req.Method)}
	}
	return resp
}

// ---------------------------------------------------------------------------
// Main loop
// ---------------------------------------------------------------------------

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[gpu-mcp] ")

	// CLI mode: `gpu-mcp --check` lets developers verify the GPU requirement
	// right after installation, without an MCP client. Exit codes: 0 = met,
	// 1 = not met, 2 = usage error.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--check", "-check", "check":
			report, ok := checkRequirements()
			fmt.Print(report)
			if !ok {
				os.Exit(1)
			}
			return
		case "--version", "-version", "version":
			fmt.Printf("%s v%s (%s/%s)\n", serverName, serverVersion, runtime.GOOS, runtime.GOARCH)
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown argument %q\nusage: gpu-mcp [--check|--version]  "+
				"(no arguments: run as stdio MCP server)\n", os.Args[1])
			os.Exit(2)
		}
	}

	log.Printf("starting %s v%s (%s/%s)", serverName, serverVersion, runtime.GOOS, runtime.GOARCH)

	// Non-blocking preflight: log the GPU requirement result without delaying
	// the protocol loop (nvidia-smi may take a moment on cold start).
	go func() {
		report, ok := checkRequirements()
		if ok {
			log.Printf("GPU requirement met (%s)", requirementText())
		} else {
			log.Printf("WARNING — GPU requirement NOT met:\n%s", report)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	out := bufio.NewWriter(os.Stdout)
	enc := json.NewEncoder(out)

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			_ = enc.Encode(rpcResponse{JSONRPC: "2.0", ID: json.RawMessage("null"),
				Error: &rpcError{Code: -32700, Message: "parse error: " + err.Error()}})
			_ = out.Flush()
			continue
		}
		// Notifications (no id) do not get a response.
		if len(req.ID) == 0 || string(req.ID) == "null" {
			continue
		}
		_ = enc.Encode(handle(req))
		_ = out.Flush()
	}
	if err := scanner.Err(); err != nil {
		log.Printf("stdin error: %v", err)
	}
	log.Printf("stdin closed, shutting down")
}
