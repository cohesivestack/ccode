package ccode

import (
	"crypto/sha256"
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
	ID                   string  `json:"id"`
	Content              string  `json:"content"`
	InstructionsPath     *string `json:"instructions_path"`
	Pending              bool    `json:"pending"`
	AcceleratedChecksum  string  `json:"accelerated_checksum"`
	InstructionsChecksum string  `json:"instructions_checksum"`
}

const (
	acceleratorReportStatePending             = "pending"
	acceleratorReportStateAdjusted            = "adjusted"
	acceleratorReportStateCorrupt             = "corrupt"
	acceleratorReportStateAmbiguous           = "ambiguous"
	acceleratorReportStateMissingArtifact     = "missing_artifact"
	acceleratorReportStateMissingInstructions = "missing_instructions"
)

type AcceleratorArtifactMetadata struct {
	ScopeID          string  `json:"scope_id"`
	ArtifactID       string  `json:"artifact_id"`
	InstructionsPath *string `json:"instructions_path"`
	Pending          bool    `json:"pending"`
	State            string  `json:"state"`
	Message          string  `json:"message,omitempty"`
}

type AcceleratorArtifactState struct {
	ScopeID          string  `json:"scope_id"`
	ArtifactID       string  `json:"artifact_id"`
	Content          string  `json:"content"`
	InstructionsPath *string `json:"instructions_path"`
	Pending          bool    `json:"pending"`
	State            string  `json:"state"`
	Message          string  `json:"message,omitempty"`
}

type AcceleratorInstructionReference struct {
	ScopeID          string  `json:"scope_id"`
	ArtifactID       string  `json:"artifact_id"`
	InstructionsPath *string `json:"instructions_path"`
	State            string  `json:"state"`
	Message          string  `json:"message,omitempty"`
}

