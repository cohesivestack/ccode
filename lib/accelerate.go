package ccode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
)

type acceleratorState struct {
	Version int                     `json:"version"`
	Scopes  []acceleratorStateScope `json:"scopes"`
}

type acceleratorStateScope struct {
	ID        string                     `json:"id"`
	Artifacts []acceleratorStateArtifact `json:"artifacts"`
}

type acceleratorStateArtifact struct {
	ID               string  `json:"id"`
	Content          string  `json:"content"`
	AdjustedAt       *string `json:"adjusted_at"`
	InstructionsPath *string `json:"instructions_path"`
}

type AcceleratorArtifactMetadata struct {
	ScopeID          string  `json:"scope_id"`
	ArtifactID       string  `json:"artifact_id"`
	InstructionsPath *string `json:"instructions_path"`
	AdjustedAt       *string `json:"adjusted_at"`
}

type AcceleratorArtifactState struct {
	ScopeID          string  `json:"scope_id"`
	ArtifactID       string  `json:"artifact_id"`
	Content          string  `json:"content"`
	InstructionsPath *string `json:"instructions_path"`
	AdjustedAt       *string `json:"adjusted_at"`
}

type AcceleratorInstructionReference struct {
	ScopeID          string  `json:"scope_id"`
	ArtifactID       string  `json:"artifact_id"`
	InstructionsPath *string `json:"instructions_path"`
}

func (ctx *RunnerContext) Accelerate(id string, templatePath string, data goja.Value, instructionsPath ...string) error {
	if ctx == nil || ctx.ccodeContext == nil {
		return fmt.Errorf("runner context is not initialized")
	}

	scopeName, err := normalizeAcceleratorPathPart(ctx.Scope(), "scope")
	if err != nil {
		return err
	}

	artifactID, err := normalizeAcceleratorPathPart(id, "artifact id")
	if err != nil {
		return err
	}

	templateData, err := gojaValueToTemplateData(data)
	if err != nil {
		return fmt.Errorf("convert template data: %w", err)
	}

	rendered, err := ctx.renderTemplate(templatePath, templateData)
	if err != nil {
		return err
	}

	stateFilePath := ctx.acceleratorStateFilePath()
	state, err := loadAcceleratorState(stateFilePath)
	if err != nil {
		return err
	}

	scope, artifact := state.findArtifact(scopeName, artifactID)
	outputFilePath := filepath.Clean(filepath.Join(ctx.ccodeContext.config.OutputPath, filepath.FromSlash(artifactID)))
	shouldWrite, err := ctx.shouldWriteAcceleratedArtifact(outputFilePath, scopeName, artifactID, rendered, artifact)
	if err != nil {
		return err
	}
	if !shouldWrite {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(outputFilePath), 0755); err != nil {
		return fmt.Errorf("create output directory for %q: %w", outputFilePath, err)
	}
	if err := os.WriteFile(outputFilePath, []byte(rendered), 0644); err != nil {
		return fmt.Errorf("write accelerated artifact to %q: %w", outputFilePath, err)
	}

	artifact.Content = encodeAcceleratorContentSnapshot(rendered, time.Now().UTC())
	if len(instructionsPath) > 0 {
		normalizedInstructionsPath, err := ctx.normalizeInstructionsPath(instructionsPath[0])
		if err != nil {
			return err
		}
		artifact.InstructionsPath = &normalizedInstructionsPath
	}
	artifact.AdjustedAt = nil

	if scope.Artifacts == nil {
		scope.Artifacts = []acceleratorStateArtifact{}
	}

	if err := saveAcceleratorState(stateFilePath, state); err != nil {
		return err
	}

	return nil
}

func (ctx *RunnerContext) shouldWriteAcceleratedArtifact(outputFilePath string, scopeName string, artifactID string, rendered string, artifact *acceleratorStateArtifact) (bool, error) {
	if !fileExists(outputFilePath) {
		return true, nil
	}

	currentContent, err := os.ReadFile(outputFilePath)
	if err != nil {
		return false, fmt.Errorf("read accelerated artifact %q: %w", outputFilePath, err)
	}

	if artifact == nil || isStringBlank(artifact.Content) {
		return false, nil
	}

	lastGeneratedContent, err := decodeAcceleratorContentSnapshot(artifact.Content)
	if err != nil {
		return false, fmt.Errorf("decode stored content for artifact %q in scope %q: %w", artifactID, scopeName, err)
	}

	if string(currentContent) != lastGeneratedContent {
		return false, nil
	}

	if rendered == lastGeneratedContent {
		return false, nil
	}

	return true, nil
}

