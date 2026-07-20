package main

import (
	"fmt"
	"os"
	"path/filepath"
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

			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to resolve working directory: %w", err)
			}
			projectPath := filepath.Join(workingDirectory, projectConfigFileName)
			loaded, err := loadPreflightConfig(path, projectPath)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "path: %s\nexists: %t\nroot_directories: %s\nroot_directories_resolved: %s\nworkspace_globs: %s\nworkspace_globs_resolved: %s\nfolder_blacklist: %s\nfolder_blacklist_resolved: %s\nquiet: %s\nallow_scripts: %s\nproject_path: %s\nproject_exists: %t\nproject_workspace_globs: %s\nproject_workspace_globs_resolved: %s\nproject_folder_blacklist: %s\nproject_folder_blacklist_resolved: %s\neffective_workspace_globs: %s\neffective_folder_blacklist: %s\n",
				safeDisplay(path),
				loaded.UserExists,
				displayConfigValues(loaded.RawUser.RootDirectories),
				displayConfigValues(loaded.ResolvedUser.RootDirectories),
				displayConfigValues(loaded.RawUser.WorkspaceGlobs),
				displayConfigValues(loaded.ResolvedUser.WorkspaceGlobs),
				displayConfigValues(loaded.RawUser.FolderBlacklist),
				displayConfigValues(loaded.ResolvedUser.FolderBlacklist),
				displayOptionalBool(loaded.RawUser.Quiet),
				displayOptionalBool(loaded.RawUser.AllowScripts),
				safeDisplay(projectPath),
				loaded.ProjectExists,
				displayConfigValues(loaded.RawProject.WorkspaceGlobs),
				displayConfigValues(loaded.ResolvedProject.WorkspaceGlobs),
				displayConfigValues(loaded.RawProject.FolderBlacklist),
				displayConfigValues(loaded.ResolvedProject.FolderBlacklist),
				displayConfigValues(loaded.Effective.WorkspaceGlobs),
				displayConfigValues(loaded.Effective.FolderBlacklist),
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

	displayed := make([]string, len(values))
	for index, value := range values {
		displayed[index] = safeDisplay(value)
	}
	return strings.Join(displayed, ", ")
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
