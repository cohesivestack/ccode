package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ccode "github.com/cohesivestack/ccode/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type acceleratorInstructionsAgentResult struct {
	ScopeID              string  `json:"scope_id"`
	ArtifactID           string  `json:"artifact_id"`
	InstructionsPath     *string `json:"instructions_path"`
	InstructionsMarkdown string  `json:"instructions_markdown"`
	AcceleratedContent   string  `json:"accelerated_content"`
	ComposedMarkdown     string  `json:"composed_markdown"`
}

type instructionContentAgentResult struct {
	Path     string `json:"path"`
	Markdown string `json:"markdown"`
}

func main() {
	if err := newRootCmd(ccode.Run, ccode.Init).Execute(); err != nil {
		panic(err)
	}
}

// initViper sets up viper with config file discovery.
func initViper(cmd *cobra.Command) (*viper.Viper, error) {
	v := viper.New()

	cfgFile, _ := cmd.Flags().GetString("config")
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file %q: %w", cfgFile, err)
		}
	} else {
		v.SetConfigName("ccode")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		if err := v.ReadInConfig(); err != nil {
			var nf viper.ConfigFileNotFoundError
			if !errors.As(err, &nf) {
				return nil, fmt.Errorf("failed to read config: %w", err)
			}
		}
	}

	return v, nil
}

func loadConfig(cmd *cobra.Command) (*ccode.Config, error) {
	v, err := initViper(cmd)
	if err != nil {
		return nil, err
	}

	cfg := &ccode.Config{}
	if configFile := v.ConfigFileUsed(); configFile != "" {
		cfg, err = ccode.LoadConfig(configFile)
		if err != nil {
			return nil, err
		}
	}

	applyEnvOverrides(cfg)
	if err := applyFlagOverrides(cmd, cfg); err != nil {
		return nil, err
	}

	return ccode.NewConfig(cfg)
}

func applyEnvOverrides(cfg *ccode.Config) {
	if value, ok := lookupEnv("CCODE_CCODE_PATH", "CCODE_PATH"); ok {
		cfg.CCodePath = value
	}
	if value, ok := os.LookupEnv("CCODE_OUTPUT_PATH"); ok {
		cfg.OutputPath = value
	}
	if value, ok := os.LookupEnv("CCODE_HIDDEN_PATH"); ok {
		cfg.HiddenPath = value
	}
}

func applyFlagOverrides(cmd *cobra.Command, cfg *ccode.Config) error {
	if cmd.Flags().Changed("ccode-path") {
		value, err := cmd.Flags().GetString("ccode-path")
		if err != nil {
			return err
		}
		cfg.CCodePath = value
	}
	if cmd.Flags().Changed("path") {
		value, err := cmd.Flags().GetString("path")
		if err != nil {
			return err
		}
		cfg.CCodePath = value
	}
	if cmd.Flags().Changed("output-path") {
		value, err := cmd.Flags().GetString("output-path")
		if err != nil {
			return err
		}
		cfg.OutputPath = value
	}
	return nil
}

func lookupEnv(names ...string) (string, bool) {
	for _, name := range names {
		if value, ok := os.LookupEnv(name); ok {
			return value, true
		}
	}
	return "", false
}