func (ctx *RunnerContext) acceleratorStateFilePath() string {
	return ctx.ccodeContext.acceleratorStateFilePath()
}

func (ctx *RunnerContext) normalizeInstructionsPath(instructionsPath string) (string, error) {
	return ctx.ccodeContext.normalizeAcceleratorInstructionPath(instructionsPath)
}

func normalizeAcceleratorPathPart(value string, label string) (string, error) {
	if isStringBlank(value) {
		return "", fmt.Errorf("%s is required", label)
	}

	normalizedValue := filepath.Clean(filepath.FromSlash(strings.TrimSpace(value)))
	if normalizedValue == "." || normalizedValue == ".." || filepath.IsAbs(normalizedValue) || strings.HasPrefix(normalizedValue, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%s %q is invalid", label, value)
	}

	return filepath.ToSlash(normalizedValue), nil
}

func loadAcceleratorState(path string) (*acceleratorState, error) {
	if !fileExists(path) {
		return &acceleratorState{
			Version: 1,
			Scopes:  []acceleratorStateScope{},
		}, nil
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read accelerator state %q: %w", path, err)
	}

	state := &acceleratorState{}
	if err := json.Unmarshal(payload, state); err != nil {
		return nil, fmt.Errorf("parse accelerator state %q: %w", path, err)
	}

	if state.Version == 0 {
		state.Version = 1
	}
	if state.Version != 1 {
		return nil, fmt.Errorf("unsupported accelerator state version %d", state.Version)
	}
	if state.Scopes == nil {
		state.Scopes = []acceleratorStateScope{}
	}
	for i := range state.Scopes {
		if state.Scopes[i].Artifacts == nil {
			state.Scopes[i].Artifacts = []acceleratorStateArtifact{}
		}
	}

	return state, nil
}

func saveAcceleratorState(path string, state *acceleratorState) error {
	if state == nil {
		return fmt.Errorf("accelerator state is required")
	}

	state.Version = 1
	if state.Scopes == nil {
		state.Scopes = []acceleratorStateScope{}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create accelerator state directory for %q: %w", path, err)
	}

	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode accelerator state %q: %w", path, err)
	}
	payload = append(payload, '\n')

	if err := os.WriteFile(path, payload, 0644); err != nil {
		return fmt.Errorf("write accelerator state %q: %w", path, err)
	}

	return nil
}

func (state *acceleratorState) findArtifact(scopeName string, artifactID string) (*acceleratorStateScope, *acceleratorStateArtifact) {
	scope := state.findOrCreateScope(scopeName)
	for index := range scope.Artifacts {
		if scope.Artifacts[index].ID == artifactID {
			return scope, &scope.Artifacts[index]
		}
	}

	scope.Artifacts = append(scope.Artifacts, acceleratorStateArtifact{
		ID: artifactID,
	})

	return scope, &scope.Artifacts[len(scope.Artifacts)-1]
}

func (state *acceleratorState) findOrCreateScope(scopeName string) *acceleratorStateScope {
	for index := range state.Scopes {
		if state.Scopes[index].ID == scopeName {
			return &state.Scopes[index]
		}
	}

	state.Scopes = append(state.Scopes, acceleratorStateScope{
		ID:        scopeName,
		Artifacts: []acceleratorStateArtifact{},
	})

	return &state.Scopes[len(state.Scopes)-1]
}

func (state *acceleratorState) findScopeByID(scopeID string) *acceleratorStateScope {
	for index := range state.Scopes {
		if state.Scopes[index].ID == scopeID {
			return &state.Scopes[index]
		}
	}
	return nil
}

func (scope *acceleratorStateScope) findArtifactByID(artifactID string) *acceleratorStateArtifact {
	for index := range scope.Artifacts {
		if scope.Artifacts[index].ID == artifactID {
			return &scope.Artifacts[index]
		}
	}
	return nil
}

func encodeAcceleratorContentSnapshot(content string, generatedAt time.Time) string {
	encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
	return fmt.Sprintf("%s:%s", generatedAt.UTC().Format(time.RFC3339), encodedContent)
}

