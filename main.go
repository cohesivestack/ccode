package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	ccode "github.com/cohesivestack/ccode/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

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

	accResolveCmd := &cobra.Command{
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

	procCmd := &cobra.Command{
		Use:   "proc",
		Short: "Manage processes",
	}

	procAddCmd := &cobra.Command{
		Use:   "add <proc>",
		Short: "Create a process",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}
			return ccode.NewContext(cfg).AddProcess(args[0])
		},
	}

	accCmd := &cobra.Command{
		Use:   "acc",
		Short: "Inspect and resolve accelerated artifacts",
	}

	procListCmd := &cobra.Command{
		Use:   "list",
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

	accListCmd := &cobra.Command{
		Use:   "list",
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

			return writeMetadata(cmd, items)
		},
	}
	accListCmd.Flags().Bool("include-resolved", false, "Include resolved accelerator artifacts")
	accListCmd.Flags().Bool("json", false, "Output JSON instead of YAML metadata")

	accShowCmd := &cobra.Command{
		Use:   "show <artifactId>",
		Short: "Show accelerated artifact metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			metadata, err := ccode.NewContext(cfg).GetAcceleratorMetadata(args[0])
			if err != nil {
				return err
			}
			return writeMetadata(cmd, metadata)
		},
	}
	accShowCmd.Flags().Bool("json", false, "Output JSON instead of YAML metadata")

	accContentCmd := &cobra.Command{
		Use:   "content <artifactId>",
		Short: "Show the original accelerated content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			content, err := ccode.NewContext(cfg).GetAcceleratorContent(args[0])
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write([]byte(content))
			return err
		},
	}

	accDiffCmd := &cobra.Command{
		Use:   "diff <artifactId>",
		Short: "Show differences between accelerated content and the current output",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			diff, err := ccode.NewContext(cfg).GetAcceleratorDiff(args[0])
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write([]byte(diff))
			return err
		},
	}

	accGuideCmd := &cobra.Command{
		Use:   "guide <artifactId>",
		Short: "Show an accelerated artifact's guide markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			guide, err := ccode.NewContext(cfg).GetAcceleratorArtifactGuide(args[0])
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write([]byte(guide.Markdown))
			return err
		},
	}

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(procCmd)
	rootCmd.AddCommand(accCmd)
	procCmd.AddCommand(procListCmd)
	procCmd.AddCommand(procAddCmd)
	accCmd.AddCommand(accListCmd)
	accCmd.AddCommand(accShowCmd)
	accCmd.AddCommand(accContentCmd)
	accCmd.AddCommand(accDiffCmd)
	accCmd.AddCommand(accResolveCmd)
	accCmd.AddCommand(accGuideCmd)

	return rootCmd
}

func writeJSON(cmd *cobra.Command, payload any) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}

func writeMetadata(cmd *cobra.Command, payload any) error {
	asJSON, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	if asJSON {
		return writeJSON(cmd, payload)
	}
	encoder := yaml.NewEncoder(cmd.OutOrStdout())
	defer encoder.Close()
	encoder.SetIndent(2)
	return encoder.Encode(payload)
}