type inspectedAcceleratorArtifact struct {
	ScopeID    string
	ArtifactID string
	Artifact   *acceleratorStateArtifact
	State      string
	Message    string
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

	memoryArtifact := acceleratorStateArtifact{
		ID:                  artifactID,
		Content:             encodeAcceleratorContentSnapshot(rendered, time.Now().UTC()),
		Pending:             true,
		AcceleratedChecksum: checksumAcceleratorBytes([]byte(rendered)),
	}
	if len(instructionsPath) > 0 {
		normalizedInstructionsPath, err := ctx.normalizeInstructionsPath(instructionsPath[0])
		if err != nil {
			return err
		}
		memoryArtifact.InstructionsPath = &normalizedInstructionsPath
		memoryArtifact.InstructionsChecksum = ctx.acceleratorInstructionChecksum(normalizedInstructionsPath)
	}

	stateRootPath := ctx.acceleratorStateFilePath()
	stateFilePath := acceleratorArtifactStateFilePath(stateRootPath, scopeName, artifactID)
	storedArtifact, stateExists, err := sanitizeAcceleratorArtifactStateFile(stateFilePath)
	if err != nil {
		return err
	}
	if stateExists {
		storedArtifact.ID = artifactID
	}
	if !stateExists || !acceleratorArtifactGeneratedStateMatches(storedArtifact, memoryArtifact) {
		if err := saveAcceleratorArtifactStateFile(stateFilePath, memoryArtifact); err != nil {
			return err
		}
	}
	ctx.trackAcceleratorState(scopeName, artifactID)

	var previousArtifact *acceleratorStateArtifact
	if stateExists {
		previousArtifact = &storedArtifact
	}
	outputFilePath := filepath.Clean(filepath.Join(ctx.ccodeContext.config.OutputPath, filepath.FromSlash(artifactID)))
	shouldWrite, err := ctx.shouldWriteAcceleratedArtifact(outputFilePath, scopeName, artifactID, rendered, previousArtifact)
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

func (ctx *RunnerContext) trackAcceleratorState(scopeID string, artifactID string) {
	if ctx == nil {
		return
	}
	if ctx.trackedAcceleratorStates == nil {
		ctx.trackedAcceleratorStates = map[string]struct{}{}
	}
	if ctx.trackedAcceleratorScopes == nil {
		ctx.trackedAcceleratorScopes = map[string]struct{}{}
	}
	ctx.trackedAcceleratorStates[acceleratorStateTrackingKey(scopeID, artifactID)] = struct{}{}
	ctx.trackedAcceleratorScopes[scopeID] = struct{}{}
}

func (ctx *RunnerContext) cleanupUntrackedAcceleratorStates() error {
	if ctx == nil || ctx.ccodeContext == nil {
		return fmt.Errorf("runner context is not initialized")
	}

	stateRootPath := ctx.acceleratorStateFilePath()
	if !fileExists(stateRootPath) {
		return nil
	}

	info, err := os.Stat(stateRootPath)
	if err != nil {
		return fmt.Errorf("stat accelerator state %q: %w", stateRootPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("accelerator state %q is not a directory", stateRootPath)
	}

	trackedStates := ctx.trackedAcceleratorStates
	if trackedStates == nil {
		trackedStates = map[string]struct{}{}
	}
	trackedScopes := ctx.trackedAcceleratorScopes
	if trackedScopes == nil || len(trackedScopes) == 0 {
		return nil
	}

	return filepath.WalkDir(stateRootPath, func(itemPath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".accelerated.json") {
			return nil
		}

		relativePath, err := filepath.Rel(stateRootPath, itemPath)
		if err != nil {
			return fmt.Errorf("resolve accelerator state file %q: %w", itemPath, err)
		}

		scopeID, artifactID, err := acceleratorStateIDsFromRelativePath(relativePath)
		if err != nil {
			return err
		}

		if _, ok := trackedScopes[scopeID]; !ok {
			return nil
		}
		if _, ok := trackedStates[acceleratorStateTrackingKey(scopeID, artifactID)]; ok {
			return nil
		}

		if err := removeAcceleratorArtifactStateFile(itemPath); err != nil {
			return err
		}
		return nil
	})
}

func acceleratorStateTrackingKey(scopeID string, artifactID string) string {
	return scopeID + "\x00" + artifactID
}

func (ctx *RunnerContext) normalizeInstructionsPath(instructionsPath string) (string, error) {
	return ctx.ccodeContext.normalizeAcceleratorInstructionPath(instructionsPath)
}

func (ctx *RunnerContext) acceleratorInstructionChecksum(instructionsPath string) string {
	return ctx.ccodeContext.acceleratorInstructionChecksum(instructionsPath)
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

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat accelerator state %q: %w", path, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("accelerator state %q is not a directory", path)
	}

	state := &acceleratorState{
		Version: 1,
		Scopes:  []acceleratorStateScope{},
	}

	err = filepath.WalkDir(path, func(itemPath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".accelerated.json") {
			return nil
		}

		relativePath, err := filepath.Rel(path, itemPath)
		if err != nil {
			return fmt.Errorf("resolve accelerator state file %q: %w", itemPath, err)
		}

		scopeID, artifactID, err := acceleratorStateIDsFromRelativePath(relativePath)
		if err != nil {
			return err
		}

		artifact, err := loadAcceleratorArtifactStateFile(itemPath)
		if err != nil {
			return err
		}
		artifact.ID = artifactID

		scope := state.findOrCreateScope(scopeID)
		scope.Artifacts = append(scope.Artifacts, artifact)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read accelerator state %q: %w", path, err)
	}

	return state, nil
}

