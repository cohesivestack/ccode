package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig_Defaults(t *testing.T) {
	cfg, err := NewConfig(&Config{})
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "ccode", cfg.CCodePath)
	assert.Equal(t, ".", cfg.OutputPath)
}

func TestLoadConfig(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "configs")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	cfgFile := filepath.Join(configDir, "ccode.yaml")
	content := `ccode_path: "my-ccode"
output_path: "dist"
`
	require.NoError(t, os.WriteFile(cfgFile, []byte(content), 0644))

	cfg, err := LoadConfig(cfgFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, filepath.Join(configDir, "my-ccode"), cfg.CCodePath)
	assert.Equal(t, "dist", cfg.OutputPath)
}

func TestLoadConfig_LegacyPathKey(t *testing.T) {
	tmp := t.TempDir()
	cfgFile := filepath.Join(tmp, "ccode.yaml")
	content := `path: "legacy-ccode"
`
	require.NoError(t, os.WriteFile(cfgFile, []byte(content), 0644))

	cfg, err := LoadConfig(cfgFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, filepath.Join(tmp, "legacy-ccode"), cfg.CCodePath)
}

func TestLoadConfig_MissingFile(t *testing.T) {
	cfg, err := LoadConfig("missing.yaml")
	require.Error(t, err)
	assert.Nil(t, cfg)
}
