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

type acceleratorStateArtifact struct {
	ID                  string  `json:"id"`
	Content             string  `json:"content"`
	GuidePath           *string `json:"guide_path"`
	Pending             bool    `json:"pending"`
	AcceleratedChecksum string  `json:"accelerated_checksum"`
	GuideChecksum       string  `json:"guide_checksum"`
}

const (
	acceleratorReportStatePending         = "pending"
	acceleratorReportStateResolved        = "resolved"
	acceleratorReportStateCorrupt         = "corrupt"
	acceleratorReportStateAmbiguous       = "ambiguous"
	acceleratorReportStateMissingArtifact = "missing_artifact"
	acceleratorReportStateMissingGuide    = "missing_guide"
)

type AcceleratorArtifactMetadata struct {
	ArtifactID string  `json:"artifact_id"`
	GuidePath  *string `json:"guide_path"`
	Pending    bool    `json:"-"`
	State      string  `json:"state"`
	Message    string  `json:"message,omitempty"`
}

type AcceleratorArtifactState struct {
	ArtifactID string  `json:"artifact_id"`
	Content    string  `json:"content"`
	GuidePath  *string `json:"guide_path"`
	Pending    bool    `json:"-"`
	State      string  `json:"state"`
	Message    string  `json:"message,omitempty"`
}

type AcceleratorGuideReference struct {
	ArtifactID string  `json:"artifact_id"`
	GuidePath  *string `json:"guide_path"`
	State      string  `json:"state"`
	Message    string  `json:"message,omitempty"`
}

type inspectedAcceleratorArtifact struct {
	ArtifactID string
	Artifact   *acceleratorStateArtifact
	State      string
	Message    string
}

