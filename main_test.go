package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

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

	t.Setenv("CCODE_CCODE_PATH", "")
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

func TestNewRootCmd_ListProcesses(t *testing.T) {
	tmp := t.TempDir()
	projectPath := filepath.Join(tmp, "ccode")
	configPath := filepath.Join(tmp, "ccode.yaml")
	configYAML := `ccode_path: ccode
output_path: .
hidden_path: internal-support
version: v1.2.3
`
	require.NoError(t, os.WriteFile(configPath, []byte(configYAML), 0644))

	paths := []string{
		"generate.process.ts",
		"api/client.process.ts",
		"api/helper.ts",
		".hidden/ignored.process.ts",
		"internal-support/ignored.process.ts",
	}
	for _, path := range paths {
		fullPath := filepath.Join(projectPath, filepath.FromSlash(path))
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte("export default function () {}\n"), 0644))
	}

	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "list", "processes"})

	require.NoError(t, cmd.Execute())
	assert.JSONEq(t, `["api/client", "generate"]`, stdout.String())
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
	assert.Contains(t, output, `"artifact_id": "handlers.go"`)
	assert.Contains(t, output, `"pending": true`)
	assert.NotContains(t, output, `"artifact_id": "service.go"`)
	assert.NotContains(t, output, `"content"`)
	assert.NotContains(t, output, "package handlers")
}

func TestNewRootCmd_ListAcceleratedIncludeResolved(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "list", "accelerated", "--include-resolved"})

	require.NoError(t, cmd.Execute())
	output := stdout.String()
	assert.Contains(t, output, `"artifact_id": "handlers.go"`)
	assert.Contains(t, output, `"artifact_id": "service.go"`)
	assert.Contains(t, output, `"state": "resolved"`)
}

func TestNewRootCmd_ResolveMarksAcceleratorAsResolved(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	resolveCmd := newRootCmd(nil, nil)
	resolveCmd.SetArgs([]string{"--config", configPath, "resolve", "handlers.go"})
	require.NoError(t, resolveCmd.Execute())

	listCmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	listCmd.SetOut(&stdout)
	listCmd.SetArgs([]string{"--config", configPath, "list", "accelerated", "--include-resolved"})
	require.NoError(t, listCmd.Execute())

	var items []ccode.AcceleratorArtifactMetadata
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &items))

	var resolved *ccode.AcceleratorArtifactMetadata
	for index := range items {
		if items[index].ArtifactID == "handlers.go" {
			resolved = &items[index]
			break
		}
	}
	require.NotNil(t, resolved)
	assert.False(t, resolved.Pending)
	assert.Equal(t, "resolved", resolved.State)
}

func TestNewRootCmd_ResolveReturnsErrorWhenAcceleratorDoesNotExist(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	cmd := newRootCmd(nil, nil)
	cmd.SetArgs([]string{"--config", configPath, "resolve", "missing.go"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `accelerator artifact "missing.go" not found`)
}

func TestNewRootCmd_GetAcceleratedExcludesContent(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "get", "accelerated", "handlers.go"})

	require.NoError(t, cmd.Execute())
	output := stdout.String()
	assert.Contains(t, output, `"artifact_id": "handlers.go"`)
	assert.Contains(t, output, `"pending": true`)
	assert.NotContains(t, output, `"content"`)
	assert.NotContains(t, output, "package handlers")
}

func TestNewRootCmd_GetAcceleratedGuideReturnsComposedMarkdown(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "get", "accelerated", "handlers.go", "--guide"})

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
	assert.Equal(t, "handlers.go", listPayload[0]["artifact_id"])
	assert.Equal(t, "guides/handlers.md", listPayload[0]["guide_path"])
	assert.Equal(t, true, listPayload[0]["pending"])

	getCmd := newRootCmd(nil, nil)
	var getOut bytes.Buffer
	getCmd.SetOut(&getOut)
	getCmd.SetArgs([]string{"--config", configPath, "get", "accelerated", "handlers.go", "--guide", "--for-agent"})
	require.NoError(t, getCmd.Execute())

	var getPayload map[string]any
	require.NoError(t, json.Unmarshal(getOut.Bytes(), &getPayload))
	assert.Equal(t, "handlers.go", getPayload["artifact_id"])
	assert.Equal(t, "guides/handlers.md", getPayload["guide_path"])
	assert.Equal(t, "# Update handlers", getPayload["guide_markdown"])
	assert.Equal(t, "package handlers", getPayload["accelerated_content"])
	assert.Contains(t, getPayload["composed_markdown"], "New accelerated content")
}

func TestNewRootCmd_GetGuide(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	getCmd := newRootCmd(nil, nil)
	var getOut bytes.Buffer
	getCmd.SetOut(&getOut)
	getCmd.SetArgs([]string{"--config", configPath, "get", "guide", "guides/handlers.md"})
	require.NoError(t, getCmd.Execute())
	assert.Equal(t, "# Update handlers", getOut.String())
}

func setupAcceleratorCLIProject(t *testing.T) (string, string) {
	t.Helper()

	rootDir := t.TempDir()
	ccodePath := filepath.Join(rootDir, "ccode")
	require.NoError(t, os.MkdirAll(filepath.Join(ccodePath, "guides"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ccodePath, "guides", "handlers.md"), []byte("# Update handlers"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(ccodePath, "guides", "service.md"), []byte("# Update service"), 0644))

	configPath := filepath.Join(rootDir, "ccode.yaml")
	outputPath := filepath.Join(rootDir, "out")
	configYAML := "ccode_path: " + ccodePath + "\noutput_path: " + outputPath + "\nversion: v1.2.3\n"
	require.NoError(t, os.WriteFile(configPath, []byte(configYAML), 0644))
	require.NoError(t, os.MkdirAll(outputPath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(outputPath, "handlers.go"), []byte("package handlers"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(outputPath, "service.go"), []byte("package service"), 0644))

	stateDir := filepath.Join(ccodePath, ".ccode", "accelerators")
	require.NoError(t, os.MkdirAll(stateDir, 0755))
	writeAcceleratorStateFileForCLITest(t, filepath.Join(stateDir, "handlers.go.accelerated.json"), true, "guides/handlers.md", "package handlers", "# Update handlers")
	writeAcceleratorStateFileForCLITest(t, filepath.Join(stateDir, "service.go.accelerated.json"), false, "guides/service.md", "package service", "# Update service")

	return configPath, ccodePath
}

func encodeSnapshotForCLITest(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}

func writeAcceleratorStateFileForCLITest(t *testing.T, path string, pending bool, guidePath string, code string, guide string) {
	t.Helper()

	type stateFile struct {
		Pending             bool   `json:"pending"`
		Guide               string `json:"guide"`
		AcceleratedChecksum string `json:"accelerated_checksum"`
		GuideChecksum       string `json:"guide_checksum"`
		Code                string `json:"code"`
	}

	state := stateFile{
		Pending:             pending,
		Guide:               guidePath,
		AcceleratedChecksum: checksumForCLITest(code),
		Code:                encodeSnapshotForCLITest(code),
	}
	if guide != "" {
		state.GuideChecksum = checksumForCLITest(guide)
	}

	payload, err := json.Marshal(state)
	require.NoError(t, err)
	payload = append(payload, '\n')
	require.NoError(t, os.WriteFile(path, payload, 0644))
}

func checksumForCLITest(content string) string {
	checksum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("sha256:%x", checksum)
}
