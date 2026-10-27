package ccode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrune_RemovesMatchingFilesAndStateButPreservesCurrentArtifacts(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)
	root := ctx.ccodeContext.config.OutputPath
	ctx.ccodeContext.config.HiddenPath = filepath.Join(t.TempDir(), ".ccode")
	writePruneTestFile(t, filepath.Join(ctx.ccodeContext.config.CCodePath, "artifact.tpl"), "generated")
	data := ctx.runtime.ToValue(map[string]any{})

	require.NoError(t, ctx.Accelerate("nested/stale.acc.go", "artifact.tpl", data))
	writePruneTestFile(t, filepath.Join(root, "nested/stale.acc.go"), "edited obsolete artifact")
	ctx.ccodeContext.runArtifacts = nil // Begin a new run.

	require.NoError(t, ctx.Generate("artifact.tpl", "nested/./current.gen.go", data))
	require.NoError(t, ctx.Accelerate("nested/current.acc.go", "artifact.tpl", data))
	writePruneTestFile(t, filepath.Join(root, "nested/current.acc.go"), "edited current artifact")
	require.NoError(t, ctx.Accelerate("nested/current.acc.go", "artifact.tpl", data))
	writePruneTestFile(t, filepath.Join(root, "nested/untracked.acc.go"), "existing untracked artifact")
	require.NoError(t, ctx.Accelerate("nested/untracked.acc.go", "artifact.tpl", data))
	writePruneTestFile(t, filepath.Join(root, "nested/stale.gen.go"), "stale generated file")
	writePruneTestFile(t, filepath.Join(root, "nested/manual.go"), "manual file")
	writePruneTestFile(t, filepath.Join(root, "elsewhere/stale.gen.go"), "out of pattern scope")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "nested/empty.gen.go"), 0755))

	patterns := []string{`^nested/.*\.gen\.go$`, `^nested/.*\.acc\.go$`}
	require.NoError(t, ctx.Prune(patterns))
	require.NoError(t, ctx.Prune(patterns)) // Repeated pruning is safe.
	assert.NoFileExists(t, filepath.Join(root, "nested/stale.gen.go"))
	assert.NoFileExists(t, filepath.Join(root, "nested/stale.acc.go"))
	assert.NoFileExists(t, acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "nested/stale.acc.go"))
	for _, name := range []string{"nested/current.gen.go", "nested/manual.go", "elsewhere/stale.gen.go"} {
		assert.FileExists(t, filepath.Join(root, name))
	}
	for name, expected := range map[string]string{
		"nested/current.acc.go":   "edited current artifact",
		"nested/untracked.acc.go": "existing untracked artifact",
	} {
		content, err := os.ReadFile(filepath.Join(root, name))
		require.NoError(t, err)
		assert.Equal(t, expected, string(content))
		assert.FileExists(t, acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), name))
	}
	assert.DirExists(t, filepath.Join(root, "nested/empty.gen.go"))
}

func TestPrune_ValidatesAllPatternsBeforeDeleting(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)
	path := filepath.Join(ctx.ccodeContext.config.OutputPath, "stale.gen.go")
	writePruneTestFile(t, path, "stale")

	err := ctx.Prune([]string{`\.gen\.go$`, `[`})
	require.ErrorContains(t, err, "invalid prune path pattern")
	assert.FileExists(t, path)
	require.NoError(t, ctx.Prune(nil))
	require.NoError(t, ctx.Prune([]string{}))
	assert.FileExists(t, path)
	assert.ErrorContains(t, (*RunnerContext)(nil).Prune(nil), "runner context is not initialized")
	assert.ErrorContains(t, (&RunnerContext{}).Prune(nil), "runner context is not initialized")
}

func TestPrune_MissingOutputPathIsANoOp(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)
	require.NoError(t, ctx.Prune([]string{`.*`}))
	assert.NoDirExists(t, ctx.ccodeContext.config.OutputPath)
}

func TestPrune_ProtectsActiveStateInsideOutputPath(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)
	root := ctx.ccodeContext.config.OutputPath
	ctx.ccodeContext.config.HiddenPath = filepath.Join(root, ".ccode")
	writePruneTestFile(t, filepath.Join(ctx.ccodeContext.config.CCodePath, "artifact.tpl"), "generated")
	data := ctx.runtime.ToValue(map[string]any{})
	require.NoError(t, ctx.Accelerate("active.acc.go", "artifact.tpl", data))

	require.NoError(t, ctx.Prune([]string{`.*`}))
	assert.FileExists(t, filepath.Join(root, "active.acc.go"))
	assert.FileExists(t, acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "active.acc.go"))
}