func (ctx *Context) inspectAcceleratorState() ([]inspectedAcceleratorArtifact, error) {
	if ctx == nil || ctx.config == nil {
		return nil, fmt.Errorf("context config is required")
	}

	statePath := ctx.acceleratorStateFilePath()
	if !fileExists(statePath) {
		return []inspectedAcceleratorArtifact{}, nil
	}

	info, err := os.Stat(statePath)
	if err != nil {
		return nil, fmt.Errorf("stat accelerator state %q: %w", statePath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("accelerator state %q is not a directory", statePath)
	}

	items := []inspectedAcceleratorArtifact{}
	err = filepath.WalkDir(statePath, func(itemPath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".accelerated.json") {
			return nil
		}

		relativePath, err := filepath.Rel(statePath, itemPath)
		if err != nil {
			return fmt.Errorf("resolve accelerator state file %q: %w", itemPath, err)
		}

		scopeID, artifactID, err := acceleratorStateIDsFromRelativePath(relativePath)
		if err != nil {
			return err
		}

		item, err := ctx.inspectAcceleratorArtifactStateFile(itemPath, scopeID, artifactID)
		if err != nil {
			return err
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read accelerator state %q: %w", statePath, err)
	}

	return items, nil
}

func (ctx *Context) inspectAcceleratorArtifactStateFile(path string, scopeID string, artifactID string) (inspectedAcceleratorArtifact, error) {
	item := inspectedAcceleratorArtifact{
		ScopeID:    scopeID,
		ArtifactID: artifactID,
		State:      acceleratorReportStateCorrupt,
		Message:    acceleratorStateMessage(acceleratorReportStateCorrupt),
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		return item, fmt.Errorf("read accelerator artifact state %q: %w", path, err)
	}

	lines := acceleratorStatePayloadLines(payload)
	if len(lines) == 0 {
		return item, nil
	}

	firstLine := lines[0]
	for _, line := range lines[1:] {
		if line != firstLine {
			item.State = acceleratorReportStateAmbiguous
			item.Message = acceleratorStateMessage(acceleratorReportStateAmbiguous)
			return item, nil
		}
	}

	if len(lines) > 1 || string(payload) != firstLine+"\n" {
		if err := os.WriteFile(path, []byte(firstLine+"\n"), 0644); err != nil {
			return item, fmt.Errorf("clean accelerator artifact state %q: %w", path, err)
		}
	}

	artifact, err := parseAcceleratorArtifactStatePayload([]byte(firstLine))
	if err != nil {
		item.State = acceleratorReportStateCorrupt
		item.Message = acceleratorStateMessage(acceleratorReportStateCorrupt)
		return item, nil
	}
	artifact.ID = artifactID

	return ctx.inspectParsedAcceleratorArtifact(path, scopeID, artifactID, artifact)
}

func (ctx *Context) inspectParsedAcceleratorArtifact(path string, scopeID string, artifactID string, artifact acceleratorStateArtifact) (inspectedAcceleratorArtifact, error) {
	item := inspectedAcceleratorArtifact{
		ScopeID:    scopeID,
		ArtifactID: artifactID,
		Artifact:   &artifact,
		State:      acceleratorArtifactReportState(artifact),
	}

	if artifact.InstructionsPath == nil || isStringBlank(*artifact.InstructionsPath) {
		if !ctx.acceleratorOutputArtifactExists(artifactID) {
			item.State = acceleratorReportStateMissingArtifact
			item.Message = acceleratorStateMessage(acceleratorReportStateMissingArtifact)
		}
		return item, nil
	}

	normalizedInstructionsPath, err := ctx.normalizeAcceleratorInstructionPath(*artifact.InstructionsPath)
	if err != nil {
		item.State = acceleratorReportStateCorrupt
		item.Message = acceleratorStateMessage(acceleratorReportStateCorrupt)
		return item, nil
	}
	artifact.InstructionsPath = &normalizedInstructionsPath

	if !ctx.acceleratorOutputArtifactExists(artifactID) {
		item.Artifact = &artifact
		item.State = acceleratorReportStateMissingArtifact
		item.Message = acceleratorStateMessage(acceleratorReportStateMissingArtifact)
		return item, nil
	}

	instructionsFullPath := filepath.Join(ctx.config.CCodePath, filepath.FromSlash(normalizedInstructionsPath))
	instructionsContent, err := os.ReadFile(instructionsFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			item.Artifact = &artifact
			item.State = acceleratorReportStateMissingInstructions
			item.Message = acceleratorStateMessage(acceleratorReportStateMissingInstructions)
			return item, nil
		}
		return item, fmt.Errorf("read instruction file %q: %w", normalizedInstructionsPath, err)
	}

	instructionsChecksum := checksumAcceleratorBytes(instructionsContent)
	if artifact.InstructionsChecksum != instructionsChecksum {
		artifact.InstructionsChecksum = instructionsChecksum
		artifact.Pending = true
		if path != "" {
			if err := saveAcceleratorArtifactStateFile(path, artifact); err != nil {
				return item, err
			}
		}
	}

	item.Artifact = &artifact
	item.State = acceleratorArtifactReportState(artifact)
	return item, nil
}

func (ctx *Context) acceleratorOutputArtifactExists(artifactID string) bool {
	if ctx == nil || ctx.config == nil || isStringBlank(artifactID) {
		return false
	}
	outputFilePath := filepath.Clean(filepath.Join(ctx.config.OutputPath, filepath.FromSlash(artifactID)))
	return fileExists(outputFilePath)
}

func acceleratorArtifactReportState(artifact acceleratorStateArtifact) string {
	if artifact.Pending {
		return acceleratorReportStatePending
	}
	return acceleratorReportStateAdjusted
}

func acceleratorStateMessage(state string) string {
	switch state {
	case acceleratorReportStateCorrupt:
		return "State file is corrupt. Re-run the accelerator to rebuild it."
	case acceleratorReportStateAmbiguous:
		return "State file is ambiguous. Re-run the accelerator to rebuild it."
	case acceleratorReportStateMissingArtifact:
		return "Artifact file is missing. Restore it or re-run the accelerator."
	case acceleratorReportStateMissingInstructions:
		return "Instruction file is missing. Restore it or re-run the accelerator."
	default:
		return ""
	}
}

func acceleratorReportStateIsActionable(state string) bool {
	return state == acceleratorReportStatePending || acceleratorReportStateIsProblem(state)
}

func acceleratorReportStateMatchesListFilter(state string, includeResolved bool) bool {
	return includeResolved || acceleratorReportStateIsActionable(state)
}

func acceleratorReportStateIsProblem(state string) bool {
	return state == acceleratorReportStateCorrupt ||
		state == acceleratorReportStateAmbiguous ||
		state == acceleratorReportStateMissingArtifact ||
		state == acceleratorReportStateMissingInstructions
}

func acceleratorMetadataFromInspected(item inspectedAcceleratorArtifact) AcceleratorArtifactMetadata {
	metadata := AcceleratorArtifactMetadata{
		ScopeID:    item.ScopeID,
		ArtifactID: item.ArtifactID,
		State:      item.State,
		Message:    item.Message,
	}
	if item.Artifact != nil {
		metadata.InstructionsPath = item.Artifact.InstructionsPath
		metadata.Pending = item.Artifact.Pending
	}
	return metadata
}

func acceleratorStateFromInspected(item inspectedAcceleratorArtifact) *AcceleratorArtifactState {
	state := &AcceleratorArtifactState{
		ScopeID:    item.ScopeID,
		ArtifactID: item.ArtifactID,
		State:      item.State,
		Message:    item.Message,
	}
	if item.Artifact != nil {
		state.Content = item.Artifact.Content
		state.InstructionsPath = item.Artifact.InstructionsPath
		state.Pending = item.Artifact.Pending
	}
	return state
}

func acceleratorInstructionReferenceFromInspected(item inspectedAcceleratorArtifact) AcceleratorInstructionReference {
	reference := AcceleratorInstructionReference{
		ScopeID:    item.ScopeID,
		ArtifactID: item.ArtifactID,
		State:      item.State,
		Message:    item.Message,
	}
	if item.Artifact != nil {
		reference.InstructionsPath = item.Artifact.InstructionsPath
	}
	return reference
}

func saveAcceleratorState(path string, state *acceleratorState) error {
	if state == nil {
		return fmt.Errorf("accelerator state is required")
	}

	state.Version = 1
	if state.Scopes == nil {
		state.Scopes = []acceleratorStateScope{}
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create accelerator state directory for %q: %w", path, err)
	}

	for _, scope := range state.Scopes {
		scopeID, err := normalizeAcceleratorPathPart(scope.ID, "scope id")
		if err != nil {
			return err
		}
		for _, artifact := range scope.Artifacts {
			artifactID, err := normalizeAcceleratorPathPart(artifact.ID, "artifact id")
			if err != nil {
				return err
			}
			stateFilePath := acceleratorArtifactStateFilePath(path, scopeID, artifactID)
			if err := saveAcceleratorArtifactStateFile(stateFilePath, artifact); err != nil {
				return err
			}
		}
	}

	return nil
}

type acceleratorArtifactStateFile struct {
	Pending              bool   `json:"pending"`
	Instructions         string `json:"instructions"`
	AcceleratedChecksum  string `json:"accelerated_checksum"`
	InstructionsChecksum string `json:"instructions_checksum"`
	Code                 string `json:"code"`
}

func loadAcceleratorArtifactStateFile(path string) (acceleratorStateArtifact, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return acceleratorStateArtifact{}, fmt.Errorf("read accelerator artifact state %q: %w", path, err)
	}

	artifact, err := parseAcceleratorArtifactStatePayload(payload)
	if err != nil {
		return acceleratorStateArtifact{}, fmt.Errorf("parse accelerator artifact state %q: %w", path, err)
	}

	return artifact, nil
}

