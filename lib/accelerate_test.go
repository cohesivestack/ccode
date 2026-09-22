package ccode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccelerator_AccelerateWritesArtifactAndState(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

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
	stateFilePath := filepath.Join(statePath, "handlers.accelerated.json")
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

	artifacts := readAcceleratorArtifactsForTest(t, ctx.ccodeContext)
	require.Len(t, artifacts, 1)
	require.Equal(t, "handlers", artifacts[0].ID)
	assert.True(t, artifacts[0].Pending)
	assert.Nil(t, artifacts[0].InstructionsPath)
	assert.NotContains(t, artifacts[0].Content, "Z:")
	assert.NotEmpty(t, artifacts[0].AcceleratedChecksum)

	decoded, err := decodeAcceleratorContentSnapshot(artifacts[0].Content)
	require.NoError(t, err)
	assert.Equal(t, "Hello Carlos!", decoded)
}

func TestAccelerator_AccelerateStoresRelativeInstructionsPath(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "instructions"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Hello {{ data.name }}!"), 0644))

	value, err := ctx.runtime.RunString(`({ name: "Carlos" })`)
	require.NoError(t, err)

	absoluteInstructionsPath := filepath.Join(ctx.ccodeContext.config.CCodePath, "instructions", "handlers.md")
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value, absoluteInstructionsPath))

	artifacts := readAcceleratorArtifactsForTest(t, ctx.ccodeContext)
	require.Len(t, artifacts, 1)
	require.NotNil(t, artifacts[0].InstructionsPath)
	assert.Equal(t, "instructions/handlers.md", *artifacts[0].InstructionsPath)
}

func TestAccelerator_AccelerateSkipsOverwriteForModifiedFile(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

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

	artifacts := readAcceleratorArtifactsForTest(t, ctx.ccodeContext)
	decoded, err := decodeAcceleratorContentSnapshot(artifacts[0].Content)
	require.NoError(t, err)
	assert.Equal(t, "Second", decoded)
	assert.True(t, artifacts[0].Pending)
}

func TestAccelerator_AccelerateAllowsOverwriteWhenContentMatchesLastGeneratedVersion(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

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

	artifacts := readAcceleratorArtifactsForTest(t, ctx.ccodeContext)
	decoded, err := decodeAcceleratorContentSnapshot(artifacts[0].Content)
	require.NoError(t, err)
	assert.Equal(t, "Second", decoded)
}

