package ccode

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccelerator_AccelerateWritesArtifactAndScopedState(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, ctx.SetScope("generate-api"))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Hello {{ data.name }}!"), 0644))

	value, err := ctx.runtime.RunString(`({ name: "Carlos" })`)
	require.NoError(t, err)

	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	outputFilePath := filepath.Join(ctx.ccodeContext.config.OutputPath, "handlers")
	require.FileExists(t, outputFilePath)

	content, err := os.ReadFile(outputFilePath)
	require.NoError(t, err)
	assert.Equal(t, "Hello Carlos!", string(content))

	statePath := filepath.Join(ctx.ccodeContext.config.CCodePath, DefaultHiddenFolderName, "state", "accelerators.json")
	require.FileExists(t, statePath)

	state, err := loadAcceleratorState(statePath)
	require.NoError(t, err)
	require.Equal(t, 1, state.Version)
	require.Len(t, state.Scopes, 1)
	require.Equal(t, "generate-api", state.Scopes[0].ID)
	require.Len(t, state.Scopes[0].Artifacts, 1)
	require.Equal(t, "handlers", state.Scopes[0].Artifacts[0].ID)
	assert.Nil(t, state.Scopes[0].Artifacts[0].AdjustedAt)
	assert.Nil(t, state.Scopes[0].Artifacts[0].InstructionsPath)
	assert.Contains(t, state.Scopes[0].Artifacts[0].Content, "Z:")

	decoded, err := decodeAcceleratorContentSnapshot(state.Scopes[0].Artifacts[0].Content)
	require.NoError(t, err)
	assert.Equal(t, "Hello Carlos!", decoded)
}

func TestAccelerator_AccelerateStoresRelativeInstructionsPath(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, ctx.SetScope("generate-api"))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "instructions"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Hello {{ data.name }}!"), 0644))

	value, err := ctx.runtime.RunString(`({ name: "Carlos" })`)
	require.NoError(t, err)

	absoluteInstructionsPath := filepath.Join(ctx.ccodeContext.config.CCodePath, "instructions", "handlers.md")
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value, absoluteInstructionsPath))

	statePath := filepath.Join(ctx.ccodeContext.config.CCodePath, DefaultHiddenFolderName, "state", "accelerators.json")
	state, err := loadAcceleratorState(statePath)
	require.NoError(t, err)

	require.Len(t, state.Scopes, 1)
	require.Len(t, state.Scopes[0].Artifacts, 1)
	require.NotNil(t, state.Scopes[0].Artifacts[0].InstructionsPath)
	assert.Equal(t, "instructions/handlers.md", *state.Scopes[0].Artifacts[0].InstructionsPath)
}

