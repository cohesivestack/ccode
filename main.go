package main

import (
	"errors"
	"fmt"
	"os"

	ccode "github.com/cohesivestack/ccode/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
func newRootCmd(run func(*ccode.Config, string) error, initProject func(string, string) error) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ccode",
		Short: "Cohesive Code CLI",
		Long: `Cohesive Code is an AI-enabled code generator.

Syntax:
  ccode --config [config-path] --ccode-path [path] --output-path [output-path] run [process]
  ccode --config [config-path] init [path]`,
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
			if initProject == nil {
				return nil
			}

			return initProject(projectPath, configPath)
		},
	}

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(initCmd)

	return rootCmd
}
