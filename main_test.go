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
	"gopkg.in/yaml.v3"
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

func TestNewRootCmd_ProcAdd(t *testing.T) {
	for _, override := range []string{"", "flag", "env"} {
		t.Run("config-override-"+override, func(t *testing.T) {
			tmp := t.TempDir()
			configPath := filepath.Join(tmp, "ccode.yaml")
			require.NoError(t, os.WriteFile(configPath, []byte("ccode_path: custom-ccode\nversion: v1.2.3\n"), 0644))
			t.Setenv("CCODE_CCODE_PATH", "")
			require.NoError(t, os.Unsetenv("CCODE_CCODE_PATH"))
			projectPath := filepath.Join(tmp, "custom-ccode")
			args := []string{"--config", configPath}
			switch override {
			case "flag":
				projectPath = filepath.Join(tmp, "flag-ccode")
				args = append(args, "--ccode-path", projectPath)
			case "env":
				projectPath = filepath.Join(tmp, "env-ccode")
				t.Setenv("CCODE_CCODE_PATH", projectPath)
			}
			cmd := newRootCmd(nil, nil)
			cmd.SetArgs(append(args, "proc", "add", "handlers/generator"))
			require.NoError(t, cmd.Execute())

			content, err := os.ReadFile(filepath.Join(projectPath, "handlers", "generator.proc.ts"))
			require.NoError(t, err)
			assert.Equal(t, "import type { Context } from \"@ccode/context\";\n\nexport default function run(ctx: Context) {\n\n}\n", string(content))
		})
	}
}

func TestNewRootCmd_ProcAddRequiresOneArgument(t *testing.T) {
	for _, args := range [][]string{{"proc", "add"}, {"proc", "add", "one", "two"}} {
		cmd := newRootCmd(nil, nil)
		cmd.SetArgs(args)
		require.Error(t, cmd.Execute())
	}
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
		"generate.proc.ts",
		"api/client.proc.ts",
		"api/helper.ts",
		"api/legacy.process.ts",
		".hidden/ignored.proc.ts",
		"internal-support/ignored.proc.ts",
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
	assert.Equal(t, "- artifact_id: handlers.go\n  guide_path: guides/handlers.md\n  state: pending\n", output)
}

func TestNewRootCmd_AccListIncludeResolved(t *testing.T) {
	for _, asJSON := range []bool{false, true} {
		t.Run(fmt.Sprintf("json=%t", asJSON), func(t *testing.T) {
			configPath, _ := setupAcceleratorCLIProject(t)
			cmd := newRootCmd(nil, nil)
			var stdout bytes.Buffer
			cmd.SetOut(&stdout)
			args := []string{"--config", configPath, "acc", "list", "--include-resolved"}
			if asJSON {
				args = append(args, "--json")
			}
			cmd.SetArgs(args)
			require.NoError(t, cmd.Execute())
			var items []map[string]any
			if asJSON {
				require.NoError(t, json.Unmarshal(stdout.Bytes(), &items))
			} else {
				require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &items))
			}
			assert.Equal(t, []map[string]any{
				{"artifact_id": "handlers.go", "guide_path": "guides/handlers.md", "state": "pending"},
				{"artifact_id": "service.go", "guide_path": "guides/service.md", "state": "resolved"},
			}, items)
		})
	}
}

func TestNewRootCmd_AccListEmpty(t *testing.T) {
	for _, asJSON := range []bool{false, true} {
		t.Run(fmt.Sprintf("json=%t", asJSON), func(t *testing.T) {
			configPath, ccodePath := setupAcceleratorCLIProject(t)
			require.NoError(t, os.RemoveAll(filepath.Join(ccodePath, ".ccode", "accelerators")))
			cmd := newRootCmd(nil, nil)
			var stdout bytes.Buffer
			cmd.SetOut(&stdout)
			args := []string{"--config", configPath, "acc", "list"}
			if asJSON {
				args = append(args, "--json")
			}
			cmd.SetArgs(args)
			require.NoError(t, cmd.Execute())
			assert.Equal(t, "[]\n", stdout.String())
		})
	}
}

func TestNewRootCmd_AccResolveMarksAcceleratorAsResolved(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)

	resolveCmd := newRootCmd(nil, nil)
	resolveCmd.SetArgs([]string{"--config", configPath, "acc", "resolve", "handlers.go"})
	require.NoError(t, resolveCmd.Execute())

	listCmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	listCmd.SetOut(&stdout)
	listCmd.SetArgs([]string{"--config", configPath, "acc", "list", "--include-resolved", "--json"})
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
	assert.Equal(t, "artifact_id: handlers.go\nguide_path: guides/handlers.md\nstate: pending\n", output)
}