func sanitizeAcceleratorArtifactStateFile(path string) (acceleratorStateArtifact, bool, error) {
	if !fileExists(path) {
		return acceleratorStateArtifact{}, false, nil
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		return acceleratorStateArtifact{}, false, fmt.Errorf("read accelerator artifact state %q: %w", path, err)
	}

	lines := acceleratorStatePayloadLines(payload)
	if len(lines) == 0 {
		if err := removeAcceleratorArtifactStateFile(path); err != nil {
			return acceleratorStateArtifact{}, false, err
		}
		return acceleratorStateArtifact{}, false, nil
	}

	firstLine := lines[0]
	for _, line := range lines[1:] {
		if line != firstLine {
			if err := removeAcceleratorArtifactStateFile(path); err != nil {
				return acceleratorStateArtifact{}, false, err
			}
			return acceleratorStateArtifact{}, false, nil
		}
	}

	artifact, err := parseAcceleratorArtifactStatePayload([]byte(firstLine))
	if err != nil {
		if removeErr := removeAcceleratorArtifactStateFile(path); removeErr != nil {
			return acceleratorStateArtifact{}, false, removeErr
		}
		return acceleratorStateArtifact{}, false, nil
	}

	if len(lines) > 1 || string(payload) != firstLine+"\n" {
		if err := os.WriteFile(path, []byte(firstLine+"\n"), 0644); err != nil {
			return acceleratorStateArtifact{}, false, fmt.Errorf("clean accelerator artifact state %q: %w", path, err)
		}
	}

	return artifact, true, nil
}