func TestAccelerator_AccelerateSkipsOverwriteWhenRenderedContentMatchesPreviousSnapshot(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	templatePath := filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl")
	require.NoError(t, os.WriteFile(templatePath, []byte("First"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	artifactsBefore := readAcceleratorArtifactsForTest(t, ctx.ccodeContext)
	contentBefore := artifactsBefore[0].Content

	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	artifactsAfter := readAcceleratorArtifactsForTest(t, ctx.ccodeContext)
	assert.Equal(t, contentBefore, artifactsAfter[0].Content)
}

func TestAccelerator_AccelerateSkipsWhenExistingFileIsNotTracked(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

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

	require.DirExists(t, ctx.acceleratorStateFilePath())
	artifacts := readAcceleratorArtifactsForTest(t, ctx.ccodeContext)
	require.Len(t, artifacts, 1)
	assert.True(t, artifacts[0].Pending)
}

func TestAccelerator_AcceleratePreservesStateWhenGeneratedMetadataMatches(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Generated"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	artifactID := "handlers"
	require.NoError(t, ctx.MarkAcceleratorAsResolved(&artifactID))

	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	artifacts := readAcceleratorArtifactsForTest(t, ctx.ccodeContext)
	require.Len(t, artifacts, 1)
	assert.False(t, artifacts[0].Pending)
}

func TestAccelerator_AccelerateResetsStateWhenInstructionsChecksumChanges(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "instructions"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Generated"), 0644))
	instructionsFilePath := filepath.Join(ctx.ccodeContext.config.CCodePath, "instructions", "handlers.md")
	require.NoError(t, os.WriteFile(instructionsFilePath, []byte("First instructions"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value, "instructions/handlers.md"))

	artifactID := "handlers"
	require.NoError(t, ctx.MarkAcceleratorAsResolved(&artifactID))

	require.NoError(t, os.WriteFile(instructionsFilePath, []byte("Second instructions"), 0644))
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value, "instructions/handlers.md"))

	artifacts := readAcceleratorArtifactsForTest(t, ctx.ccodeContext)
	require.Len(t, artifacts, 1)
	assert.True(t, artifacts[0].Pending)
}

func TestAccelerator_AccelerateCleansDuplicatedIdenticalStateLines(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Generated"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	artifactID := "handlers"
	require.NoError(t, ctx.MarkAcceleratorAsResolved(&artifactID))

	stateFilePath := filepath.Join(ctx.acceleratorStateFilePath(), "handlers.accelerated.json")
	stateFileContent, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	trimmedStateLine := strings.TrimSpace(string(stateFileContent))
	require.NoError(t, os.WriteFile(stateFilePath, []byte(trimmedStateLine+"\n"+trimmedStateLine+"\n"), 0644))

	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	cleanedContent, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	assert.Equal(t, trimmedStateLine+"\n", string(cleanedContent))

	artifacts := readAcceleratorArtifactsForTest(t, ctx.ccodeContext)
	require.Len(t, artifacts, 1)
	assert.False(t, artifacts[0].Pending)
}

func TestAccelerator_AccelerateResetsStateWhenMultipleStateLinesDiffer(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Generated"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	stateFilePath := filepath.Join(ctx.acceleratorStateFilePath(), "handlers.accelerated.json")
	stateFileContent, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	trimmedStateLine := strings.TrimSpace(string(stateFileContent))
	require.NoError(t, os.WriteFile(stateFilePath, []byte(trimmedStateLine+"\n"+`{"pending":false,"instructions":"","accelerated_checksum":"sha256:different","instructions_checksum":"","code":"RGlmZmVyZW50"}`+"\n"), 0644))

	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	stateFileContent, err = os.ReadFile(stateFilePath)
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(string(stateFileContent), "\n"))

	artifacts := readAcceleratorArtifactsForTest(t, ctx.ccodeContext)
	require.Len(t, artifacts, 1)
	assert.True(t, artifacts[0].Pending)
}

func TestAccelerator_AccelerateResetsStateWhenStateFileIsCorrupt(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)

	require.NoError(t, os.MkdirAll(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.ccodeContext.config.CCodePath, "templates", "handlers.tpl"), []byte("Generated"), 0644))

	stateFilePath := filepath.Join(ctx.acceleratorStateFilePath(), "handlers.accelerated.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(stateFilePath), 0755))
	require.NoError(t, os.WriteFile(stateFilePath, []byte("{not-json}\n"), 0644))

	value, err := ctx.runtime.RunString(`({})`)
	require.NoError(t, err)
	require.NoError(t, ctx.Accelerate("handlers", "templates/handlers.tpl", value))

	artifacts := readAcceleratorArtifactsForTest(t, ctx.ccodeContext)
	require.Len(t, artifacts, 1)
	assert.True(t, artifacts[0].Pending)
}

func TestAccelerator_RunStoresStateByArtifactID(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestAccelerator_RunStoresStateByArtifactID")
	processFile := filepath.Join(projectDir, "x", "generate-api.process.ts")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "templates", "handlers.tpl"), []byte("handlers"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "templates", "service.tpl"), []byte("service"), 0644))
	require.NoError(t, os.WriteFile(processFile, []byte("import type { Context } from \"@ccode/context\";\n\nexport default function main(ctx: Context) {\n\tctx.accelerate(\"handlers\", \"templates/handlers.tpl\", {});\n\tctx.accelerate(\"generated/service\", \"templates/service.tpl\", {});\n}\n"), 0644))

	require.NoError(t, ctx.Run("x/generate-api"))

	require.FileExists(t, filepath.Join(filepath.Dir(projectDir), "output", "handlers"))
	require.FileExists(t, filepath.Join(filepath.Dir(projectDir), "output", "generated", "service"))
	require.FileExists(t, acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "handlers"))
	require.FileExists(t, acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "generated/service"))

	artifacts := readAcceleratorArtifactsForTest(t, ctx)
	require.Len(t, artifacts, 2)
	assert.Equal(t, "generated/service", artifacts[0].ID)
	assert.Equal(t, "handlers", artifacts[1].ID)
}

func TestAccelerator_RunPreservesStateForArtifactsNotEmittedByProcess(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestAccelerator_RunPreservesStateForArtifactsNotEmittedByProcess")
	processFile := filepath.Join(projectDir, "generate.process.ts")
	require.NoError(t, os.WriteFile(processFile, []byte("export default function main() {}\n"), 0644))

	stateFilePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "existing")
	require.NoError(t, saveAcceleratorArtifactStateFile(stateFilePath, acceleratorStateArtifact{
		ID:      "existing",
		Content: encodeAcceleratorContentSnapshot("existing", time.Time{}),
		Pending: true,
	}))

	require.NoError(t, ctx.Run("generate"))
	require.FileExists(t, stateFilePath)
}