func TestPrune_ResolvesOutputAliasesAndDoesNotFollowSymlinks(t *testing.T) {
	ctx := newRunnerTemplateTestContext(t)
	root := ctx.ccodeContext.config.OutputPath
	require.NoError(t, os.MkdirAll(root, 0755))
	alias := filepath.Join(t.TempDir(), "output-alias")
	require.NoError(t, os.Symlink(root, alias))
	ctx.ccodeContext.config.OutputPath = alias
	writePruneTestFile(t, filepath.Join(ctx.ccodeContext.config.CCodePath, "artifact.tpl"), "generated")
	// An absolute output path through another alias must still be protected.
	require.NoError(t, ctx.Generate("artifact.tpl", filepath.Join(root, "current.gen.go"), ctx.runtime.ToValue(nil)))
	writePruneTestFile(t, filepath.Join(root, "stale.gen.go"), "stale")
	outside := filepath.Join(t.TempDir(), "outside.gen.go")
	writePruneTestFile(t, outside, "outside output_path")
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "link.gen.go")))
	require.NoError(t, os.Symlink(filepath.Dir(outside), filepath.Join(root, "linked-directory")))
	require.NoError(t, os.Symlink(filepath.Join(root, "missing"), filepath.Join(root, "broken.gen.go")))

	require.NoError(t, ctx.Prune([]string{`\.gen\.go$`}))
	assert.NoFileExists(t, filepath.Join(root, "stale.gen.go"))
	assert.FileExists(t, filepath.Join(root, "current.gen.go"))
	assert.FileExists(t, outside)
	for _, name := range []string{"link.gen.go", "linked-directory", "broken.gen.go"} {
		info, err := os.Lstat(filepath.Join(root, name))
		require.NoError(t, err)
		assert.NotZero(t, info.Mode()&os.ModeSymlink)
	}
}

func TestPrune_RunSharesArtifactsAcrossNestedProcessesAndResetsBetweenRuns(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, t.Name())
	writePruneTestFile(t, filepath.Join(projectDir, "artifact.tpl"), "generated")
	writePruneTestFile(t, filepath.Join(projectDir, "parent.proc.ts"), `import type { Context } from "@ccode/context";
export default function main(ctx: Context) {
  ctx.generate("artifact.tpl", "parent.gen.go", {});
  ctx.runProcess("child");
  ctx.runProcess("child");
  ctx.prune(["\\.gen\\.go$", "\\.acc\\.go$"]);
}`)
	writePruneTestFile(t, filepath.Join(projectDir, "child.proc.ts"), `import type { Context } from "@ccode/context";
export default function main(ctx: Context) {
  ctx.generate("artifact.tpl", "child.gen.go", {});
  ctx.runProcess("grandchild");
  ctx.prune(["\\.gen\\.go$", "\\.acc\\.go$"]);
}`)
	writePruneTestFile(t, filepath.Join(projectDir, "grandchild.proc.ts"), `import type { Context } from "@ccode/context";
export default function main(ctx: Context) {
  ctx.accelerate("grandchild.acc.go", "artifact.tpl", {});
}`)
	writePruneTestFile(t, filepath.Join(projectDir, "prune.proc.ts"), `import type { Context } from "@ccode/context";
export default function main(ctx: Context) {
  ctx.prune(["\\.gen\\.go$", "\\.acc\\.go$"]);
}`)
	root := ctx.config.OutputPath
	writePruneTestFile(t, filepath.Join(root, "stale.gen.go"), "stale")
	writePruneTestFile(t, filepath.Join(root, "manual.go"), "manual")

	// The second run exercises accelerate's unchanged-output branch.
	for range 2 {
		require.NoError(t, ctx.Run("parent"))
		for _, name := range []string{"parent.gen.go", "child.gen.go", "grandchild.acc.go"} {
			assert.FileExists(t, filepath.Join(root, name))
		}
		assert.FileExists(t, acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "grandchild.acc.go"))
		assert.NoFileExists(t, filepath.Join(root, "stale.gen.go"))
	}

	require.NoError(t, ctx.Run("prune"))
	for _, name := range []string{"parent.gen.go", "child.gen.go", "grandchild.acc.go"} {
		assert.NoFileExists(t, filepath.Join(root, name))
	}
	assert.NoFileExists(t, acceleratorArtifactStateFilePath(ctx.acceleratorStateFilePath(), "grandchild.acc.go"))
	assert.FileExists(t, filepath.Join(root, "manual.go"))
}

func TestPrune_RunDiscardsTrackingAfterFailureAndReportsPatternErrors(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, t.Name())
	writePruneTestFile(t, filepath.Join(projectDir, "artifact.tpl"), "generated")
	writePruneTestFile(t, filepath.Join(projectDir, "failed.proc.ts"), `export default function main(ctx) {
  ctx.generate("artifact.tpl", "failed.gen.go", {});
  throw new Error("stop process");
}`)
	writePruneTestFile(t, filepath.Join(projectDir, "invalid.proc.ts"), `export default function main(ctx) {
  ctx.prune([".*", "["]);
}`)
	writePruneTestFile(t, filepath.Join(projectDir, "prune.proc.ts"), `export default function main(ctx) {
  ctx.prune(["\\.gen\\.go$"]);
}`)
	path := filepath.Join(ctx.config.OutputPath, "failed.gen.go")
	require.ErrorContains(t, ctx.Run("failed"), "stop process")
	assert.FileExists(t, path)
	require.ErrorContains(t, ctx.Run("invalid"), "invalid prune path pattern")
	assert.FileExists(t, path)
	require.NoError(t, ctx.Run("prune"))
	assert.NoFileExists(t, path)
}

func writePruneTestFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}