func decodeAcceleratorContentSnapshot(snapshot string) (string, error) {
	separatorIndex := strings.Index(snapshot, "Z:")
	if separatorIndex <= 0 {
		return "", fmt.Errorf("invalid content snapshot format")
	}

	encodedContent := snapshot[separatorIndex+2:]
	if encodedContent == "" {
		return "", fmt.Errorf("content snapshot payload is empty")
	}

	content, err := base64.StdEncoding.DecodeString(encodedContent)
	if err != nil {
		return "", fmt.Errorf("decode content snapshot payload: %w", err)
	}

	return string(content), nil
}

func DecodeAcceleratorContentSnapshot(snapshot string) (string, error) {
	return decodeAcceleratorContentSnapshot(snapshot)
}

func (ctx *Context) acceleratorStateFilePath() string {
	hiddenPath := ctx.config.HiddenPath
	if isStringBlank(hiddenPath) {
		hiddenPath = filepath.Join(ctx.config.CCodePath, DefaultHiddenFolderName)
	} else if !filepath.IsAbs(hiddenPath) {
		hiddenPath = filepath.Join(ctx.config.CCodePath, hiddenPath)
	}

	return filepath.Join(hiddenPath, "state", "accelerators.json")
}

func (ctx *Context) normalizeAcceleratorInstructionPath(instructionsPath string) (string, error) {
	if isStringBlank(instructionsPath) {
		return "", fmt.Errorf("instructions path is required when provided")
	}

	normalizedPath := filepath.Clean(filepath.FromSlash(strings.TrimSpace(instructionsPath)))
	if filepath.IsAbs(normalizedPath) {
		relativePath, err := filepath.Rel(ctx.config.CCodePath, normalizedPath)
		if err != nil {
			return "", fmt.Errorf("resolve instructions path %q: %w", instructionsPath, err)
		}
		normalizedPath = filepath.Clean(relativePath)
	}

	if normalizedPath == "." || normalizedPath == ".." || strings.HasPrefix(normalizedPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("instructions path %q must be inside ccode path", instructionsPath)
	}

	return filepath.ToSlash(normalizedPath), nil
}

func (ctx *Context) MarkAcceleratorAsAdjusted(scopeID *string, artifactID *string) error {
	if ctx == nil || ctx.config == nil {
		return fmt.Errorf("context config is required")
	}

	if scopeID == nil && artifactID != nil {
		return fmt.Errorf("scope id is required when artifact id is provided")
	}

	statePath := ctx.acceleratorStateFilePath()
	state, err := loadAcceleratorState(statePath)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)

	switch {
	case scopeID == nil && artifactID == nil:
		for scopeIndex := range state.Scopes {
			for artifactIndex := range state.Scopes[scopeIndex].Artifacts {
				state.Scopes[scopeIndex].Artifacts[artifactIndex].AdjustedAt = &now
			}
		}
	case scopeID != nil && artifactID == nil:
		normalizedScopeID, err := normalizeAcceleratorPathPart(*scopeID, "scope id")
		if err != nil {
			return err
		}

		scope := state.findScopeByID(normalizedScopeID)
		if scope == nil {
			return fmt.Errorf("accelerator scope %q not found", normalizedScopeID)
		}
		for artifactIndex := range scope.Artifacts {
			scope.Artifacts[artifactIndex].AdjustedAt = &now
		}
	case scopeID != nil && artifactID != nil:
		normalizedScopeID, err := normalizeAcceleratorPathPart(*scopeID, "scope id")
		if err != nil {
			return err
		}
		normalizedArtifactID, err := normalizeAcceleratorPathPart(*artifactID, "artifact id")
		if err != nil {
			return err
		}

		scope := state.findScopeByID(normalizedScopeID)
		if scope == nil {
			return fmt.Errorf("accelerator scope %q not found", normalizedScopeID)
		}
		artifact := scope.findArtifactByID(normalizedArtifactID)
		if artifact == nil {
			return fmt.Errorf("accelerator artifact %q not found in scope %q", normalizedArtifactID, normalizedScopeID)
		}
		artifact.AdjustedAt = &now
	}

	return saveAcceleratorState(statePath, state)
}

