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

func TestAcceleratorState_MarkSingleArtifactAsAdjusted(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{ID: "handlers", Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC()), Pending: true},
					{ID: "service", Content: encodeAcceleratorContentSnapshot("service", time.Now().UTC()), Pending: true},
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
	assert.False(t, state.Scopes[0].Artifacts[0].Pending)
	assert.True(t, state.Scopes[0].Artifacts[1].Pending)
}

func TestAcceleratorState_MarkAllArtifactsInScopeAsAdjusted(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{ID: "handlers", Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC()), Pending: true},
					{ID: "service", Content: encodeAcceleratorContentSnapshot("service", time.Now().UTC()), Pending: true},
				},
			},
			{
				ID: "other",
				Artifacts: []acceleratorStateArtifact{
					{ID: "ignored", Content: encodeAcceleratorContentSnapshot("ignored", time.Now().UTC()), Pending: true},
				},
			},
		},
	})

	scopeID := "generate-api"
	require.NoError(t, ctx.MarkAcceleratorAsAdjusted(&scopeID, nil))

	state := readAcceleratorStateForTest(t, ctx)
	assert.False(t, state.Scopes[0].Artifacts[0].Pending)
	assert.False(t, state.Scopes[0].Artifacts[1].Pending)
	assert.True(t, state.Scopes[1].Artifacts[0].Pending)
}

func TestAcceleratorState_MarkAllArtifactsGloballyAsAdjusted(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{ID: "handlers", Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC()), Pending: true},
				},
			},
			{
				ID: "generate-web",
				Artifacts: []acceleratorStateArtifact{
					{ID: "page.ts", Content: encodeAcceleratorContentSnapshot("page", time.Now().UTC()), Pending: true},
				},
			},
		},
	})

	require.NoError(t, ctx.MarkAcceleratorAsAdjusted(nil, nil))

	state := readAcceleratorStateForTest(t, ctx)
	assert.False(t, state.Scopes[0].Artifacts[0].Pending)
	assert.False(t, state.Scopes[1].Artifacts[0].Pending)
}

func TestAcceleratorState_ListNotAdjustedAccelerators(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{ID: "handlers", Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC()), Pending: true},
					{ID: "service", Content: encodeAcceleratorContentSnapshot("service", time.Now().UTC()), Pending: false},
				},
			},
		},
	})

	items, err := ctx.ListNotAdjustedAccelerators(nil)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "generate-api", items[0].ScopeID)
	assert.Equal(t, "handlers", items[0].ArtifactID)
	assert.True(t, items[0].Pending)
}

func TestAcceleratorState_ListNotAdjustedAcceleratorsByScope(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{ID: "handlers", Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC()), Pending: true},
				},
			},
			{
				ID: "generate-web",
				Artifacts: []acceleratorStateArtifact{
					{ID: "page.ts", Content: encodeAcceleratorContentSnapshot("page", time.Now().UTC()), Pending: true},
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
						Pending:          true,
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
	assert.True(t, item.Pending)
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

func TestAcceleratorState_ListReportsCorruptStateWithoutDeletingFile(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	stateFilePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "generate-api", "handlers.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(stateFilePath), 0755))
	require.NoError(t, os.WriteFile(stateFilePath, []byte("{not json}\n"), 0644))

	items, err := ctx.ListNotAdjustedAccelerators(nil)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "generate-api", items[0].ScopeID)
	assert.Equal(t, "handlers.go", items[0].ArtifactID)
	assert.Equal(t, acceleratorReportStateCorrupt, items[0].State)
	assert.Contains(t, items[0].Message, "Re-run the accelerator")

	payload, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	assert.Equal(t, "{not json}\n", string(payload))
}

func TestAcceleratorState_ListReportsAmbiguousStateWithoutDeletingFile(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	stateFilePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "generate-api", "handlers.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(stateFilePath), 0755))
	require.NoError(t, os.WriteFile(stateFilePath, []byte(
		`{"pending":true,"code":"`+encodeAcceleratorContentSnapshot("handlers", time.Now().UTC())+`"}`+"\n"+
			`{"pending":false,"code":"`+encodeAcceleratorContentSnapshot("handlers", time.Now().UTC())+`"}`+"\n",
	), 0644))

	items, err := ctx.ListNotAdjustedAccelerators(nil)
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
	stateFilePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "generate-api", "handlers.go")
	artifact := acceleratorStateArtifact{
		ID:      "handlers.go",
		Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC()),
		Pending: true,
	}
	require.NoError(t, saveAcceleratorArtifactStateFile(stateFilePath, artifact))
	payload, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(stateFilePath, append(payload, payload...), 0644))

	items, err := ctx.ListNotAdjustedAccelerators(nil)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, acceleratorReportStatePending, items[0].State)

	cleanedPayload, err := os.ReadFile(stateFilePath)
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(string(cleanedPayload), "\n"))
}

func TestAcceleratorState_ListReportsMissingInstructions(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	stateFilePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "generate-api", "handlers.go")
	artifact := acceleratorStateArtifact{
		ID:                   "handlers.go",
		Content:              encodeAcceleratorContentSnapshot("handlers", time.Now().UTC()),
		InstructionsPath:     strPtr("instructions/missing.md"),
		Pending:              false,
		InstructionsChecksum: "sha256:old",
	}
	require.NoError(t, saveAcceleratorArtifactStateFile(stateFilePath, artifact))

	items, err := ctx.ListNotAdjustedAccelerators(nil)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.False(t, items[0].Pending)
	assert.Equal(t, acceleratorReportStateMissingInstructions, items[0].State)
	assert.Contains(t, items[0].Message, "Instruction file is missing")

	references, err := ctx.ListAcceleratorInstructions()
	require.NoError(t, err)
	require.Len(t, references, 1)
	assert.Equal(t, acceleratorReportStateMissingInstructions, references[0].State)
}

func TestAcceleratorState_InspectionRefreshesChangedInstructionChecksum(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	require.NoError(t, os.MkdirAll(filepath.Join(ctx.config.CCodePath, "instructions"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ctx.config.CCodePath, "instructions", "handlers.md"), []byte("# New instructions"), 0644))

	stateFilePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "generate-api", "handlers.go")
	artifact := acceleratorStateArtifact{
		ID:                   "handlers.go",
		Content:              encodeAcceleratorContentSnapshot("handlers", time.Now().UTC()),
		InstructionsPath:     strPtr("instructions/handlers.md"),
		Pending:              false,
		InstructionsChecksum: checksumAcceleratorBytes([]byte("# Old instructions")),
	}
	require.NoError(t, saveAcceleratorArtifactStateFile(stateFilePath, artifact))

	item, err := ctx.GetAcceleratorState("generate-api", "handlers.go")
	require.NoError(t, err)
	assert.True(t, item.Pending)
	assert.Equal(t, acceleratorReportStatePending, item.State)

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

func TestAcceleratorState_MarkAdjustedReturnsErrorWhenScopeOrArtifactNotFound(t *testing.T) {
	ctx := newAcceleratorStateContext(t)
	writeAcceleratorStateForTest(t, ctx, &acceleratorState{
		Version: 1,
		Scopes: []acceleratorStateScope{
			{
				ID: "generate-api",
				Artifacts: []acceleratorStateArtifact{
					{ID: "handlers", Content: encodeAcceleratorContentSnapshot("handlers", time.Now().UTC()), Pending: true},
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
