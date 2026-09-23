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

type acceleratorGuideAgentResult struct {
	ArtifactID         string  `json:"artifact_id"`
	GuidePath          *string `json:"guide_path"`
	State              string  `json:"state"`
	Message            string  `json:"message,omitempty"`
	GuideMarkdown      string  `json:"guide_markdown"`
	AcceleratedContent string  `json:"accelerated_content"`
	ComposedMarkdown   string  `json:"composed_markdown"`
}

type guideContentAgentResult struct {
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
	if value, ok := os.LookupEnv("CCODE_CCODE_PATH"); ok {
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
	if cmd.Flags().Changed("output-path") {
		value, err := cmd.Flags().GetString("output-path")
		if err != nil {
			return err
		}
		cfg.OutputPath = value
	}
	return nil
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
	rootCmd.PersistentFlags().String("output-path", "", "Root path where generated artifacts are written")

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

	resolveCmd := &cobra.Command{
		Use:   "resolve <artifactId>",
		Short: "Mark an accelerated artifact as resolved",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			context := ccode.NewContext(cfg)
			return context.MarkAcceleratorAsResolved(&args[0])
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List processes and accelerator metadata",
	}

	listProcessesCmd := &cobra.Command{
		Use:   "processes",
		Short: "List runnable processes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			processes, err := ccode.NewContext(cfg).ListProcesses()
			if err != nil {
				return err
			}

			return writeJSON(cmd, processes)
		},
	}

	listAccCmd := &cobra.Command{
		Use:   "acc",
		Short: "List pending accelerated artifacts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			includeResolved, err := cmd.Flags().GetBool("include-resolved")
			if err != nil {
				return err
			}

			items, err := ccode.NewContext(cfg).ListAccelerators(includeResolved)
			if err != nil {
				return err
			}

			return writeJSON(cmd, items)
		},
	}
	listAccCmd.Flags().Bool("include-resolved", false, "Include resolved accelerator artifacts")
	listAccCmd.Flags().Bool("for-agent", false, "Output machine-readable JSON")

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get accelerator details and guides",
	}

	getAcceleratedCmd := &cobra.Command{
		Use:   "accelerated <artifactId>",
		Short: "Get accelerated artifact metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			context := ccode.NewContext(cfg)
			state, err := context.GetAcceleratorState(args[0])
			if err != nil {
				return err
			}

			emitGuide, err := cmd.Flags().GetBool("guide")
			if err != nil {
				return err
			}
			forAgent, err := cmd.Flags().GetBool("for-agent")
			if err != nil {
				return err
			}

			if !emitGuide {
				return writeJSON(cmd, ccode.AcceleratorArtifactMetadata{
					ArtifactID: state.ArtifactID,
					GuidePath:  state.GuidePath,
					Pending:    state.Pending,
					State:      state.State,
					Message:    state.Message,
				})
			}

			acceleratedContent := ""
			if !isStringBlank(state.Content) {
				acceleratedContent, err = ccode.DecodeAcceleratorContentSnapshot(state.Content)
				if err != nil {
					return err
				}
			}

			guideMarkdown := ""
			if state.GuidePath != nil && !isStringBlank(*state.GuidePath) {
				guideMarkdown, err = context.GetAcceleratorGuide(*state.GuidePath)
				if err != nil && state.Message == "" {
					return err
				}
			}

			composedMarkdown := composeAcceleratedGuideMarkdown(guideMarkdown, acceleratedContent, state.ArtifactID, state.Message)
			if forAgent {
				return writeJSON(cmd, acceleratorGuideAgentResult{
					ArtifactID:         state.ArtifactID,
					GuidePath:          state.GuidePath,
					State:              state.State,
					Message:            state.Message,
					GuideMarkdown:      guideMarkdown,
					AcceleratedContent: acceleratedContent,
					ComposedMarkdown:   composedMarkdown,
				})
			}

			_, err = cmd.OutOrStdout().Write([]byte(composedMarkdown))
			return err
		},
	}
	getAcceleratedCmd.Flags().Bool("guide", false, "Include guide markdown with decoded accelerated content")
	getAcceleratedCmd.Flags().Bool("for-agent", false, "Output machine-readable JSON")

	getGuideCmd := &cobra.Command{
		Use:   "guide <path>",
		Short: "Get raw accelerator guide markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			context := ccode.NewContext(cfg)
			markdown, err := context.GetAcceleratorGuide(args[0])
			if err != nil {
				return err
			}

			forAgent, err := cmd.Flags().GetBool("for-agent")
			if err != nil {
				return err
			}
			if forAgent {
				return writeJSON(cmd, guideContentAgentResult{
					Path:     args[0],
					Markdown: markdown,
				})
			}

			_, err = cmd.OutOrStdout().Write([]byte(markdown))
			return err
		},
	}
	getGuideCmd.Flags().Bool("for-agent", false, "Output machine-readable JSON")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(resolveCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	listCmd.AddCommand(listProcessesCmd)
	listCmd.AddCommand(listAccCmd)
	getCmd.AddCommand(getAcceleratedCmd)
	getCmd.AddCommand(getGuideCmd)

	return rootCmd
}

func writeJSON(cmd *cobra.Command, payload any) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}

func composeAcceleratedGuideMarkdown(guideMarkdown string, acceleratedContent string, artifactID string, stateMessage string) string {
	builder := &strings.Builder{}

	if !isStringBlank(guideMarkdown) {
		builder.WriteString(strings.TrimRight(guideMarkdown, "\n"))
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
	if !isStringBlank(stateMessage) {
		builder.WriteString("\n")
		builder.WriteString(strings.TrimSpace(stateMessage))
		builder.WriteString("\n")
	}

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
