package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show resolved user config",
		Long:  "Show the resolved user config path and active config values.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveUserConfigPath()
			if err != nil {
				return err
			}

			rawConfig, err := readUserConfig(path)
			if err != nil {
				return err
			}

			resolvedConfig, err := resolveUserConfig(rawConfig)
			if err != nil {
				return err
			}

			_, statErr := os.Stat(path)
			exists := statErr == nil
			if statErr != nil && !os.IsNotExist(statErr) {
				return fmt.Errorf("failed to stat %s: %w", path, statErr)
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "path: %s\nexists: %t\nroot_directory: %s\nroot_directory_resolved: %s\n",
				path,
				exists,
				displayConfigValue(rawConfig.RootDirectory),
				displayConfigValue(resolvedConfig.RootDirectory),
			)
			return err
		},
	}

	return cmd
}

func displayConfigValue(value string) string {
	if value == "" {
		return "(not set)"
	}

	return value
}
