package main

import (
	"errors"
	"fmt"
	"strings"

	ccode "github.com/cohesivestack/ccode/lib"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	if err := newRootCmd(func(cfg *ccode.Config, process string) error {
		_ = cfg
		_ = process
		return nil
	}, ccode.Init).Execute(); err != nil {
		panic(err)
	}
}

// loadConfigFromViper builds a Config using the given viper instance,
// applies defaults and validations via NewConfig, and returns it.
// It respects `yaml` struct tags when unmarshalling.
func loadConfigFromViper(v *viper.Viper) (*ccode.Config, error) {
	raw := &ccode.Config{}

	if err := v.Unmarshal(raw, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "yaml"
	}); err != nil {
		return nil, fmt.Errorf("unable to decode configuration: %w", err)
	}

	cfg, err := ccode.NewConfig(raw)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// initViper sets up viper with config file (optional), env prefix + replacer.
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

	v.SetEnvPrefix("CCODE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	return v, nil
}

// applyChangedFlagsToViper sets only flags explicitly provided by the user.
func applyChangedFlagsToViper(cmd *cobra.Command, v *viper.Viper) error {
	get := func(name string) (any, error) {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			return nil, fmt.Errorf("flag %q not found", name)
		}
		switch f.Value.Type() {
		case "bool":
			return cmd.Flags().GetBool(name)
		case "string":
			return cmd.Flags().GetString(name)
		case "stringSlice":
			return cmd.Flags().GetStringSlice(name)
		case "int":
			return cmd.Flags().GetInt(name)
		default:
			return f.Value.String(), nil
		}
	}

	cmd.Flags().Visit(func(f *pflag.Flag) {
		if f.Name == "config" {
			return
		}

		val, err := get(f.Name)
		if err != nil {
			return
		}

		key := f.Name
		switch key {
		case "output-path":
			key = "output_path"
		}

		switch valType := val.(type) {
		case string:
			if valType != "" {
				v.Set(key, val)
			}
		case []string:
			if len(valType) > 0 {
				v.Set(key, val)
			}
		case bool:
			v.Set(key, val)
		default:
			v.Set(key, val)
		}
	})

	return nil
}

// newRootCmd creates the Cobra CLI and wires viper + config loading.
func newRootCmd(run func(*ccode.Config, string) error, initProject func(string, string) error) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ccode",
		Short: "Cohesive Code CLI",
		Long: `Cohesive Code is an AI-enabled code generator.

Syntax:
  ccode --config [config-path] --path [path] --output-path [output-path] run [process]
  ccode --config [config-path] init [path]`,
	}

	rootCmd.PersistentFlags().String("config", "", "Path to YAML config file")
	rootCmd.PersistentFlags().String("path", "", "Path where the project structure resides")
	rootCmd.PersistentFlags().String("output-path", "", "Root path where generated artifacts are written")

	runCmd := &cobra.Command{
		Use:   "run [process]",
		Short: "Run a process",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := initViper(cmd)
			if err != nil {
				return err
			}
			if err := applyChangedFlagsToViper(cmd, v); err != nil {
				return err
			}

			cfg, err := loadConfigFromViper(v)
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
