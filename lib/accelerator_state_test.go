package ccode

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcceleratorState_MarkSingleArtifactAsAdjusted(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{ID: "handlers", Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC())},
					{ID: "service", Content: encodeAcceleratorContentSnapshot("service", time.Now().UTC())},
				},
			},
		},
	})

	scopeID := "generate-api"
	artifactID := "handlers"
	require.NoError(t, ctx.MarkAcceleratorAsAdjusted(&scopeID, &artifactID))

	state := readAcceleratorStateForTest(t, ctx)
	require.Len(t, state.Scopes, 1)
	require.Len(t, state.Scopes[0].Artifacts, 2)
	require.NotNil(t, state.Scopes[0].Artifacts[0].AdjustedAt)
	assertRFC3339(t, *state.Scopes[0].Artifacts[0].AdjustedAt)
	assert.Nil(t, state.Scopes[0].Artifacts[1].AdjustedAt)
}

func TestAcceleratorState_MarkAllArtifactsInScopeAsAdjusted(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{ID: "handlers", Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC())},
					{ID: "service", Content: encodeAcceleratorContentSnapshot("service", time.Now().UTC())},
				},
			},
			{
				ID: "other",
				Artifacts: []acceleratorStateArtifact{
					{ID: "ignored", Content: encodeAcceleratorContentSnapshot("ignored", time.Now().UTC())},
				},
			},
		},
	})

	scopeID := "generate-api"
	require.NoError(t, ctx.MarkAcceleratorAsAdjusted(&scopeID, nil))

	state := readAcceleratorStateForTest(t, ctx)
	require.NotNil(t, state.Scopes[0].Artifacts[0].AdjustedAt)
	require.NotNil(t, state.Scopes[0].Artifacts[1].AdjustedAt)
	assert.Nil(t, state.Scopes[1].Artifacts[0].AdjustedAt)
}

func TestAcceleratorState_MarkAllArtifactsGloballyAsAdjusted(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{ID: "handlers", Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC())},
				},
			},
			{
				ID: "generate-web",
				Artifacts: []acceleratorStateArtifact{
					{ID: "page.ts", Content: encodeAcceleratorContentSnapshot("page", time.Now().UTC())},
				},
			},
		},
	})

	require.NoError(t, ctx.MarkAcceleratorAsAdjusted(nil, nil))

	state := readAcceleratorStateForTest(t, ctx)
	require.NotNil(t, state.Scopes[0].Artifacts[0].AdjustedAt)
	require.NotNil(t, state.Scopes[1].Artifacts[0].AdjustedAt)
}

func TestAcceleratorState_ListNotAdjustedAccelerators(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	now := time.Now().UTC().Format(time.RFC3339)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{ID: "handlers", Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC()), AdjustedAt: nil},
					{ID: "service", Content: encodeAcceleratorContentSnapshot("service", time.Now().UTC()), AdjustedAt: &now},
				},
			},
		},
	})

	items, err := ctx.ListNotAdjustedAccelerators(nil)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "generate-api", items[0].ScopeID)
	assert.Equal(t, "handlers", items[0].ArtifactID)
	assert.Nil(t, items[0].AdjustedAt)
}

func TestAcceleratorState_ListNotAdjustedAcceleratorsByScope(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{ID: "handlers", Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC())},
				},
			},
			{
				ID: "generate-web",
				Artifacts: []acceleratorStateArtifact{
					{ID: "page.ts", Content: encodeAcceleratorContentSnapshot("page", time.Now().UTC())},
				},
			},
		},
	})

	scopeID := "generate-web"
	items, err := ctx.ListNotAdjustedAccelerators(&scopeID)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "generate-web", items[0].ScopeID)
	assert.Equal(t, "page.ts", items[0].ArtifactID)
}

func TestAcceleratorState_GetAcceleratorState(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{
						ID:               "handlers.go",
						Content:          encodeAcceleratorContentSnapshot("package handlers", time.Now().UTC()),
						InstructionsPath: strPtr("instructions/handlers.md"),
					},
				},
			},
		},
	})

	item, err := ctx.GetAcceleratorState("generate-api", "handlers.go")
	require.NoError(t, err)
	assert.Equal(t, "generate-api", item.ScopeID)
	assert.Equal(t, "handlers.go", item.ArtifactID)
	assert.NotEmpty(t, item.Content)
	require.NotNil(t, item.InstructionsPath)
	assert.Equal(t, "instructions/handlers.md", *item.InstructionsPath)
}

func TestAcceleratorState_ListAcceleratorInstructions(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{
						ID:               "handlers.go",
						Content:          encodeAcceleratorContentSnapshot("handlers", time.Now().UTC()),
						InstructionsPath: strPtr("instructions/handlers.md"),
					},
					{
						ID:      "service.go",
						Content: encodeAcceleratorContentSnapshot("service", time.Now().UTC()),
					},
				},
			},
		},
	})

	items, err := ctx.ListAcceleratorInstructions()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "generate-api", items[0].ScopeID)
	assert.Equal(t, "handlers.go", items[0].ArtifactID)
	require.NotNil(t, items[0].InstructionsPath)
	assert.Equal(t, "instructions/handlers.md", *items[0].InstructionsPath)
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

func TestAcceleratorState_MarkAdjustedReturnsErrorWhenScopeOrArtifactNotFound(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{ID: "handlers", Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC())},
				},
			},
		},
	})

	scopeID := "missing-scope"
	err := ctx.MarkAcceleratorAsAdjusted(&scopeID, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scope")

	scopeID = "generate-api"
	artifactID := "missing-artifact"
	err = ctx.MarkAcceleratorAsAdjusted(&scopeID, &artifactID)
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

func writeAcceleratorStateForTest(t *testing.T, ctx *Context, state *acceleratorState) {
	t.Helper()
	require.NoError(t, saveAcceleratorState(ctx.acceleratorStateFilePath(), state))
}

func readAcceleratorStateForTest(t *testing.T, ctx *Context) *acceleratorState {
	t.Helper()
	state, err := loadAcceleratorState(ctx.acceleratorStateFilePath())
	require.NoError(t, err)
	return state
}

func strPtr(value string) *string {
	return &value
}

func assertRFC3339(t *testing.T, value string) {
	t.Helper()
	_, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err)
}
