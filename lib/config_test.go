package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_NewConfigDefaults(t *testing.T) {
	cfg, err := NewConfig(&Config{})
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "ccode", cfg.CCodePath)
	assert.Equal(t, ".", cfg.OutputPath)
}

func TestConfig_LoadConfig(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "configs")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	cfgFile := filepath.Join(configDir, "ccode.yaml")
	content := `ccode_path: "my-ccode"
output_path: "dist"
version: "v1.2.3"
`
	require.NoError(t, os.WriteFile(cfgFile, []byte(content), 0644))

	cfg, err := LoadConfig(cfgFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, filepath.Join(configDir, "my-ccode"), cfg.CCodePath)
	assert.Equal(t, "dist", cfg.OutputPath)
	assert.Equal(t, "v1.2.3", cfg.Version)
}

func TestConfig_LoadConfigRejectsUnknownKeys(t *testing.T) {
	tmp := t.TempDir()
	cfgFile := filepath.Join(tmp, "ccode.yaml")
	content := `path: "old-ccode"
version: "v1.2.3"
`
	require.NoError(t, os.WriteFile(cfgFile, []byte(content), 0644))

	cfg, err := LoadConfig(cfgFile)
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "field path not found")
}

func TestConfig_LoadConfigRequiresVersion(t *testing.T) {
	tmp := t.TempDir()
	cfgFile := filepath.Join(tmp, "ccode.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte(`ccode_path: "ccode"`), 0644))

	cfg, err := LoadConfig(cfgFile)
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "version is required")
}

func TestConfig_LoadConfigMissingFile(t *testing.T) {
	cfg, err := LoadConfig("missing.yaml")
	require.Error(t, err)
	assert.Nil(t, cfg)
}
