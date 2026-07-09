package ccode

import (
	"bytes"
	"encoding/json"
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

	statePath := ctx.acceleratorStateFilePath()
	require.DirExists(t, statePath)
	stateFilePath := filepath.Join(statePath, "generate-api", "handlers.accelerated.json")
	require.FileExists(t, stateFilePath)
	stateFileContent, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(string(stateFileContent), "\n"))
	var stateFile map[string]any
	require.NoError(t, json.Unmarshal(stateFileContent, &stateFile))
	assert.Equal(t, true, stateFile["pending"])
	assert.Contains(t, stateFile, "instructions")
	assert.Contains(t, stateFile, "accelerated_checksum")
	assert.Contains(t, stateFile, "instructions_checksum")
	assert.Contains(t, stateFile, "code")

	state, err := loadAcceleratorState(statePath)
	require.NoError(t, err)
	require.Equal(t, 1, state.Version)
	require.Len(t, state.Scopes, 1)
	require.Equal(t, "generate-api", state.Scopes[0].ID)
	require.Len(t, state.Scopes[0].Artifacts, 1)
	require.Equal(t, "handlers", state.Scopes[0].Artifacts[0].ID)
	assert.True(t, state.Scopes[0].Artifacts[0].Pending)
	assert.Nil(t, state.Scopes[0].Artifacts[0].InstructionsPath)
	assert.NotContains(t, state.Scopes[0].Artifacts[0].Content, "Z:")
	assert.NotEmpty(t, state.Scopes[0].Artifacts[0].AcceleratedChecksum)

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

	statePath := ctx.acceleratorStateFilePath()
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

	statePath := ctx.acceleratorStateFilePath()
	state, err := loadAcceleratorState(statePath)
	require.NoError(t, err)
	decoded, err := decodeAcceleratorContentSnapshot(state.Scopes[0].Artifacts[0].Content)
	require.NoError(t, err)
	assert.Equal(t, "Second", decoded)
	assert.True(t, state.Scopes[0].Artifacts[0].Pending)
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

	statePath := ctx.acceleratorStateFilePath()
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

	statePath := ctx.acceleratorStateFilePath()
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

	statePath := ctx.acceleratorStateFilePath()
	require.DirExists(t, statePath)
	state, err := loadAcceleratorState(statePath)
	require.NoError(t, err)
	require.Len(t, state.Scopes, 1)
	require.Len(t, state.Scopes[0].Artifacts, 1)
	assert.True(t, state.Scopes[0].Artifacts[0].Pending)
}

func TestAccelerator_AcceleratePreservesStateWhenGeneratedMetadataMatches(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, ctx.SetScope("generate-api"))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Generated"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	scopeID := "generate-api"
	artifactID := "handlers"
	require.NoError(t, ctx.MarkAcceleratorAsAdjusted(&scopeID, &artifactID))

	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	state := readAcceleratorStateForTest(t, ctx.ccodeContext)
	require.Len(t, state.Scopes, 1)
	require.Len(t, state.Scopes[0].Artifacts, 1)
	assert.False(t, state.Scopes[0].Artifacts[0].Pending)
}

func TestAccelerator_AccelerateResetsStateWhenInstructionsChecksumChanges(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, ctx.SetScope("generate-api"))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "instructions"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Generated"), 0644))
	instructionsFilePath := filepath.Join(ctx.ccodeContext.config.CCodePath, "instructions", "handlers.md")
	require.NoError(t, os.WriteFile(instructionsFilePath, []byte("First instructions"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value, "instructions/handlers.md"))

	scopeID := "generate-api"
	artifactID := "handlers"
	require.NoError(t, ctx.MarkAcceleratorAsAdjusted(&scopeID, &artifactID))

	require.NoError(t, os.WriteFile(instructionsFilePath, []byte("Second instructions"), 0644))
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value, "instructions/handlers.md"))

	state := readAcceleratorStateForTest(t, ctx.ccodeContext)
	require.Len(t, state.Scopes, 1)
	require.Len(t, state.Scopes[0].Artifacts, 1)
	assert.True(t, state.Scopes[0].Artifacts[0].Pending)
}

func TestAccelerator_AccelerateCleansDuplicatedIdenticalStateLines(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, ctx.SetScope("generate-api"))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Generated"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	scopeID := "generate-api"
	artifactID := "handlers"
	require.NoError(t, ctx.MarkAcceleratorAsAdjusted(&scopeID, &artifactID))

	stateFilePath := filepath.Join(ctx.acceleratorStateFilePath(), "generate-api", "handlers.accelerated.json")
	stateFileContent, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	trimmedStateLine := strings.TrimSpace(string(stateFileContent))
	require.NoError(t, os.WriteFile(stateFilePath, []byte(trimmedStateLine+"\n"+trimmedStateLine+"\n"), 0644))

	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	cleanedContent, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	assert.Equal(t, trimmedStateLine+"\n", string(cleanedContent))

	state := readAcceleratorStateForTest(t, ctx.ccodeContext)
	require.Len(t, state.Scopes, 1)
	require.Len(t, state.Scopes[0].Artifacts, 1)
	assert.False(t, state.Scopes[0].Artifacts[0].Pending)
}

func TestAccelerator_AccelerateResetsStateWhenMultipleStateLinesDiffer(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, ctx.SetScope("generate-api"))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Generated"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	stateFilePath := filepath.Join(ctx.acceleratorStateFilePath(), "generate-api", "handlers.accelerated.json")
	stateFileContent, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	trimmedStateLine := strings.TrimSpace(string(stateFileContent))
	require.NoError(t, os.WriteFile(stateFilePath, []byte(trimmedStateLine+"\n"+`{"pending":false,"instructions":"","accelerated_checksum":"sha256:different","instructions_checksum":"","code":"RGlmZmVyZW50"}`+"\n"), 0644))

	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	stateFileContent, err = os.ReadFile(stateFilePath)
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(string(stateFileContent), "\n"))

	state := readAcceleratorStateForTest(t, ctx.ccodeContext)
	require.Len(t, state.Scopes, 1)
	require.Len(t, state.Scopes[0].Artifacts, 1)
	assert.True(t, state.Scopes[0].Artifacts[0].Pending)
}

func TestAccelerator_AccelerateResetsStateWhenStateFileIsCorrupt(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, ctx.SetScope("generate-api"))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Generated"), 0644))

	stateFilePath := filepath.Join(ctx.acceleratorStateFilePath(), "generate-api", "handlers.accelerated.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(stateFilePath), 0755))
	require.NoError(t, os.WriteFile(stateFilePath, []byte("{not-json}\n"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	state := readAcceleratorStateForTest(t, ctx.ccodeContext)
	require.Len(t, state.Scopes, 1)
	require.Len(t, state.Scopes[0].Artifacts, 1)
	assert.True(t, state.Scopes[0].Artifacts[0].Pending)
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

	statePath := ctx.acceleratorStateFilePath()
	state, err := loadAcceleratorState(statePath)
	require.NoError(t, err)
	require.Len(t, state.Scopes, 2)
	defaultScope := state.findScopeByID("generate-api")
	require.NotNil(t, defaultScope)
	require.Len(t, defaultScope.Artifacts, 1)
	assert.Equal(t, "handlers", defaultScope.Artifacts[0].ID)
	customScope := state.findScopeByID("custom-scope")
	require.NotNil(t, customScope)
	require.Len(t, customScope.Artifacts, 1)
	assert.Equal(t, "service", customScope.Artifacts[0].ID)
}
