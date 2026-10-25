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

func TestNewRootCmd_ProcList(t *testing.T) {
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
	cmd.SetArgs([]string{"--config", configPath, "proc", "list"})

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

func TestNewRootCmd_AccListExcludesContent(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "acc", "list"})

	require.NoError(t, cmd.Execute())
	output := stdout.String()
	assert.Contains(t, output, `"artifact_id": "handlers.go"`)
	assert.Contains(t, output, `"state": "pending"`)
	assert.NotContains(t, output, `"pending":`)
	assert.NotContains(t, output, `"artifact_id": "service.go"`)
	assert.NotContains(t, output, `"content"`)
	assert.NotContains(t, output, "package handlers")
}

func TestNewRootCmd_AccListIncludeResolved(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "acc", "list", "--include-resolved"})

	require.NoError(t, cmd.Execute())
	output := stdout.String()
	assert.Contains(t, output, `"artifact_id": "handlers.go"`)
	assert.Contains(t, output, `"artifact_id": "service.go"`)
	assert.Contains(t, output, `"state": "resolved"`)
	assert.NotContains(t, output, `"pending":`)
}

func TestNewRootCmd_AccResolveMarksAcceleratorAsResolved(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	resolveCmd := newRootCmd(nil, nil)
	resolveCmd.SetArgs([]string{"--config", configPath, "acc", "resolve", "handlers.go"})
	require.NoError(t, resolveCmd.Execute())

	listCmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	listCmd.SetOut(&stdout)
	listCmd.SetArgs([]string{"--config", configPath, "acc", "list", "--include-resolved"})
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
	assert.Equal(t, "resolved", resolved.State)
}

func TestNewRootCmd_AccResolveReturnsErrorWhenAcceleratorDoesNotExist(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	cmd := newRootCmd(nil, nil)
	cmd.SetArgs([]string{"--config", configPath, "acc", "resolve", "missing.go"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `accelerator artifact "missing.go" not found`)
}

func TestNewRootCmd_AccShowExcludesContent(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "acc", "show", "handlers.go"})

	require.NoError(t, cmd.Execute())
	output := stdout.String()
	assert.Contains(t, output, `"artifact_id": "handlers.go"`)
	assert.Contains(t, output, `"state": "pending"`)
	assert.NotContains(t, output, `"pending":`)
	assert.NotContains(t, output, `"content"`)
	assert.NotContains(t, output, "package handlers")
}

func TestNewRootCmd_AccShowGuideReturnsComposedMarkdown(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "acc", "show", "handlers.go", "--guide"})

	require.NoError(t, cmd.Execute())
	output := stdout.String()
	assert.Contains(t, output, "# Update handlers")
	assert.Contains(t, output, "---\nNew accelerated content:")
	assert.Contains(t, output, "```go\npackage handlers")
}

func TestNewRootCmd_AccListAndShowForAgentOutputValidJSON(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	listCmd := newRootCmd(nil, nil)
	var listOut bytes.Buffer
	listCmd.SetOut(&listOut)
	listCmd.SetArgs([]string{"--config", configPath, "acc", "list", "--for-agent"})
	require.NoError(t, listCmd.Execute())

	var listPayload []map[string]any
	require.NoError(t, json.Unmarshal(listOut.Bytes(), &listPayload))
	require.Len(t, listPayload, 1)
	assert.Equal(t, "handlers.go", listPayload[0]["artifact_id"])
	assert.Equal(t, "guides/handlers.md", listPayload[0]["guide_path"])
	assert.Equal(t, "pending", listPayload[0]["state"])
	assert.NotContains(t, listPayload[0], "pending")

	getCmd := newRootCmd(nil, nil)
	var getOut bytes.Buffer
	getCmd.SetOut(&getOut)
	getCmd.SetArgs([]string{"--config", configPath, "acc", "show", "handlers.go", "--guide", "--for-agent"})
	require.NoError(t, getCmd.Execute())

	var getPayload map[string]any
	require.NoError(t, json.Unmarshal(getOut.Bytes(), &getPayload))
	assert.Equal(t, "handlers.go", getPayload["artifact_id"])
	assert.Equal(t, "guides/handlers.md", getPayload["guide_path"])
	assert.Equal(t, "pending", getPayload["state"])
	assert.NotContains(t, getPayload, "pending")
	assert.Equal(t, "# Update handlers", getPayload["guide_markdown"])
	assert.Equal(t, "package handlers", getPayload["accelerated_content"])
	assert.Contains(t, getPayload["composed_markdown"], "New accelerated content")
}

func TestNewRootCmd_AccGuide(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	for _, artifact := range []struct {
		id       string
		path     string
		markdown string
	}{
		{"handlers.go", "guides/handlers.md", "# Update handlers"},
		{"service.go", "guides/service.md", "# Update service"},
	} {
		for _, forAgent := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/for-agent=%t", artifact.id, forAgent), func(t *testing.T) {
				cmd := newRootCmd(nil, nil)
				var stdout bytes.Buffer
				cmd.SetOut(&stdout)
				args := []string{"--config", configPath, "acc", "guide", artifact.id}
				if forAgent {
					args = append(args, "--for-agent")
				}
				cmd.SetArgs(args)
				require.NoError(t, cmd.Execute())

				if forAgent {
					var payload map[string]any
					require.NoError(t, json.Unmarshal(stdout.Bytes(), &payload))
					assert.Equal(t, map[string]any{"path": artifact.path, "markdown": artifact.markdown}, payload)
				} else {
					assert.Equal(t, artifact.markdown, stdout.String())
				}
			})
		}
	}
}

func TestNewRootCmd_AccGuideErrors(t *testing.T) {
	for _, test := range []struct {
		name       string
		artifactID string
		guidePath  string
		wantError  string
	}{
		{"unknown artifact", "missing.go", "guides/handlers.md", `accelerator artifact "missing.go" not found`},
		{"guide path is not an artifact ID", "guides/handlers.md", "guides/handlers.md", `accelerator artifact "guides/handlers.md" not found`},
		{"no attached guide", "handlers.go", "", `accelerator artifact "handlers.go" has no attached guide`},
		{"missing guide file", "handlers.go", "guides/missing.md", "guide file not found: guides/missing.md"},
	} {
		t.Run(test.name, func(t *testing.T) {
			configPath, ccodePath := setupAcceleratorCLIProject(t)
			statePath := filepath.Join(ccodePath, ".ccode", "accelerators", "handlers.go.accelerated.json")
			writeAcceleratorStateFileForCLITest(t, statePath, true, test.guidePath, "package handlers", "")

			cmd := newRootCmd(nil, nil)
			var stdout, stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs([]string{"--config", configPath, "acc", "guide", test.artifactID})
			require.ErrorContains(t, cmd.Execute(), test.wantError)
			assert.NotContains(t, stdout.String(), "# Update handlers")
		})
	}
}

func TestNewRootCmd_RejectsReplacedCommands(t *testing.T) {
	for _, args := range [][]string{
		{"list", "processes"},
		{"list", "acc"},
		{"get", "accelerated", "handlers.go"},
		{"get", "guide", "guides/handlers.md"},
		{"resolve", "handlers.go"},
	} {
		t.Run(fmt.Sprint(args), func(t *testing.T) {
			cmd := newRootCmd(nil, nil)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(args)

			err := cmd.Execute()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unknown command")
		})
	}
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
