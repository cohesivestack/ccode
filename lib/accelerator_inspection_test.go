package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcceleratorReadersRejectInvalidState(t *testing.T) {
	for _, test := range []struct {
		name    string
		payload string
	}{
		{"corrupt", "invalid json\n"},
		{"ambiguous", "{\"pending\":true}\n{\"pending\":false}\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := newAcceleratorStateContext(t)
			artifactID := "handlers.go"
			writeAcceleratorStateForTest(t, ctx, testAcceleratorArtifact(artifactID, "package handlers", true))
			statePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), artifactID)
			require.NoError(t, os.WriteFile(statePath, []byte(test.payload), 0644))

			content, err := ctx.GetAcceleratorContent(artifactID)
			require.ErrorContains(t, err, test.name)
			assert.Empty(t, content)
			guide, err := ctx.GetAcceleratorArtifactGuide(artifactID)
			require.ErrorContains(t, err, test.name)
			assert.Nil(t, guide)
			diff, err := ctx.GetAcceleratorDiff(artifactID)
			require.ErrorContains(t, err, test.name)
			assert.Empty(t, diff)
		})
	}
}

func TestAcceleratorInspectionPreservesDiagnostics(t *testing.T) {
	for _, test := range []struct {
		name         string
		payload      string
		guidePath    *string
		removeOutput bool
		wantState    string
		wantContent  string
	}{
		{"corrupt", "invalid json\n", nil, false, acceleratorReportStateCorrupt, ""},
		{"ambiguous", "{\"pending\":true}\n{\"pending\":false}\n", nil, false, acceleratorReportStateAmbiguous, ""},
		{"missing guide", "", strPtr("guides/missing.md"), false, acceleratorReportStateMissingGuide, "package handlers"},
		{"missing output", "", nil, true, acceleratorReportStateMissingArtifact, "package handlers"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := newAcceleratorStateContext(t)
			artifactID := "handlers.go"
			artifact := testAcceleratorArtifact(artifactID, "package handlers", true)
			artifact.GuidePath = test.guidePath
			writeAcceleratorStateForTest(t, ctx, artifact)
			if test.payload != "" {
				statePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), artifactID)
				require.NoError(t, os.WriteFile(statePath, []byte(test.payload), 0644))
			}
			if test.removeOutput {
				require.NoError(t, os.Remove(filepath.Join(ctx.config.OutputPath, artifactID)))
			}

			metadata, err := ctx.GetAcceleratorMetadata(artifactID)
			require.NoError(t, err)
			assert.Equal(t, artifactID, metadata.ArtifactID)
			assert.Equal(t, test.wantState, metadata.State)
			assert.NotEmpty(t, metadata.Message)
			bundle, err := ctx.GetAcceleratorBundle(artifactID)
			require.NoError(t, err)
			assert.Equal(t, *metadata, bundle.AcceleratorArtifactMetadata)
			assert.Equal(t, test.wantContent, bundle.AcceleratedContent)
			assert.Empty(t, bundle.GuideMarkdown)
		})
	}
}
