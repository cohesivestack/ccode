package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	ccode "github.com/cohesivestack/ccode/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmd_LoadsConfigFile(t *testing.T) {
	tmp := t.TempDir()
	cfgFile := filepath.Join(tmp, "ccode.yaml")
	configYAML := `ccode_path: "from-config"
output_path: "build/from-config"
version: "v1.2.3"
`
	require.NoError(t, os.WriteFile(cfgFile, []byte(configYAML), 0644))

	var captured *ccode.Config
	var process string

	cmd := newRootCmd(func(cfg *ccode.Config, proc string) error {
		captured = cfg
		process = proc
		return nil
	}, nil)
	cmd.SetArgs([]string{"--config", cfgFile, "run", "openapi/generate"})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)
	assert.Equal(t, "openapi/generate", process)
	assert.Equal(t, filepath.Join(tmp, "from-config"), captured.CCodePath)
	assert.Equal(t, "build/from-config", captured.OutputPath)
}

func TestNewRootCmd_FlagsOverrideConfig(t *testing.T) {
	tmp := t.TempDir()
	cfgFile := filepath.Join(tmp, "ccode.yaml")
	configYAML := `ccode_path: "from-config"
output_path: "build/from-config"
version: "v1.2.3"
`
	require.NoError(t, os.WriteFile(cfgFile, []byte(configYAML), 0644))

	var captured *ccode.Config

	cmd := newRootCmd(func(cfg *ccode.Config, proc string) error {
		captured = cfg
		return nil
	}, nil)
	cmd.SetArgs([]string{
		"--config", cfgFile,
		"--ccode-path", "from-cli",
		"--output-path", "build/from-cli",
		"run", "openapi/generate",
	})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)
	assert.Equal(t, "from-cli", captured.CCodePath)
	assert.Equal(t, "build/from-cli", captured.OutputPath)
}

func TestNewRootCmd_DefaultConfigWhenNoFile(t *testing.T) {
	originalWD, err := os.Getwd()
	require.NoError(t, err)

	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	t.Setenv("CCODE_PATH", "")
	t.Setenv("CCODE_OUTPUT_PATH", "")

	var captured *ccode.Config

	cmd := newRootCmd(func(cfg *ccode.Config, proc string) error {
		captured = cfg
		return nil
	}, nil)
	cmd.SetArgs([]string{"run", "openapi/generate"})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)
	assert.Equal(t, "ccode", captured.CCodePath)
	assert.Equal(t, ".", captured.OutputPath)
}

func TestNewRootCmd_InitUsesArgsAndConfigFlag(t *testing.T) {
	var capturedProjectPath string
	var capturedConfigPath string
	var capturedVersion string

	cmd := newRootCmd(nil, func(projectPath string, configPath string, version string) error {
		capturedProjectPath = projectPath
		capturedConfigPath = configPath
		capturedVersion = version
		return nil
	})
	cmd.SetArgs([]string{"--config", "custom/ccode.yaml", "init", "cohesive", "--version", "v1.2.3"})

	require.NoError(t, cmd.Execute())
	assert.Equal(t, "cohesive", capturedProjectPath)
	assert.Equal(t, "custom/ccode.yaml", capturedConfigPath)
	assert.Equal(t, "v1.2.3", capturedVersion)
}

func TestNewRootCmd_ListAcceleratedExcludesContent(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "list", "accelerated"})

	require.NoError(t, cmd.Execute())
	output := stdout.String()
	assert.Contains(t, output, `"scope_id": "generate-api"`)
	assert.Contains(t, output, `"artifact_id": "handlers.go"`)
	assert.NotContains(t, output, `"content"`)
	assert.NotContains(t, output, "package handlers")
}

func TestNewRootCmd_GetAcceleratedExcludesContent(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "get", "accelerated", "generate-api:handlers.go"})

	require.NoError(t, cmd.Execute())
	output := stdout.String()
	assert.Contains(t, output, `"scope_id": "generate-api"`)
	assert.Contains(t, output, `"artifact_id": "handlers.go"`)
	assert.NotContains(t, output, `"content"`)
	assert.NotContains(t, output, "package handlers")
}

func TestNewRootCmd_GetAcceleratedInstructionsReturnsComposedMarkdown(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "get", "accelerated", "generate-api:handlers.go", "--instructions"})

	require.NoError(t, cmd.Execute())
	output := stdout.String()
	assert.Contains(t, output, "# Update handlers")
	assert.Contains(t, output, "---\nNew accelerated content:")
	assert.Contains(t, output, "```go\npackage handlers")
}