func acceleratorStatePayloadLines(payload []byte) []string {
	rawLines := strings.Split(string(payload), "\n")
	lines := []string{}
	for _, line := range rawLines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		lines = append(lines, trimmedLine)
	}
	return lines
}

func parseAcceleratorArtifactStatePayload(payload []byte) (acceleratorStateArtifact, error) {
	fileState := acceleratorArtifactStateFile{}
	if err := json.Unmarshal(payload, &fileState); err != nil {
		return acceleratorStateArtifact{}, err
	}

	var instructionsPath *string
	if !isStringBlank(fileState.Instructions) {
		instructionsPath = &fileState.Instructions
	}

	return acceleratorStateArtifact{
		Content:              fileState.Code,
		InstructionsPath:     instructionsPath,
		Pending:              fileState.Pending,
		AcceleratedChecksum:  fileState.AcceleratedChecksum,
		InstructionsChecksum: fileState.InstructionsChecksum,
	}, nil
}

func removeAcceleratorArtifactStateFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove accelerator artifact state %q: %w", path, err)
	}
	return nil
}

func acceleratorArtifactGeneratedStateMatches(storedArtifact acceleratorStateArtifact, memoryArtifact acceleratorStateArtifact) bool {
	return storedArtifact.AcceleratedChecksum == memoryArtifact.AcceleratedChecksum &&
		acceleratorArtifactInstructionsPath(storedArtifact) == acceleratorArtifactInstructionsPath(memoryArtifact) &&
		storedArtifact.InstructionsChecksum == memoryArtifact.InstructionsChecksum
}

