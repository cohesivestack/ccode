package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAcceleratorDiff(t *testing.T) {
	for _, test := range []struct {
		name     string
		original string
		current  string
		pending  bool
		want     string
	}{
		{"unchanged", "package handlers\n", "package handlers\n", true, ""},
		{"changed", "package original\n", "package edited\n", true, "@@ -1 +1 @@\n-package original\n+package edited\n"},
		{"resolved", "package original\n", "package edited\n", false, "@@ -1 +1 @@\n-package original\n+package edited\n"},
		{"empty snapshot", "", "package handlers\n", true, "@@ -0,0 +1 @@\n+package handlers\n"},
		{"empty output", "package handlers\n", "", true, "@@ -1 +0,0 @@\n-package handlers\n"},
		{"both empty", "", "", true, ""},
		{"unicode", "// café\n", "// té\n", true, "@@ -1 +1 @@\n-// café\n+// té\n"},
		{"missing newline", "package handlers\n", "package handlers", true, "@@ -1 +1 @@\n-package handlers\n+package handlers\n\\ No newline at end of file\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := newAcceleratorStateContext(t)
			artifactID := "nested/handlers.go"
			artifact := testAcceleratorArtifact(artifactID, test.original, test.pending)
			// A missing guide must not prevent comparing the available code.
			artifact.GuidePath = strPtr("guides/missing.md")
			writeAcceleratorStateForTest(t, ctx, artifact)
			writeAcceleratorOutputForTest(t, ctx, artifactID, test.current)
			statePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), artifactID)
			stateBefore, err := os.ReadFile(statePath)
			require.NoError(t, err)

			diff, err := ctx.GetAcceleratorDiff(artifactID)
			require.NoError(t, err)
			want := test.want
			if want != "" {
				want = "--- a/nested/handlers.go\n+++ b/nested/handlers.go\n" + want
			}
			assert.Equal(t, want, diff)
			stateAfter, err := os.ReadFile(statePath)
			require.NoError(t, err)
			assert.Equal(t, stateBefore, stateAfter)
			output, err := os.ReadFile(filepath.Join(ctx.config.OutputPath, filepath.FromSlash(artifactID)))
			require.NoError(t, err)
			assert.Equal(t, test.current, string(output))
		})
	}
}

func TestGetAcceleratorDiffErrors(t *testing.T) {
	for _, test := range []struct {
		name       string
		artifactID string
		state      string
		wantError  string
	}{
		{"unknown artifact", "missing.go", "", `accelerator artifact "missing.go" not found`},
		{"invalid artifact ID", "../handlers.go", "", "is invalid"},
		{"corrupt state", "handlers.go", "invalid json\n", "corrupt"},
		{"ambiguous state", "handlers.go", "{\"pending\":true}\n{\"pending\":false}\n", "ambiguous"},
		{"invalid snapshot", "handlers.go", "{\"code\":\"not base64!\"}\n", "decode content snapshot payload"},
		{"missing output", "handlers.go", "{\"code\":\"cGFja2FnZSBoYW5kbGVycw==\"}\n", `read accelerator artifact "handlers.go"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := newAcceleratorStateContext(t)
			if test.state != "" {
				statePath := acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "handlers.go")
				require.NoError(t, os.MkdirAll(filepath.Dir(statePath), 0755))
				require.NoError(t, os.WriteFile(statePath, []byte(test.state), 0644))
			}

			diff, err := ctx.GetAcceleratorDiff(test.artifactID)
			require.ErrorContains(t, err, test.wantError)
			assert.Empty(t, diff)
		})
	}
}