func TestNewRootCmd_ListAndGetForAgentOutputValidJSON(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	listCmd := newRootCmd(nil, nil)
	var listOut bytes.Buffer
	listCmd.SetOut(&listOut)
	listCmd.SetArgs([]string{"--config", configPath, "list", "accelerated", "--for-agent"})
	require.NoError(t, listCmd.Execute())

	var listPayload []map[string]any
	require.NoError(t, json.Unmarshal(listOut.Bytes(), &listPayload))
	require.Len(t, listPayload, 1)
	assert.Equal(t, "generate-api", listPayload[0]["scope_id"])
	assert.Equal(t, "handlers.go", listPayload[0]["artifact_id"])

	getCmd := newRootCmd(nil, nil)
	var getOut bytes.Buffer
	getCmd.SetOut(&getOut)
	getCmd.SetArgs([]string{"--config", configPath, "get", "accelerated", "generate-api:handlers.go", "--instructions", "--for-agent"})
	require.NoError(t, getCmd.Execute())

	var getPayload map[string]any
	require.NoError(t, json.Unmarshal(getOut.Bytes(), &getPayload))
	assert.Equal(t, "generate-api", getPayload["scope_id"])
	assert.Equal(t, "handlers.go", getPayload["artifact_id"])
	assert.Equal(t, "package handlers", getPayload["accelerated_content"])
	assert.Contains(t, getPayload["composed_markdown"], "New accelerated content")
}

func TestNewRootCmd_ListInstructionsAndGetInstruction(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	listCmd := newRootCmd(nil, nil)
	var listOut bytes.Buffer
	listCmd.SetOut(&listOut)
	listCmd.SetArgs([]string{"--config", configPath, "list", "instructions"})
	require.NoError(t, listCmd.Execute())
	assert.Contains(t, listOut.String(), `"instructions_path": "instructions/handlers.md"`)
	assert.NotContains(t, listOut.String(), "package handlers")

	getCmd := newRootCmd(nil, nil)
	var getOut bytes.Buffer
	getCmd.SetOut(&getOut)
	getCmd.SetArgs([]string{"--config", configPath, "get", "instruction", "instructions/handlers.md"})
	require.NoError(t, getCmd.Execute())
	assert.Equal(t, "# Update handlers", getOut.String())
}

func TestParseAcceleratorSelector(t *testing.T) {
	scopeID, artifactID, err := parseAcceleratorSelector("generate-api:handlers.go")
	require.NoError(t, err)
	assert.Equal(t, "generate-api", scopeID)
	assert.Equal(t, "handlers.go", artifactID)

	_, _, err = parseAcceleratorSelector("invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected <scopeId>:<artifactId>")

	_, _, err = parseAcceleratorSelector("a:b:c")
	require.Error(t, err)
}

func setupAcceleratorCLIProject(t *testing.T) (string, string) {
	t.Helper()

	rootDir := t.TempDir()
	ccodePath := filepath.Join(rootDir, "ccode")
	require.NoError(t, os.MkdirAll(filepath.Join(ccodePath, "instructions"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ccodePath, "instructions", "handlers.md"), []byte("# Update handlers"), 0644))

	configPath := filepath.Join(rootDir, "ccode.yaml")
	configYAML := "ccode_path: " + ccodePath + "\noutput_path: " + filepath.Join(rootDir, "out") + "\nversion: v1.2.3\n"
	require.NoError(t, os.WriteFile(configPath, []byte(configYAML), 0644))

	type stateArtifact struct {
		ID               string  `json:"id"`
		Content          string  `json:"content"`
		AdjustedAt       *string `json:"adjusted_at"`
		InstructionsPath *string `json:"instructions_path"`
	}
	type stateScope struct {
		ID        string          `json:"id"`
		Artifacts []stateArtifact `json:"artifacts"`
	}
	type stateRoot struct {
		Version int          `json:"version"`
		Scopes  []stateScope `json:"scopes"`
	}

	adjustedAt := "2026-04-11T00:00:00Z"
	instructionsPath := "instructions/handlers.md"
	state := stateRoot{
		Version: 1,
		Scopes: []stateScope{
			{
				ID: "generate-api",
				Artifacts: []stateArtifact{
					{
						ID:               "handlers.go",
						Content:          encodeSnapshotForCLITest("package handlers"),
						AdjustedAt:       nil,
						InstructionsPath: &instructionsPath,
					},
					{
						ID:         "service.go",
						Content:    encodeSnapshotForCLITest("package service"),
						AdjustedAt: &adjustedAt,
					},
				},
			},
		},
	}

	stateBytes, err := json.MarshalIndent(state, "", "  ")
	require.NoError(t, err)
	stateBytes = append(stateBytes, '\n')
	statePath := filepath.Join(ccodePath, ".ccode", "state", "accelerators.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(statePath), 0755))
	require.NoError(t, os.WriteFile(statePath, stateBytes, 0644))

	return configPath, ccodePath
}

func encodeSnapshotForCLITest(content string) string {
	return time.Now().UTC().Format(time.RFC3339) + ":" + base64.StdEncoding.EncodeToString([]byte(content))
}
