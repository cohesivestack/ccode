package ccode

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcceleratorState_MarkSingleArtifactAsResolved(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx,
		testAcceleratorArtifact("handlers", "handlers", true),
		testAcceleratorArtifact("service", "service", true),
	)

	artifactID := "handlers"
	require.NoError(t, ctx.MarkAcceleratorAsResolved(&artifactID))

	artifacts := readAcceleratorArtifactsForTest(t, ctx)
	require.Len(t, artifacts, 2)
	assert.False(t, artifacts[0].Pending)
	assert.True(t, artifacts[1].Pending)
}

func TestAcceleratorState_MarkAllArtifactsAsResolved(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx,
		testAcceleratorArtifact("handlers", "handlers", true),
		testAcceleratorArtifact("nested/page.ts", "page", true),
	)

	require.NoError(t, ctx.MarkAcceleratorAsResolved(nil))

	artifacts := readAcceleratorArtifactsForTest(t, ctx)
	require.Len(t, artifacts, 2)
	assert.False(t, artifacts[0].Pending)
	assert.False(t, artifacts[1].Pending)
}

func TestAcceleratorState_ListNotResolvedAccelerators(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx,
		testAcceleratorArtifact("handlers", "handlers", true),
		testAcceleratorArtifact("service", "service", false),
	)

	items, err := ctx.ListNotResolvedAccelerators()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "handlers", items[0].ArtifactID)
	assert.True(t, items[0].Pending)
}

func TestAcceleratorState_ListAcceleratorsIncludesResolvedWhenRequested(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx,
		testAcceleratorArtifact("handlers", "handlers", true),
		testAcceleratorArtifact("service", "service", false),
	)

	items, err := ctx.ListAccelerators(true)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "handlers", items[0].ArtifactID)
	assert.Equal(t, acceleratorReportStatePending, items[0].State)
	assert.Equal(t, "service", items[1].ArtifactID)
	assert.Equal(t, acceleratorReportStateResolved, items[1].State)
}

func TestAcceleratorState_GetAcceleratorState(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	artifact := testAcceleratorArtifact("handlers.go", "package handlers", true)
	artifact.InstructionsPath = strPtr("instructions/handlers.md")
	writeAcceleratorStateForTest(t, ctx, artifact)
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.config.CCodePath, "instructions"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.config.CCodePath, "instructions", "handlers.md"), nil, 0644))

	item, err := ctx.GetAcceleratorState("handlers.go")
	require.NoError(t, err)
	assert.Equal(t, "handlers.go", item.ArtifactID)
	assert.NotEmpty(t, item.Content)
	require.NotNil(t, item.InstructionsPath)
	assert.Equal(t, "instructions/handlers.md", *item.InstructionsPath)
	assert.True(t, item.Pending)
}

func TestAcceleratorState_ListAcceleratorInstructionsFiltersResolved(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	instructionsPath := filepath.Join(ctx.config.CCodePath, "instructions")
	require.NoError(t, os.MkdirAll(instructionsPath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(instructionsPath, "handlers.md"), []byte("# Handlers"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(instructionsPath, "service.md"), []byte("# Service"), 0644))

	handlers := testAcceleratorArtifact("handlers.go", "handlers", true)
	handlers.InstructionsPath = strPtr("instructions/handlers.md")
	handlers.InstructionsChecksum = checksumAcceleratorBytes([]byte("# Handlers"))
	service := testAcceleratorArtifact("service.go", "service", false)
	service.InstructionsPath = strPtr("instructions/service.md")
	service.InstructionsChecksum = checksumAcceleratorBytes([]byte("# Service"))
	writeAcceleratorStateForTest(t, ctx, handlers, service)

	items, err := ctx.ListAcceleratorInstructions()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "handlers.go", items[0].ArtifactID)

	items, err = ctx.ListAcceleratorInstructionsWithResolved(true)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "service.go", items[1].ArtifactID)
	assert.Equal(t, acceleratorReportStateResolved, items[1].State)
}

func TestAcceleratorState_ListReportsCorruptStateWithoutDeletingFile(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	stateFilePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "handlers.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(stateFilePath), 0755))
	require.NoError(t, os.WriteFile(stateFilePath, []byte("{not json}\n"), 0644))

	items, err := ctx.ListNotResolvedAccelerators()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "handlers.go", items[0].ArtifactID)
	assert.Equal(t, acceleratorReportStateCorrupt, items[0].State)
	assert.Contains(t, items[0].Message, "Re-run the accelerator")
	require.FileExists(t, stateFilePath)
}

func TestAcceleratorState_ListReportsAmbiguousStateWithoutDeletingFile(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	stateFilePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "handlers.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(stateFilePath), 0755))
	require.NoError(t, os.WriteFile(stateFilePath, []byte(
		`{"pending":true,"code":"`+encodeAcceleratorContentSnapshot("handlers", time.Now().UTC())+`"}`+"\n"+
			`{"pending":false,"code":"`+encodeAcceleratorContentSnapshot("handlers", time.Now().UTC())+`"}`+"\n",
	), 0644))

	items, err := ctx.ListNotResolvedAccelerators()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, acceleratorReportStateAmbiguous, items[0].State)
	assert.Contains(t, items[0].Message, "Re-run the accelerator")

	payload, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	assert.Contains(t, string(payload), `"pending":true`)
	assert.Contains(t, string(payload), `"pending":false`)
}