func acceleratorArtifactInstructionsPath(artifact acceleratorStateArtifact) string {
	if artifact.InstructionsPath == nil {
		return ""
	}
	return *artifact.InstructionsPath
}

func saveAcceleratorArtifactStateFile(path string, artifact acceleratorStateArtifact) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create accelerator artifact state directory for %q: %w", path, err)
	}

	instructions := ""
	if artifact.InstructionsPath != nil {
		instructions = *artifact.InstructionsPath
	}

	fileState := acceleratorArtifactStateFile{
		Pending:              artifact.Pending,
		Instructions:         instructions,
		AcceleratedChecksum:  artifact.AcceleratedChecksum,
		InstructionsChecksum: artifact.InstructionsChecksum,
		Code:                 artifact.Content,
	}
	if fileState.AcceleratedChecksum == "" && artifact.Content != "" {
		content, err := decodeAcceleratorContentSnapshot(artifact.Content)
		if err == nil {
			fileState.AcceleratedChecksum = checksumAcceleratorBytes([]byte(content))
		}
	}

	payload, err := json.Marshal(fileState)
	if err != nil {
		return fmt.Errorf("encode accelerator artifact state %q: %w", path, err)
	}
	payload = append(payload, '\n')

	if err := os.WriteFile(path, payload, 0644); err != nil {
		return fmt.Errorf("write accelerator artifact state %q: %w", path, err)
	}

	return nil
}

func acceleratorArtifactStateFilePath(rootPath string, scopeID string, artifactID string) string {
	return filepath.Join(rootPath, filepath.FromSlash(scopeID), filepath.FromSlash(artifactID)+".accelerated.json")
}

func acceleratorStateIDsFromRelativePath(relativePath string) (string, string, error) {
	parts := strings.Split(filepath.ToSlash(relativePath), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid accelerator state file path %q", relativePath)
	}

	scopeID := parts[0]
	artifactPath := strings.Join(parts[1:], "/")
	artifactID := strings.TrimSuffix(artifactPath, ".accelerated.json")
	if artifactID == artifactPath || isStringBlank(artifactID) {
		return "", "", fmt.Errorf("invalid accelerator state file path %q", relativePath)
	}

	return scopeID, artifactID, nil
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

func encodeAcceleratorContentSnapshot(content string, _ time.Time) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}

func decodeAcceleratorContentSnapshot(snapshot string) (string, error) {
	encodedContent := snapshot
	if separatorIndex := strings.Index(snapshot, "Z:"); separatorIndex > 0 {
		encodedContent = snapshot[separatorIndex+2:]
	}

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
	return ctx.acceleratorStateRootPath()
}

func (ctx *Context) acceleratorStateRootPath() string {
	hiddenPath := ctx.config.HiddenPath
	if isStringBlank(hiddenPath) {
		hiddenPath = filepath.Join(ctx.config.CCodePath, DefaultHiddenFolderName)
	} else if !filepath.IsAbs(hiddenPath) {
		hiddenPath = filepath.Join(ctx.config.CCodePath, hiddenPath)
	}

	return filepath.Join(hiddenPath, "accelerators")
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

func (ctx *Context) acceleratorInstructionChecksum(instructionsPath string) string {
	if ctx == nil || ctx.config == nil || isStringBlank(instructionsPath) {
		return ""
	}

	fullPath := filepath.Join(ctx.config.CCodePath, filepath.FromSlash(instructionsPath))
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return ""
	}

	return checksumAcceleratorBytes(content)
}