func (ctx *RunnerContext) Accelerate(id string, templatePath string, data goja.Value, guidePath ...string) error {
	if ctx == nil || ctx.ccodeContext == nil {
		return fmt.Errorf("runner context is not initialized")
	}

	artifactID, err := normalizeAcceleratorPathPart(id, "artifact id")
	if err != nil {
		return err
	}

	templateData, err := gojaValueToTemplateData(ctx.runtime, data)
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
	if len(guidePath) > 0 {
		normalizedGuidePath, err := ctx.normalizeGuidePath(guidePath[0])
		if err != nil {
			return err
		}
		memoryArtifact.GuidePath = &normalizedGuidePath
		memoryArtifact.GuideChecksum = ctx.acceleratorGuideChecksum(normalizedGuidePath)
	}

	stateRootPath := ctx.acceleratorStateFilePath()
	stateFilePath := acceleratorArtifactStateFilePath(stateRootPath, artifactID)
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
	var previousArtifact *acceleratorStateArtifact
	if stateExists {
		previousArtifact = &storedArtifact
	}
	outputFilePath := filepath.Clean(filepath.Join(ctx.ccodeContext.config.OutputPath, filepath.FromSlash(artifactID)))
	shouldWrite, err := ctx.shouldWriteAcceleratedArtifact(outputFilePath, artifactID, rendered, previousArtifact)
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

func (ctx *RunnerContext) shouldWriteAcceleratedArtifact(outputFilePath string, artifactID string, rendered string, artifact *acceleratorStateArtifact) (bool, error) {
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
		return false, fmt.Errorf("decode stored content for artifact %q: %w", artifactID, err)
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

func (ctx *RunnerContext) normalizeGuidePath(guidePath string) (string, error) {
	return ctx.ccodeContext.normalizeAcceleratorGuidePath(guidePath)
}

func (ctx *RunnerContext) acceleratorGuideChecksum(guidePath string) string {
	return ctx.ccodeContext.acceleratorGuideChecksum(guidePath)
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

		artifactID, err := acceleratorStateIDFromRelativePath(relativePath)
		if err != nil {
			return err
		}

		item, err := ctx.inspectAcceleratorArtifactStateFile(itemPath, artifactID)
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

func (ctx *Context) inspectAcceleratorArtifactStateFile(path string, artifactID string) (inspectedAcceleratorArtifact, error) {
	item := inspectedAcceleratorArtifact{
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

	return ctx.inspectParsedAcceleratorArtifact(path, artifactID, artifact)
}

func (ctx *Context) inspectParsedAcceleratorArtifact(path string, artifactID string, artifact acceleratorStateArtifact) (inspectedAcceleratorArtifact, error) {
	item := inspectedAcceleratorArtifact{
		ArtifactID: artifactID,
		Artifact:   &artifact,
		State:      acceleratorArtifactReportState(artifact),
	}

	if artifact.GuidePath == nil || isStringBlank(*artifact.GuidePath) {
		if !ctx.acceleratorOutputArtifactExists(artifactID) {
			item.State = acceleratorReportStateMissingArtifact
			item.Message = acceleratorStateMessage(acceleratorReportStateMissingArtifact)
		}
		return item, nil
	}

	normalizedGuidePath, err := ctx.normalizeAcceleratorGuidePath(*artifact.GuidePath)
	if err != nil {
		item.State = acceleratorReportStateCorrupt
		item.Message = acceleratorStateMessage(acceleratorReportStateCorrupt)
		return item, nil
	}
	artifact.GuidePath = &normalizedGuidePath

	if !ctx.acceleratorOutputArtifactExists(artifactID) {
		item.Artifact = &artifact
		item.State = acceleratorReportStateMissingArtifact
		item.Message = acceleratorStateMessage(acceleratorReportStateMissingArtifact)
		return item, nil
	}

	guideFullPath := filepath.Join(ctx.config.CCodePath, filepath.FromSlash(normalizedGuidePath))
	guideContent, err := os.ReadFile(guideFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			item.Artifact = &artifact
			item.State = acceleratorReportStateMissingGuide
			item.Message = acceleratorStateMessage(acceleratorReportStateMissingGuide)
			return item, nil
		}
		return item, fmt.Errorf("read guide file %q: %w", normalizedGuidePath, err)
	}

	guideChecksum := checksumAcceleratorBytes(guideContent)
	if artifact.GuideChecksum != guideChecksum {
		artifact.GuideChecksum = guideChecksum
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
	return acceleratorReportStateResolved
}

func acceleratorStateMessage(state string) string {
	switch state {
	case acceleratorReportStateCorrupt:
		return "State file is corrupt. Re-run the accelerator to rebuild it."
	case acceleratorReportStateAmbiguous:
		return "State file is ambiguous. Re-run the accelerator to rebuild it."
	case acceleratorReportStateMissingArtifact:
		return "Artifact file is missing. Restore it or re-run the accelerator."
	case acceleratorReportStateMissingGuide:
		return "Guide file is missing. Restore it or re-run the accelerator."
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
		state == acceleratorReportStateMissingGuide
}

func acceleratorMetadataFromInspected(item inspectedAcceleratorArtifact) AcceleratorArtifactMetadata {
	metadata := AcceleratorArtifactMetadata{
		ArtifactID: item.ArtifactID,
		State:      item.State,
		Message:    item.Message,
	}
	if item.Artifact != nil {
		metadata.GuidePath = item.Artifact.GuidePath
		metadata.Pending = item.Artifact.Pending
	}
	return metadata
}

func acceleratorStateFromInspected(item inspectedAcceleratorArtifact) *AcceleratorArtifactState {
	state := &AcceleratorArtifactState{
		ArtifactID: item.ArtifactID,
		State:      item.State,
		Message:    item.Message,
	}
	if item.Artifact != nil {
		state.Content = item.Artifact.Content
		state.GuidePath = item.Artifact.GuidePath
		state.Pending = item.Artifact.Pending
	}
	return state
}

func acceleratorGuideReferenceFromInspected(item inspectedAcceleratorArtifact) AcceleratorGuideReference {
	reference := AcceleratorGuideReference{
		ArtifactID: item.ArtifactID,
		State:      item.State,
		Message:    item.Message,
	}
	if item.Artifact != nil {
		reference.GuidePath = item.Artifact.GuidePath
	}
	return reference
}

type acceleratorArtifactStateFile struct {
	Pending             bool   `json:"pending"`
	Guide               string `json:"guide"`
	AcceleratedChecksum string `json:"accelerated_checksum"`
	GuideChecksum       string `json:"guide_checksum"`
	Code                string `json:"code"`
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

	var guidePath *string
	if !isStringBlank(fileState.Guide) {
		guidePath = &fileState.Guide
	}

	return acceleratorStateArtifact{
		Content:             fileState.Code,
		GuidePath:           guidePath,
		Pending:             fileState.Pending,
		AcceleratedChecksum: fileState.AcceleratedChecksum,
		GuideChecksum:       fileState.GuideChecksum,
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
		acceleratorArtifactGuidePath(storedArtifact) == acceleratorArtifactGuidePath(memoryArtifact) &&
		storedArtifact.GuideChecksum == memoryArtifact.GuideChecksum
}

func acceleratorArtifactGuidePath(artifact acceleratorStateArtifact) string {
	if artifact.GuidePath == nil {
		return ""
	}
	return *artifact.GuidePath
}

func saveAcceleratorArtifactStateFile(path string, artifact acceleratorStateArtifact) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create accelerator artifact state directory for %q: %w", path, err)
	}

	guide := ""
	if artifact.GuidePath != nil {
		guide = *artifact.GuidePath
	}

	fileState := acceleratorArtifactStateFile{
		Pending:             artifact.Pending,
		Guide:               guide,
		AcceleratedChecksum: artifact.AcceleratedChecksum,
		GuideChecksum:       artifact.GuideChecksum,
		Code:                artifact.Content,
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

func acceleratorArtifactStateFilePath(rootPath string, artifactID string) string {
	return filepath.Join(rootPath, filepath.FromSlash(artifactID)+".accelerated.json")
}

func acceleratorStateIDFromRelativePath(relativePath string) (string, error) {
	artifactPath := filepath.ToSlash(relativePath)
	artifactID := strings.TrimSuffix(artifactPath, ".accelerated.json")
	if artifactID == artifactPath || isStringBlank(artifactID) {
		return "", fmt.Errorf("invalid accelerator state file path %q", relativePath)
	}

	return artifactID, nil
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

func (ctx *Context) normalizeAcceleratorGuidePath(guidePath string) (string, error) {
	if isStringBlank(guidePath) {
		return "", fmt.Errorf("guide path is required when provided")
	}

	normalizedPath := filepath.Clean(filepath.FromSlash(strings.TrimSpace(guidePath)))
	if filepath.IsAbs(normalizedPath) {
		relativePath, err := filepath.Rel(ctx.config.CCodePath, normalizedPath)
		if err != nil {
			return "", fmt.Errorf("resolve guide path %q: %w", guidePath, err)
		}
		normalizedPath = filepath.Clean(relativePath)
	}

	if normalizedPath == "." || normalizedPath == ".." || strings.HasPrefix(normalizedPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("guide path %q must be inside ccode path", guidePath)
	}

	return filepath.ToSlash(normalizedPath), nil
}

func (ctx *Context) acceleratorGuideChecksum(guidePath string) string {
	if ctx == nil || ctx.config == nil || isStringBlank(guidePath) {
		return ""
	}

	fullPath := filepath.Join(ctx.config.CCodePath, filepath.FromSlash(guidePath))
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

func (ctx *Context) MarkAcceleratorAsResolved(artifactID *string) error {
	if ctx == nil || ctx.config == nil {
		return fmt.Errorf("context config is required")
	}

	stateRootPath := ctx.acceleratorStateFilePath()
	if artifactID != nil {
		normalizedArtifactID, err := normalizeAcceleratorPathPart(*artifactID, "artifact id")
		if err != nil {
			return err
		}
		stateFilePath := acceleratorArtifactStateFilePath(stateRootPath, normalizedArtifactID)
		if !fileExists(stateFilePath) {
			return fmt.Errorf("accelerator artifact %q not found", normalizedArtifactID)
		}
		return markAcceleratorStateFileAsResolved(stateFilePath)
	}

	if !fileExists(stateRootPath) {
		return nil
	}
	return filepath.WalkDir(stateRootPath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".accelerated.json") {
			return nil
		}
		return markAcceleratorStateFileAsResolved(path)
	})
}

func markAcceleratorStateFileAsResolved(path string) error {
	artifact, err := loadAcceleratorArtifactStateFile(path)
	if err != nil {
		return err
	}
	artifact.Pending = false
	return saveAcceleratorArtifactStateFile(path, artifact)
}

func (ctx *Context) ListNotResolvedAccelerators() ([]AcceleratorArtifactMetadata, error) {
	return ctx.ListAccelerators(false)
}

func (ctx *Context) ListAccelerators(includeResolved bool) ([]AcceleratorArtifactMetadata, error) {
	if ctx == nil || ctx.config == nil {
		return nil, fmt.Errorf("context config is required")
	}

	items, err := ctx.inspectAcceleratorState()
	if err != nil {
		return nil, err
	}

	metadataItems := []AcceleratorArtifactMetadata{}
	for _, item := range items {
		if !acceleratorReportStateMatchesListFilter(item.State, includeResolved) {
			continue
		}
		metadataItems = append(metadataItems, acceleratorMetadataFromInspected(item))
	}

	return metadataItems, nil
}

func (ctx *Context) GetAcceleratorState(artifactID string) (*AcceleratorArtifactState, error) {
	if ctx == nil || ctx.config == nil {
		return nil, fmt.Errorf("context config is required")
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
		if item.ArtifactID != normalizedArtifactID {
			continue
		}
		return acceleratorStateFromInspected(item), nil
	}

	return nil, fmt.Errorf("accelerator artifact %q not found", normalizedArtifactID)
}

func (ctx *Context) ListAcceleratorGuides() ([]AcceleratorGuideReference, error) {
	return ctx.ListAcceleratorGuidesWithResolved(false)
}

func (ctx *Context) ListAcceleratorGuidesWithResolved(includeResolved bool) ([]AcceleratorGuideReference, error) {
	if ctx == nil || ctx.config == nil {
		return nil, fmt.Errorf("context config is required")
	}

	items, err := ctx.inspectAcceleratorState()
	if err != nil {
		return nil, err
	}

	references := []AcceleratorGuideReference{}
	for _, item := range items {
		if !acceleratorReportStateMatchesListFilter(item.State, includeResolved) {
			continue
		}
		if item.Artifact == nil || item.Artifact.GuidePath == nil || isStringBlank(*item.Artifact.GuidePath) {
			if !acceleratorReportStateIsProblem(item.State) {
				continue
			}
		}
		references = append(references, acceleratorGuideReferenceFromInspected(item))
	}

	return references, nil
}

func (ctx *Context) GetAcceleratorGuide(guidePath string) (string, error) {
	if ctx == nil || ctx.config == nil {
		return "", fmt.Errorf("context config is required")
	}

	normalizedPath, err := ctx.normalizeAcceleratorGuidePath(guidePath)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(ctx.config.CCodePath, filepath.FromSlash(normalizedPath))
	if !fileExists(fullPath) {
		return "", fmt.Errorf("guide file not found: %s", normalizedPath)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("read guide file %q: %w", normalizedPath, err)
	}

	return string(content), nil
}

func (ctx *RunnerContext) MarkAcceleratorAsResolved(artifactID *string) error {
	if ctx == nil || ctx.ccodeContext == nil {
		return fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.MarkAcceleratorAsResolved(artifactID)
}

func (ctx *RunnerContext) ListNotResolvedAccelerators() ([]AcceleratorArtifactMetadata, error) {
	return ctx.ListAccelerators(false)
}

func (ctx *RunnerContext) ListAccelerators(includeResolved bool) ([]AcceleratorArtifactMetadata, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.ListAccelerators(includeResolved)
}

func (ctx *RunnerContext) GetAcceleratorState(artifactID string) (*AcceleratorArtifactState, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.GetAcceleratorState(artifactID)
}

func (ctx *RunnerContext) ListAcceleratorGuides() ([]AcceleratorGuideReference, error) {
	return ctx.ListAcceleratorGuidesWithResolved(false)
}

func (ctx *RunnerContext) ListAcceleratorGuidesWithResolved(includeResolved bool) ([]AcceleratorGuideReference, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.ListAcceleratorGuidesWithResolved(includeResolved)
}

func (ctx *RunnerContext) GetAcceleratorGuide(guidePath string) (string, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return "", fmt.Errorf("runner context is not initialized")
	}

	return ctx.ccodeContext.GetAcceleratorGuide(guidePath)
}
