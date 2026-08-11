package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Show resolved user config",
		Long:    "Show the resolved user config path and active config values.",
		Example: `pupdate config`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to resolve working directory: %w", err)
			}
			snapshot, err := dependenciesFor(cmd).ConfigLoader.Load(workingDirectory)
			if err != nil {
				return err
			}
			path := snapshot.User.Path
			projectPath := snapshot.Project.Path

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "path: %s\nexists: %t\nroot_directories: %s\nroot_directories_resolved: %s\nworkspace_globs: %s\nworkspace_globs_resolved: %s\nfolder_blacklist: %s\nfolder_blacklist_resolved: %s\nquiet: %s\nallow_scripts: %s\nproject_path: %s\nproject_exists: %t\nproject_workspace_globs: %s\nproject_workspace_globs_resolved: %s\nproject_folder_blacklist: %s\nproject_folder_blacklist_resolved: %s\neffective_workspace_globs: %s\neffective_folder_blacklist: %s\n",
				safeDisplay(path),
				snapshot.User.Exists,
				displayConfigValues(snapshot.User.Raw.RootDirectories),
				displayConfigValues(snapshot.User.Resolved.RootDirectories),
				displayConfigValues(snapshot.User.Raw.WorkspaceGlobs),
				displayConfigValues(snapshot.User.Resolved.WorkspaceGlobs),
				displayConfigValues(snapshot.User.Raw.FolderBlacklist),
				displayConfigValues(snapshot.User.Resolved.FolderBlacklist),
				displayOptionalBool(snapshot.User.Raw.Quiet),
				displayOptionalBool(snapshot.User.Raw.AllowScripts),
				safeDisplay(projectPath),
				snapshot.Project.Exists,
				displayConfigValues(snapshot.Project.Raw.WorkspaceGlobs),
				displayConfigValues(snapshot.Project.Resolved.WorkspaceGlobs),
				displayConfigValues(snapshot.Project.Raw.FolderBlacklist),
				displayConfigValues(snapshot.Project.Resolved.FolderBlacklist),
				displayConfigValues(snapshot.Effective.WorkspaceGlobs),
				displayConfigValues(snapshot.Effective.FolderBlacklist),
			)
			return err
		},
	}

	return cmd
}

func displayConfigValues(values []string) string {
	return displayStatusValues(values, "(not set)")
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
