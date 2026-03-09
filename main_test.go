package main

import (
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
	configYAML := `path: "from-config"
output_path: "build/from-config"
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
	assert.Equal(t, "from-config", captured.Path)
	assert.Equal(t, "build/from-config", captured.OutputPath)
}

func TestNewRootCmd_FlagsOverrideConfig(t *testing.T) {
	tmp := t.TempDir()
	cfgFile := filepath.Join(tmp, "ccode.yaml")
	configYAML := `path: "from-config"
output_path: "build/from-config"
`
	require.NoError(t, os.WriteFile(cfgFile, []byte(configYAML), 0644))

	var captured *ccode.Config

	cmd := newRootCmd(func(cfg *ccode.Config, proc string) error {
		captured = cfg
		return nil
	}, nil)
	cmd.SetArgs([]string{
		"--config", cfgFile,
		"--path", "from-cli",
		"--output-path", "build/from-cli",
		"run", "openapi/generate",
	})

	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)
	assert.Equal(t, "from-cli", captured.Path)
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
	assert.Equal(t, "ccode", captured.Path)
	assert.Equal(t, ".", captured.OutputPath)
}

func TestNewRootCmd_InitUsesArgsAndConfigFlag(t *testing.T) {
	var capturedProjectPath string
	var capturedConfigPath string

	cmd := newRootCmd(nil, func(projectPath string, configPath string) error {
		capturedProjectPath = projectPath
		capturedConfigPath = configPath
		return nil
	})
	cmd.SetArgs([]string{"--config", "custom/ccode.yaml", "init", "cohesive"})

	require.NoError(t, cmd.Execute())
	assert.Equal(t, "cohesive", capturedProjectPath)
	assert.Equal(t, "custom/ccode.yaml", capturedConfigPath)
}