func TestNewRootCmd_AccShowJSON(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)
	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "acc", "show", "handlers.go", "--json"})
	require.NoError(t, cmd.Execute())
	var metadata map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &metadata))
	assert.Equal(t, map[string]any{
		"artifact_id": "handlers.go", "guide_path": "guides/handlers.md", "state": "pending",
	}, metadata)
}

func TestNewRootCmd_AccShowYAMLPreservesDiagnostics(t *testing.T) {
	configPath, ccodePath := setupAcceleratorCLIProject(t)
	statePath := filepath.Join(ccodePath, ".ccode", "accelerators", "handlers.go.accelerated.json")
	require.NoError(t, os.WriteFile(statePath, []byte("invalid json\n"), 0644))
	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "acc", "show", "handlers.go"})
	require.NoError(t, cmd.Execute())
	var metadata map[string]any
	require.NoError(t, yaml.Unmarshal(stdout.Bytes(), &metadata))
	assert.Equal(t, "handlers.go", metadata["artifact_id"])
	assert.Contains(t, metadata, "guide_path")
	assert.Nil(t, metadata["guide_path"])
	assert.Equal(t, "corrupt", metadata["state"])
	assert.NotEmpty(t, metadata["message"])
	assert.NotContains(t, metadata, "pending")
	assert.NotContains(t, metadata, "content")
}

func TestNewRootCmd_AccListJSON(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)
	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "acc", "list", "--json"})
	require.NoError(t, cmd.Execute())
	var items []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &items))
	assert.Equal(t, []map[string]any{
		{"artifact_id": "handlers.go", "guide_path": "guides/handlers.md", "state": "pending"},
	}, items)
}

func TestNewRootCmd_AccRejectsRemovedOutputFlags(t *testing.T) {
	for _, args := range [][]string{
		{"acc", "list", "--for-agent"},
		{"acc", "show", "handlers.go", "--for-agent"},
		{"acc", "guide", "handlers.go", "--for-agent"},
		{"acc", "show", "handlers.go", "--guide"},
		{"acc", "guide", "handlers.go", "--json"},
	} {
		t.Run(fmt.Sprint(args), func(t *testing.T) {
			cmd := newRootCmd(nil, nil)
			var stdout, stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SilenceUsage = true
			cmd.SetArgs(args)
			require.ErrorContains(t, cmd.Execute(), "unknown flag: "+args[len(args)-1])
			assert.Empty(t, stdout.String())
		})
	}
}

func TestNewRootCmd_AccContentReturnsSavedSnapshot(t *testing.T) {
	for _, test := range []struct {
		name         string
		artifactID   string
		content      string
		pending      bool
		guidePath    string
		removeOutput bool
	}{
		{"pending", "handlers.go", "package handlers", true, "guides/handlers.md", false},
		{"resolved", "service.go", "package service", false, "", false},
		{"whitespace and unicode", "handlers.go", "\t// café\r\npackage handlers\n\n", true, "guides/handlers.md", false},
		{"empty content", "handlers.go", "", true, "guides/handlers.md", false},
		{"no guide", "handlers.go", "package handlers", true, "", false},
		{"missing guide", "handlers.go", "package handlers", true, "guides/missing.md", false},
		{"missing output", "handlers.go", "package handlers", true, "guides/handlers.md", true},
		{"nested artifact", "nested/handlers.go", "package handlers", true, "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			configPath, ccodePath := setupAcceleratorCLIProject(t)
			statePath := filepath.Join(ccodePath, ".ccode", "accelerators", filepath.FromSlash(test.artifactID)+".accelerated.json")
			require.NoError(t, os.MkdirAll(filepath.Dir(statePath), 0755))
			writeAcceleratorStateFileForCLITest(t, statePath, test.pending, test.guidePath, test.content, "")
			outputPath := filepath.Join(filepath.Dir(configPath), "out", filepath.FromSlash(test.artifactID))
			if test.removeOutput {
				if err := os.Remove(outputPath); err != nil {
					require.True(t, os.IsNotExist(err))
				}
			} else {
				require.NoError(t, os.WriteFile(outputPath, []byte("// edited output"), 0644))
			}

			cmd := newRootCmd(nil, nil)
			var stdout bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetArgs([]string{"--config", configPath, "acc", "content", test.artifactID})

			require.NoError(t, cmd.Execute())
			assert.Equal(t, test.content, stdout.String())
		})
	}
}