func checksumAcceleratorBytes(content []byte) string {
	checksum := sha256.Sum256(content)
	return fmt.Sprintf("sha256:%x", checksum)
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

	switch {
	case scopeID == nil && artifactID == nil:
		for scopeIndex := range state.Scopes {
			for artifactIndex := range state.Scopes[scopeIndex].Artifacts {
				state.Scopes[scopeIndex].Artifacts[artifactIndex].Pending = false
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
			scope.Artifacts[artifactIndex].Pending = false
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
		artifact.Pending = false
	}

	return saveAcceleratorState(statePath, state)
}

func (ctx *Context) ListNotAdjustedAccelerators(scopeID *string) ([]AcceleratorArtifactMetadata, error) {
	return ctx.ListAccelerators(scopeID, false)
}

func (ctx *Context) ListAccelerators(scopeID *string, includeResolved bool) ([]AcceleratorArtifactMetadata, error) {
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

	items, err := ctx.inspectAcceleratorState()
	if err != nil {
		return nil, err
	}

	metadataItems := []AcceleratorArtifactMetadata{}
	for _, item := range items {
		if filterScopeID != nil && item.ScopeID != *filterScopeID {
			continue
		}
		if !acceleratorReportStateMatchesListFilter(item.State, includeResolved) {
			continue
		}
		metadataItems = append(metadataItems, acceleratorMetadataFromInspected(item))
	}

	return metadataItems, nil
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

	items, err := ctx.inspectAcceleratorState()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.ScopeID != normalizedScopeID || item.ArtifactID != normalizedArtifactID {
			continue
		}
		return acceleratorStateFromInspected(item), nil
	}

	return nil, fmt.Errorf("accelerator artifact %q not found in scope %q", normalizedArtifactID, normalizedScopeID)
}

func (ctx *Context) ListAcceleratorInstructions() ([]AcceleratorInstructionReference, error) {
	return ctx.ListAcceleratorInstructionsWithResolved(false)
}

func (ctx *Context) ListAcceleratorInstructionsWithResolved(includeResolved bool) ([]AcceleratorInstructionReference, error) {
	if ctx == nil || ctx.config == nil {
		return nil, fmt.Errorf("context config is required")
	}

	items, err := ctx.inspectAcceleratorState()
	if err != nil {
		return nil, err
	}

	references := []AcceleratorInstructionReference{}
	for _, item := range items {
		if !acceleratorReportStateMatchesListFilter(item.State, includeResolved) {
			continue
		}
		if item.Artifact == nil || item.Artifact.InstructionsPath == nil || isStringBlank(*item.Artifact.InstructionsPath) {
			if !acceleratorReportStateIsProblem(item.State) {
				continue
			}
		}
		references = append(references, acceleratorInstructionReferenceFromInspected(item))
	}

	return references, nil
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
	return ctx.ListAccelerators(scopeID, false)
}

func (ctx *RunnerContext) ListAccelerators(scopeID *string, includeResolved bool) ([]AcceleratorArtifactMetadata, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.ListAccelerators(scopeID, includeResolved)
}

func (ctx *RunnerContext) GetAcceleratorState(scopeID string, artifactID string) (*AcceleratorArtifactState, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.GetAcceleratorState(scopeID, artifactID)
}

func (ctx *RunnerContext) ListAcceleratorInstructions() ([]AcceleratorInstructionReference, error) {
	return ctx.ListAcceleratorInstructionsWithResolved(false)
}

func (ctx *RunnerContext) ListAcceleratorInstructionsWithResolved(includeResolved bool) ([]AcceleratorInstructionReference, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.ListAcceleratorInstructionsWithResolved(includeResolved)
}

func (ctx *RunnerContext) GetAcceleratorInstruction(markdownPath string) (string, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return "", fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.GetAcceleratorInstruction(markdownPath)
}
