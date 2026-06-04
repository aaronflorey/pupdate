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

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "path: %s\nexists: %t\nroot_directories: %s\nroot_directories_resolved: %s\nworkspace_globs: %s\nworkspace_globs_resolved: %s\nfolder_blacklist: %s\nfolder_blacklist_resolved: %s\nquiet: %s\nallow_scripts: %s\n",
				path,
				exists,
				displayConfigValues(rawConfig.RootDirectories),
				displayConfigValues(resolvedConfig.RootDirectories),
				displayConfigValues(rawConfig.WorkspaceGlobs),
				displayConfigValues(resolvedConfig.WorkspaceGlobs),
				displayConfigValues(rawConfig.FolderBlacklist),
				displayConfigValues(resolvedConfig.FolderBlacklist),
				displayOptionalBool(rawConfig.Quiet),
				displayOptionalBool(rawConfig.AllowScripts),
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

func displayOptionalBool(value *bool) string {
	if value == nil {
		return "(not set)"
	}

	if *value {
		return "true"
	}

	return "false"
}