func (ctx *Context) ListNotAdjustedAccelerators(scopeID *string) ([]AcceleratorArtifactMetadata, error) {
	if ctx == nil || ctx.config == nil {
		return nil, fmt.Errorf("context config is required")
	}

	var filterScopeID *string
	if scopeID != nil {
		normalizedScopeID, err := normalizeAcceleratorPathPart(*scopeID, "scope id")
		if err != nil {
			return nil, err
		}
		filterScopeID = &normalizedScopeID
	}

	state, err := loadAcceleratorState(ctx.acceleratorStateFilePath())
	if err != nil {
		return nil, err
	}

	items := []AcceleratorArtifactMetadata{}
	for _, scope := range state.Scopes {
		if filterScopeID != nil && scope.ID != *filterScopeID {
			continue
		}
		for _, artifact := range scope.Artifacts {
			if artifact.AdjustedAt != nil {
				continue
			}
			items = append(items, AcceleratorArtifactMetadata{
				ScopeID:          scope.ID,
				ArtifactID:       artifact.ID,
				InstructionsPath: artifact.InstructionsPath,
				AdjustedAt:       artifact.AdjustedAt,
			})
		}
	}

	return items, nil
}

func (ctx *Context) GetAcceleratorState(scopeID string, artifactID string) (*AcceleratorArtifactState, error) {
	if ctx == nil || ctx.config == nil {
		return nil, fmt.Errorf("context config is required")
	}

	normalizedScopeID, err := normalizeAcceleratorPathPart(scopeID, "scope id")
	if err != nil {
		return nil, err
	}
	normalizedArtifactID, err := normalizeAcceleratorPathPart(artifactID, "artifact id")
	if err != nil {
		return nil, err
	}

	state, err := loadAcceleratorState(ctx.acceleratorStateFilePath())
	if err != nil {
		return nil, err
	}

	scope := state.findScopeByID(normalizedScopeID)
	if scope == nil {
		return nil, fmt.Errorf("accelerator scope %q not found", normalizedScopeID)
	}
	artifact := scope.findArtifactByID(normalizedArtifactID)
	if artifact == nil {
		return nil, fmt.Errorf("accelerator artifact %q not found in scope %q", normalizedArtifactID, normalizedScopeID)
	}

	return &AcceleratorArtifactState{
		ScopeID:          scope.ID,
		ArtifactID:       artifact.ID,
		Content:          artifact.Content,
		InstructionsPath: artifact.InstructionsPath,
		AdjustedAt:       artifact.AdjustedAt,
	}, nil
}

func (ctx *Context) ListAcceleratorInstructions() ([]AcceleratorInstructionReference, error) {
	if ctx == nil || ctx.config == nil {
		return nil, fmt.Errorf("context config is required")
	}

	state, err := loadAcceleratorState(ctx.acceleratorStateFilePath())
	if err != nil {
		return nil, err
	}

	items := []AcceleratorInstructionReference{}
	for _, scope := range state.Scopes {
		for _, artifact := range scope.Artifacts {
			if artifact.InstructionsPath == nil || isStringBlank(*artifact.InstructionsPath) {
				continue
			}
			items = append(items, AcceleratorInstructionReference{
				ScopeID:          scope.ID,
				ArtifactID:       artifact.ID,
				InstructionsPath: artifact.InstructionsPath,
			})
		}
	}

	return items, nil
}

func (ctx *Context) GetAcceleratorInstruction(markdownPath string) (string, error) {
	if ctx == nil || ctx.config == nil {
		return "", fmt.Errorf("context config is required")
	}

	normalizedPath, err := ctx.normalizeAcceleratorInstructionPath(markdownPath)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(ctx.config.CCodePath, filepath.FromSlash(normalizedPath))
	if !fileExists(fullPath) {
		return "", fmt.Errorf("instruction file not found: %s", normalizedPath)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("read instruction file %q: %w", normalizedPath, err)
	}

	return string(content), nil
}

func (ctx *RunnerContext) MarkAcceleratorAsAdjusted(scopeID *string, artifactID *string) error {
	if ctx == nil || ctx.ccodeContext == nil {
		return fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.MarkAcceleratorAsAdjusted(scopeID, artifactID)
}

func (ctx *RunnerContext) ListNotAdjustedAccelerators(scopeID *string) ([]AcceleratorArtifactMetadata, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.ListNotAdjustedAccelerators(scopeID)
}

func (ctx *RunnerContext) GetAcceleratorState(scopeID string, artifactID string) (*AcceleratorArtifactState, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.GetAcceleratorState(scopeID, artifactID)
}

func (ctx *RunnerContext) ListAcceleratorInstructions() ([]AcceleratorInstructionReference, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.ListAcceleratorInstructions()
}

func (ctx *RunnerContext) GetAcceleratorInstruction(markdownPath string) (string, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return "", fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.GetAcceleratorInstruction(markdownPath)
}