func TestAcceleratorState_ListDeduplicatesRepeatedStateLines(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	artifact := testAcceleratorArtifact("handlers.go", "handlers", true)
	writeAcceleratorStateForTest(t, ctx, artifact)
	stateFilePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), artifact.ID)
	payload, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(stateFilePath, append(payload, payload...), 0644))

	items, err := ctx.ListNotResolvedAccelerators()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, acceleratorReportStatePending, items[0].State)

	cleanedPayload, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(string(cleanedPayload), "\n"))
}

func TestAcceleratorState_ListReportsMissingInstructions(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	artifact := testAcceleratorArtifact("handlers.go", "handlers", false)
	artifact.InstructionsPath = strPtr("instructions/missing.md")
	artifact.InstructionsChecksum = "sha256:old"
	writeAcceleratorStateForTest(t, ctx, artifact)

	items, err := ctx.ListNotResolvedAccelerators()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.False(t, items[0].Pending)
	assert.Equal(t, acceleratorReportStateMissingInstructions, items[0].State)

	references, err := ctx.ListAcceleratorInstructions()
	require.NoError(t, err)
	require.Len(t, references, 1)
	assert.Equal(t, acceleratorReportStateMissingInstructions, references[0].State)
}

func TestAcceleratorState_ListReportsMissingArtifactBeforeMissingInstructions(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	artifact := testAcceleratorArtifact("handlers.go", "handlers", false)
	artifact.InstructionsPath = strPtr("instructions/missing.md")
	artifact.InstructionsChecksum = "sha256:old"
	stateFilePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), artifact.ID)
	require.NoError(t, saveAcceleratorArtifactStateFile(stateFilePath, artifact))

	items, err := ctx.ListNotResolvedAccelerators()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, acceleratorReportStateMissingArtifact, items[0].State)
}

func TestAcceleratorState_InspectionRefreshesChangedInstructionChecksum(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.config.CCodePath, "instructions"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.config.CCodePath, "instructions", "handlers.md"), []byte("# New instructions"), 0644))
	artifact := testAcceleratorArtifact("handlers.go", "handlers", false)
	artifact.InstructionsPath = strPtr("instructions/handlers.md")
	artifact.InstructionsChecksum = checksumAcceleratorBytes([]byte("# Old instructions"))
	writeAcceleratorStateForTest(t, ctx, artifact)

	item, err := ctx.GetAcceleratorState("handlers.go")
	require.NoError(t, err)
	assert.True(t, item.Pending)

	stateFilePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), artifact.ID)
	storedArtifact, err := loadAcceleratorArtifactStateFile(stateFilePath)
	require.NoError(t, err)
	assert.True(t, storedArtifact.Pending)
	assert.Equal(t, checksumAcceleratorBytes([]byte("# New instructions")), storedArtifact.InstructionsChecksum)
}

func TestAcceleratorState_GetAcceleratorInstruction(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.config.CCodePath, "instructions"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.config.CCodePath, "instructions", "handlers.md"), []byte("# Instructions"), 0644))

	content, err := ctx.GetAcceleratorInstruction("instructions/handlers.md")
	require.NoError(t, err)
	assert.Equal(t, "# Instructions", content)
}

func TestAcceleratorState_GetAcceleratorInstructionReturnsErrorWhenMissing(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	_, err := ctx.GetAcceleratorInstruction("instructions/missing.md")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "instruction file not found")
}

func TestAcceleratorState_MarkResolvedReturnsErrorWhenArtifactNotFound(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	artifactID := "missing-artifact"
	err := ctx.MarkAcceleratorAsResolved(&artifactID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "artifact")
}

func newAcceleratorStateContext(t *testing.T) *Context {
	t.Helper()
	rootDir := t.TempDir()
	ccodePath := filepath.Join(rootDir, "ccode")
	require.NoError(t, os.MkdirAll(ccodePath, 0755))
	return NewContext(&Config{
		CCodePath:  ccodePath,
		OutputPath: filepath.Join(rootDir, "output"),
	})
}

func testAcceleratorArtifact(id string, content string, pending bool) acceleratorStateArtifact {
	return acceleratorStateArtifact{
		ID:      id,
		Content: encodeAcceleratorContentSnapshot(content, time.Now().UTC()),
		Pending: pending,
	}
}

func writeAcceleratorStateForTest(t *testing.T, ctx *Context, artifacts ...acceleratorStateArtifact) {
	t.Helper()
	for _, artifact := range artifacts {
		stateFilePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), artifact.ID)
		require.NoError(t, saveAcceleratorArtifactStateFile(stateFilePath, artifact))
		content, err := decodeAcceleratorContentSnapshot(artifact.Content)
		if err == nil {
			writeAcceleratorOutputForTest(t, ctx, artifact.ID, content)
		}
	}
}

func readAcceleratorArtifactsForTest(t *testing.T, ctx *Context) []acceleratorStateArtifact {
	t.Helper()
	items, err := ctx.inspectAcceleratorState()
	require.NoError(t, err)
	artifacts := []acceleratorStateArtifact{}
	for _, item := range items {
		if item.Artifact == nil {
			continue
		}
		artifact := *item.Artifact
		artifact.ID = item.ArtifactID
		artifacts = append(artifacts, artifact)
	}
	return artifacts
}

func writeAcceleratorOutputForTest(t *testing.T, ctx *Context, artifactID string, content string) {
	t.Helper()
	outputFilePath := filepath.Join(ctx.config.OutputPath, filepath.FromSlash(artifactID))
	require.NoError(t, os.MkdirAll(filepath.Dir(outputFilePath), 0755))
	require.NoError(t, os.WriteFile(outputFilePath, []byte(content), 0644))
}

func strPtr(value string) *string {
	return &value
}
