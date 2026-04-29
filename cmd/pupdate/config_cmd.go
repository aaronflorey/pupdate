package main

import (
	"fmt"
	"os"
	"strings"

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

			_, statErr := os.Stat(path)
			exists := statErr == nil
			if statErr != nil && !os.IsNotExist(statErr) {
				return fmt.Errorf("failed to stat %s: %w", path, statErr)
			}

			rawConfig, err := readUserConfig(path)
			if err != nil {
				return err
			}

			resolvedConfig, err := resolveUserConfig(rawConfig)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "path: %s\nexists: %t\nroot_directories: %s\nroot_directories_resolved: %s\n",
				path,
				exists,
				displayConfigValues(rawConfig.RootDirectories),
				displayConfigValues(resolvedConfig.RootDirectories),
			)
			return err
		},
	}

	return cmd
}

func displayConfigValues(values []string) string {
	if len(values) == 0 {
		return "(not set)"
	}

	return strings.Join(values, ", ")
}