func TestAccelerator_AccelerateSkipsOverwriteForModifiedFile(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, ctx.SetScope("generate-api"))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	templatePath := filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl")
	require.NoError(t, os.WriteFile(templatePath, []byte("First"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)

	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	outputFilePath := filepath.Join(ctx.ccodeContext.config.OutputPath, "handlers")
	require.NoError(t, os.WriteFile(outputFilePath, []byte("manually changed"), 0644))
	require.NoError(t, os.WriteFile(templatePath, []byte("Second"), 0644))

	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	content, readErr := os.ReadFile(outputFilePath)
	require.NoError(t, readErr)
	assert.Equal(t, "manually changed", string(content))

	statePath := filepath.Join(ctx.ccodeContext.config.CCodePath, DefaultHiddenFolderName, "state", "accelerators.json")
	state, err := loadAcceleratorState(statePath)
	require.NoError(t, err)
	decoded, err := decodeAcceleratorContentSnapshot(state.Scopes[0].Artifacts[0].Content)
	require.NoError(t, err)
	assert.Equal(t, "First", decoded)
}

func TestAccelerator_AccelerateAllowsOverwriteWhenContentMatchesLastGeneratedVersion(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, ctx.SetScope("generate-api"))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	templatePath := filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl")
	require.NoError(t, os.WriteFile(templatePath, []byte("First"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	require.NoError(t, os.WriteFile(templatePath, []byte("Second"), 0644))
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	outputFilePath := filepath.Join(ctx.ccodeContext.config.OutputPath, "handlers")
	content, err := os.ReadFile(outputFilePath)
	require.NoError(t, err)
	assert.Equal(t, "Second", string(content))

	statePath := filepath.Join(ctx.ccodeContext.config.CCodePath, DefaultHiddenFolderName, "state", "accelerators.json")
	state, err := loadAcceleratorState(statePath)
	require.NoError(t, err)

	decoded, err := decodeAcceleratorContentSnapshot(state.Scopes[0].Artifacts[0].Content)
	require.NoError(t, err)
	assert.Equal(t, "Second", decoded)
}

func TestAccelerator_AccelerateSkipsOverwriteWhenRenderedContentMatchesPreviousSnapshot(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, ctx.SetScope("generate-api"))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	templatePath := filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl")
	require.NoError(t, os.WriteFile(templatePath, []byte("First"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	statePath := filepath.Join(ctx.ccodeContext.config.CCodePath, DefaultHiddenFolderName, "state", "accelerators.json")
	stateBefore, err := loadAcceleratorState(statePath)
	require.NoError(t, err)
	contentBefore := stateBefore.Scopes[0].Artifacts[0].Content

	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	stateAfter, err := loadAcceleratorState(statePath)
	require.NoError(t, err)
	assert.Equal(t, contentBefore, stateAfter.Scopes[0].Artifacts[0].Content)
}

func TestAccelerator_AccelerateSkipsWhenExistingFileIsNotTracked(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, ctx.SetScope("generate-api"))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Generated"), 0644))

	outputFilePath := filepath.Join(ctx.ccodeContext.config.OutputPath, "handlers")
	require.NoError(t, os.MkdirAll(filepath.Dir(outputFilePath), 0755))
	require.NoError(t, os.WriteFile(outputFilePath, []byte("manual existing content"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	content, err := os.ReadFile(outputFilePath)
	require.NoError(t, err)
	assert.Equal(t, "manual existing content", string(content))

	statePath := filepath.Join(ctx.ccodeContext.config.CCodePath, DefaultHiddenFolderName, "state", "accelerators.json")
	assert.NoFileExists(t, statePath)
}

func TestAccelerator_RunUsesDefaultScopeFromProcessFileAndAllowsSetScope(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestAccelerator_RunUsesDefaultScopeFromProcessFileAndAllowsSetScope")
	processFile := filepath.Join(projectDir, "x", "generate-api.ts")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "templates", "handlers.tpl"), []byte("handlers"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "templates", "service.tpl"), []byte("service"), 0644))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\n\nexport default function main(ctx: Context) {\n\tctx.println(ctx.scope());\n\tctx.accelerate(\"handlers\", \"templates/handlers.tpl\", {});\n\tctx.setScope(\"custom-scope\");\n\tctx.println(ctx.scope());\n\tctx.accelerate(\"service\", \"templates/service.tpl\", {});\n}\n"), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("x/generate-api"))

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.Len(t, lines, 2)
	assert.Equal(t, "generate-api", lines[0])
	assert.Equal(t, "custom-scope", lines[1])

	defaultScopeFile := filepath.Join(filepath.Dir(projectDir), "output", "handlers")
	customScopeFile := filepath.Join(filepath.Dir(projectDir), "output", "service")
	require.FileExists(t, defaultScopeFile)
	require.FileExists(t, customScopeFile)

	statePath := filepath.Join(projectDir, DefaultHiddenFolderName, "state", "accelerators.json")
	state, err := loadAcceleratorState(statePath)
	require.NoError(t, err)
	require.Len(t, state.Scopes, 2)
	assert.Equal(t, "generate-api", state.Scopes[0].ID)
	assert.Equal(t, "handlers", state.Scopes[0].Artifacts[0].ID)
	assert.Equal(t, "custom-scope", state.Scopes[1].ID)
	assert.Equal(t, "service", state.Scopes[1].Artifacts[0].ID)
}