func TestNewRootCmd_AccContentErrors(t *testing.T) {
	for _, test := range []struct {
		name      string
		args      []string
		state     string
		wantError string
	}{
		{"unknown artifact", []string{"missing.go"}, "", `accelerator artifact "missing.go" not found`},
		{"missing argument", nil, "", "accepts 1 arg(s), received 0"},
		{"extra argument", []string{"handlers.go", "service.go"}, "", "accepts 1 arg(s), received 2"},
		{"corrupt state", []string{"handlers.go"}, "invalid json\n", "corrupt"},
		{"ambiguous state", []string{"handlers.go"}, "{\"pending\":true}\n{\"pending\":false}\n", "ambiguous"},
		{"invalid snapshot", []string{"handlers.go"}, "{\"code\":\"not base64!\"}\n", "decode content snapshot payload"},
	} {
		t.Run(test.name, func(t *testing.T) {
			configPath, ccodePath := setupAcceleratorCLIProject(t)
			if test.state != "" {
				statePath := filepath.Join(ccodePath, ".ccode", "accelerators", "handlers.go.accelerated.json")
				require.NoError(t, os.WriteFile(statePath, []byte(test.state), 0644))
			}

			cmd := newRootCmd(nil, nil)
			var stdout, stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(append([]string{"--config", configPath, "acc", "content"}, test.args...))
			cmd.SilenceUsage = true

			require.ErrorContains(t, cmd.Execute(), test.wantError)
			assert.Empty(t, stdout.String())
		})
	}
}

func TestNewRootCmd_AccDiff(t *testing.T) {
	configPath, ccodePath := setupAcceleratorCLIProject(t)
	statePath := filepath.Join(ccodePath, ".ccode", "accelerators", "handlers.go.accelerated.json")
	writeAcceleratorStateFileForCLITest(t, statePath, true, "", "package original\n", "")
	outputPath := filepath.Join(filepath.Dir(configPath), "out", "handlers.go")
	require.NoError(t, os.WriteFile(outputPath, []byte("package edited\n"), 0644))

	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "acc", "diff", "handlers.go"})

	require.NoError(t, cmd.Execute())
	assert.Equal(t, "--- a/handlers.go\n+++ b/handlers.go\n@@ -1 +1 @@\n-package original\n+package edited\n", stdout.String())
}

func TestNewRootCmd_AccDiffUnchanged(t *testing.T) {
	configPath, _ := setupAcceleratorCLIProject(t)
	cmd := newRootCmd(nil, nil)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--config", configPath, "acc", "diff", "service.go"})

	require.NoError(t, cmd.Execute())
	assert.Empty(t, stdout.String())
}

func TestNewRootCmd_AccDiffErrors(t *testing.T) {
	for _, test := range []struct {
		name      string
		args      []string
		wantError string
	}{
		{"unknown artifact", []string{"missing.go"}, `accelerator artifact "missing.go" not found`},
		{"missing argument", nil, "accepts 1 arg(s), received 0"},
		{"extra argument", []string{"handlers.go", "service.go"}, "accepts 1 arg(s), received 2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			configPath, _ := setupAcceleratorCLIProject(t)
			cmd := newRootCmd(nil, nil)
			var stdout, stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SilenceUsage = true
			cmd.SetArgs(append([]string{"--config", configPath, "acc", "diff"}, test.args...))

			require.ErrorContains(t, cmd.Execute(), test.wantError)
			assert.Empty(t, stdout.String())
		})
	}
}

func TestNewRootCmd_AccGuide(t *testing.T) {
	configPath, ccodePath := setupAcceleratorCLIProject(t)
	for _, artifact := range []struct {
		id       string
		path     string
		markdown string
	}{
		{"handlers.go", "guides/handlers.md", "# Update handlers"},
		{"service.go", "guides/service.md", "# Update service\r\n\r\nPreserve café naming.\r\n"},
	} {
		t.Run(artifact.id, func(t *testing.T) {
			require.NoError(t, os.WriteFile(filepath.Join(ccodePath, artifact.path), []byte(artifact.markdown), 0644))
			cmd := newRootCmd(nil, nil)
			var stdout bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetArgs([]string{"--config", configPath, "acc", "guide", artifact.id})
			require.NoError(t, cmd.Execute())
			assert.Equal(t, artifact.markdown, stdout.String())
		})
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