// newRootCmd creates the Cobra CLI and wires viper + config loading.
func newRootCmd(run func(*ccode.Config, string) error, initProject func(string, string, string) error) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ccode",
		Short: "Cohesive Code CLI",
		Long: `Cohesive Code is an AI-enabled code generator.

Syntax:
  ccode --config [config-path] --ccode-path [path] --output-path [output-path] run [process]
  ccode --config [config-path] init [path] --version [version]`,
	}

	rootCmd.PersistentFlags().String("config", "", "Path to YAML config file")
	rootCmd.PersistentFlags().String("ccode-path", "", "Path where the project structure resides")
	rootCmd.PersistentFlags().String("path", "", "Deprecated alias for --ccode-path")
	rootCmd.PersistentFlags().String("output-path", "", "Root path where generated artifacts are written")
	_ = rootCmd.PersistentFlags().MarkDeprecated("path", "use --ccode-path instead")

	runCmd := &cobra.Command{
		Use:   "run [process]",
		Short: "Run a process",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}
			if run == nil {
				return nil
			}
			return run(cfg, args[0])
		},
	}

	initCmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a ccode project",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath := ""
			if len(args) == 1 {
				projectPath = args[0]
			}

			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			version, err := cmd.Flags().GetString("version")
			if err != nil {
				return err
			}
			if initProject == nil {
				return nil
			}

			return initProject(projectPath, configPath, version)
		},
	}
	initCmd.Flags().String("version", "", "ccode version to write to ccode.yaml")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List accelerator metadata",
	}

	listAcceleratedCmd := &cobra.Command{
		Use:   "accelerated [scopeId]",
		Short: "List not-adjusted accelerated artifacts",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			context := ccode.NewContext(cfg)
			var scopeID *string
			if len(args) == 1 {
				scopeID = &args[0]
			}

			items, err := context.ListNotAdjustedAccelerators(scopeID)
			if err != nil {
				return err
			}

			return writeJSON(cmd, items)
		},
	}
	listAcceleratedCmd.Flags().Bool("for-agent", false, "Output machine-readable JSON")

	listInstructionsCmd := &cobra.Command{
		Use:   "instructions",
		Short: "List accelerator instruction references",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			context := ccode.NewContext(cfg)
			items, err := context.ListAcceleratorInstructions()
			if err != nil {
				return err
			}

			return writeJSON(cmd, items)
		},
	}
	listInstructionsCmd.Flags().Bool("for-agent", false, "Output machine-readable JSON")

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get accelerator details and instructions",
	}

	getAcceleratedCmd := &cobra.Command{
		Use:   "accelerated <scopeId>:<artifactId>",
		Short: "Get accelerated artifact metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			scopeID, artifactID, err := parseAcceleratorSelector(args[0])
			if err != nil {
				return err
			}

			context := ccode.NewContext(cfg)
			state, err := context.GetAcceleratorState(scopeID, artifactID)
			if err != nil {
				return err
			}

			emitInstructions, err := cmd.Flags().GetBool("instructions")
			if err != nil {
				return err
			}
			forAgent, err := cmd.Flags().GetBool("for-agent")
			if err != nil {
				return err
			}

			if !emitInstructions {
				return writeJSON(cmd, ccode.AcceleratorArtifactMetadata{
					ScopeID:          state.ScopeID,
					ArtifactID:       state.ArtifactID,
					InstructionsPath: state.InstructionsPath,
					AdjustedAt:       state.AdjustedAt,
				})
			}

			acceleratedContent, err := ccode.DecodeAcceleratorContentSnapshot(state.Content)
			if err != nil {
				return err
			}

			instructionsMarkdown := ""
			if state.InstructionsPath != nil && !isStringBlank(*state.InstructionsPath) {
				instructionsMarkdown, err = context.GetAcceleratorInstruction(*state.InstructionsPath)
				if err != nil {
					return err
				}
			}

			composedMarkdown := composeAcceleratedInstructionsMarkdown(instructionsMarkdown, acceleratedContent, state.ArtifactID)
			if forAgent {
				return writeJSON(cmd, acceleratorInstructionsAgentResult{
					ScopeID:              state.ScopeID,
					ArtifactID:           state.ArtifactID,
					InstructionsPath:     state.InstructionsPath,
					InstructionsMarkdown: instructionsMarkdown,
					AcceleratedContent:   acceleratedContent,
					ComposedMarkdown:     composedMarkdown,
				})
			}

			_, err = cmd.OutOrStdout().Write([]byte(composedMarkdown))
			return err
		},
	}
	getAcceleratedCmd.Flags().Bool("instructions", false, "Include instruction markdown with decoded accelerated content")
	getAcceleratedCmd.Flags().Bool("for-agent", false, "Output machine-readable JSON")

	getInstructionCmd := &cobra.Command{
		Use:   "instruction <path>",
		Short: "Get raw accelerator instruction markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			context := ccode.NewContext(cfg)
			markdown, err := context.GetAcceleratorInstruction(args[0])
			if err != nil {
				return err
			}

			forAgent, err := cmd.Flags().GetBool("for-agent")
			if err != nil {
				return err
			}
			if forAgent {
				return writeJSON(cmd, instructionContentAgentResult{
					Path:     args[0],
					Markdown: markdown,
				})
			}

			_, err = cmd.OutOrStdout().Write([]byte(markdown))
			return err
		},
	}
	getInstructionCmd.Flags().Bool("for-agent", false, "Output machine-readable JSON")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	listCmd.AddCommand(listAcceleratedCmd)
	listCmd.AddCommand(listInstructionsCmd)
	getCmd.AddCommand(getAcceleratedCmd)
	getCmd.AddCommand(getInstructionCmd)

	return rootCmd
}

func writeJSON(cmd *cobra.Command, payload any) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}

func parseAcceleratorSelector(value string) (string, string, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return "", "", fmt.Errorf("accelerator selector is required")
	}
	if strings.Count(trimmedValue, ":") != 1 {
		return "", "", fmt.Errorf("invalid accelerator selector %q: expected <scopeId>:<artifactId>", value)
	}

	parts := strings.SplitN(trimmedValue, ":", 2)
	if len(parts) != 2 || isStringBlank(parts[0]) || isStringBlank(parts[1]) {
		return "", "", fmt.Errorf("invalid accelerator selector %q: expected <scopeId>:<artifactId>", value)
	}

	return parts[0], parts[1], nil
}

func composeAcceleratedInstructionsMarkdown(instructionsMarkdown string, acceleratedContent string, artifactID string) string {
	builder := &strings.Builder{}

	if !isStringBlank(instructionsMarkdown) {
		builder.WriteString(strings.TrimRight(instructionsMarkdown, "\n"))
		builder.WriteString("\n")
	}

	builder.WriteString("---\n")
	builder.WriteString("New accelerated content:\n")

	language := inferCodeFenceLanguageFromArtifactID(artifactID)
	if language == "" {
		builder.WriteString("```\n")
	} else {
		builder.WriteString("```")
		builder.WriteString(language)
		builder.WriteString("\n")
	}

	builder.WriteString(acceleratedContent)
	if !strings.HasSuffix(acceleratedContent, "\n") {
		builder.WriteString("\n")
	}
	builder.WriteString("```\n")

	return builder.String()
}

func inferCodeFenceLanguageFromArtifactID(artifactID string) string {
	switch strings.ToLower(filepath.Ext(artifactID)) {
	case ".go":
		return "go"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "tsx"
	case ".js":
		return "javascript"
	case ".jsx":
		return "jsx"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md":
		return "markdown"
	case ".sql":
		return "sql"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".sh":
		return "bash"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".xml":
		return "xml"
	default:
		return ""
	}
}

func isStringBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}
